package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TXC/roadcraft-mod/internal/parser"
	"github.com/TXC/roadcraft-mod/internal/patch"
)

// Note: We use text-based parsing to preserve all file content

func main() {
	pakPath := flag.String("pak", "", "Path to the default_other.pak file")
	outputPath := flag.String("output", "", "Output path for modified pak file (default: overwrites input)")
	targetFile := flag.String("file", "", "Target .cls file within pak to modify (optional, modifies all if not specified)")
	allowedPercentage := flag.Float64("percentage", 0.0, "New allowedpercentage value (default: 0.0)")
	listFiles := flag.Bool("list", false, "List all .cls files in the pak")
	listPresets := flag.Bool("list-presets", false, "List all available preset patches")
	usePreset := flag.String("preset", "", "Apply a preset patch by name")
	usePatchMode := flag.Bool("patch", false, "Use patch mode (applies defined patches instead of simple percentage change)")
	flag.Parse()

	if *pakPath == "" {
		fmt.Println("Usage: roadcraft-mod -pak <path-to-pak> [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Legacy mode - modify allowedPercent")
		fmt.Println("  roadcraft-mod -pak default_other.pak -percentage 0")
		fmt.Println()
		fmt.Println("  # Patch mode - apply predefined patches")
		fmt.Println("  roadcraft-mod -pak default_other.pak -patch")
		fmt.Println()
		fmt.Println("  # Use specific preset")
		fmt.Println("  roadcraft-mod -pak default_other.pak -preset \"Zikz 605E Mobile Scalper\"")
		os.Exit(1)
	}

	if *listPresets {
		fmt.Println("Available preset patches:")
		for _, name := range patch.ListPresets() {
			fmt.Printf("  - %s\n", name)
		}
		return
	}

	if *listFiles {
		if err := listClsFiles(*pakPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
			os.Exit(1)
		}
		return
	}

	output := *outputPath
	if output == "" {
		output = *pakPath
	}

	// Determine which mode to use
	var patches []patch.VehiclePatch
	if *usePatchMode || *usePreset != "" {
		if *usePreset != "" {
			// Use specific preset
			preset := patch.GetPatchByName(*usePreset)
			if preset == nil {
				fmt.Fprintf(os.Stderr, "Error: preset '%s' not found\n", *usePreset)
				fmt.Println("\nAvailable presets:")
				for _, name := range patch.ListPresets() {
					fmt.Printf("  - %s\n", name)
				}
				os.Exit(1)
			}
			patches = []patch.VehiclePatch{*preset}
		} else {
			// Use all presets
			patches = patch.GetPresetPatches()
		}
		if err := modifyPakFileWithPatches(*pakPath, output, *targetFile, patches); err != nil {
			fmt.Fprintf(os.Stderr, "Error modifying pak file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Legacy mode - simple percentage change
		if err := modifyPakFile(*pakPath, output, *targetFile, *allowedPercentage); err != nil {
			fmt.Fprintf(os.Stderr, "Error modifying pak file: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Successfully modified pak file: %s\n", output)
}

func listClsFiles(pakPath string) error {
	r, err := zip.OpenReader(pakPath)
	if err != nil {
		return fmt.Errorf("failed to open pak file: %w", err)
	}
	defer func() {
		err := r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pak file: %v\n", err)
		}
	}()

	fmt.Println("Found .cls files:")
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".cls") {
			fmt.Printf("  %s\n", f.Name)
		}
	}
	return nil
}

func modifyPakFile(inputPath, outputPath, targetFile string, allowedPercentage float64) error {
	// Read the input pak file
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open pak file: %w", err)
	}
	defer func() {
		err := r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pak file: %v\n", err)
		}
	}()

	// Create a temporary file for output
	tempFile, err := os.CreateTemp("", "roadcraft-mod-*.pak")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		err := os.Remove(tempPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file: %v\n", err)
		}
	}()

	// Create a new zip writer
	w := zip.NewWriter(tempFile)

	modifiedCount := 0

	// Process each file in the pak
	for _, f := range r.File {
		shouldModify := strings.HasSuffix(strings.ToLower(f.Name), ".cls")
		if targetFile != "" {
			// Clean paths for comparison
			cleanTarget := filepath.ToSlash(filepath.Clean(targetFile))
			cleanName := filepath.ToSlash(filepath.Clean(f.Name))
			shouldModify = shouldModify && (cleanName == cleanTarget || strings.HasSuffix(cleanName, cleanTarget))
		}

		if shouldModify {
			// Read the file content
			rc, err := f.Open()
			if err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				return fmt.Errorf("failed to open file %s in pak: %w", f.Name, err)
			}

			content, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				return fmt.Errorf("failed to read file %s: %w", f.Name, err)
			}

			// Try to modify the content
			modifiedContent, modified, err := modifyClsFile(content, allowedPercentage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", f.Name, err)
				// Write original content if parsing fails
				modifiedContent = content
			} else if modified {
				modifiedCount++
				fmt.Printf("Modified: %s\n", f.Name)
			}

			// Write to new pak with Store compression (no compression)
			fh := &zip.FileHeader{
				Name:   f.Name,
				Method: zip.Store,
			}
			fw, err := w.CreateHeader(fh)
			if err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				return fmt.Errorf("failed to create file header: %w", err)
			}
			if _, err := fw.Write(modifiedContent); err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	if err := w.Close(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Replace the output file
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to move temp file to output: %w", err)
	}

	fmt.Printf("Modified %d file(s)\n", modifiedCount)
	return nil
}

func modifyClsFile(content []byte, allowedPercentage float64) ([]byte, bool, error) {
	// Use text-based replacement to preserve all file content
	return modifyClsFileText(content, allowedPercentage)
}

// modifyPakFileWithPatches applies patches to CLS files in the pak
func modifyPakFileWithPatches(inputPath, outputPath, targetFile string, vehiclePatches []patch.VehiclePatch) error {
	// Read the input pak file
	r, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open pak file: %w", err)
	}
	defer func() {
		err := r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pak file: %v\n", err)
		}
	}()

	// Create a temporary file for output
	tempFile, err := os.CreateTemp("", "roadcraft-mod-*.pak")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		err := os.Remove(tempPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file: %v\n", err)
		}
	}()

	// Create a new zip writer
	w := zip.NewWriter(tempFile)

	modifiedCount := 0

	// Process each file in the pak
	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".cls") {
			// Copy non-CLS files as-is
			if err := copyZipFile(w, f); err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				return err
			}
			continue
		}

		// Check if this file matches any patch pattern
		shouldModify := false
		var applicablePatches []patch.Patch

		if targetFile != "" {
			// If target file specified, only modify that file
			cleanTarget := filepath.ToSlash(filepath.Clean(targetFile))
			cleanName := filepath.ToSlash(filepath.Clean(f.Name))
			shouldModify = cleanName == cleanTarget || strings.HasSuffix(cleanName, cleanTarget)
			if shouldModify {
				// Collect all patches
				for _, vp := range vehiclePatches {
					applicablePatches = append(applicablePatches, vp.Patches...)
				}
			}
		} else {
			// Check each vehicle patch pattern
			for _, vp := range vehiclePatches {
				matched, err := filepath.Match(vp.FilePattern, filepath.Base(f.Name))
				if err == nil && matched {
					shouldModify = true
					applicablePatches = append(applicablePatches, vp.Patches...)
				}
			}
		}

		// Read the file content
		rc, err := f.Open()
		if err != nil {
			_ = w.Close()
			_ = tempFile.Close()
			return fmt.Errorf("failed to open file %s in pak: %w", f.Name, err)
		}

		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			_ = w.Close()
			_ = tempFile.Close()
			return fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}

		if shouldModify && len(applicablePatches) > 0 {
			// Parse, apply patches, and serialize
			clsMap, err := parser.ParseClsFile(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", f.Name, err)
				// Write original content if parsing fails
				if err := writeZipFile(w, f, content); err != nil {
					_ = w.Close()
					_ = tempFile.Close()
					return err
				}
				continue
			}

			// Apply all patches
			modified := false
			for _, p := range applicablePatches {
				if err := p.Apply(clsMap); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to apply patch '%s' to %s: %v\n", p.Name, f.Name, err)
				} else {
					modified = true
				}
			}

			if modified {
				// Serialize back to CLS format
				modifiedContent, err := parser.SerializeClsFile(clsMap)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to serialize %s: %v\n", f.Name, err)
					// Write original content if serialization fails
					if err := writeZipFile(w, f, content); err != nil {
						_ = w.Close()
						_ = tempFile.Close()
						return err
					}
					continue
				}

				modifiedCount++
				fmt.Printf("Modified: %s\n", f.Name)
				if err := writeZipFile(w, f, modifiedContent); err != nil {
					_ = w.Close()
					_ = tempFile.Close()
					return err
				}
			} else {
				// No patches applied, write original
				if err := writeZipFile(w, f, content); err != nil {
					_ = w.Close()
					_ = tempFile.Close()
					return err
				}
			}
		} else {
			// Write original content
			if err := writeZipFile(w, f, content); err != nil {
				_ = w.Close()
				_ = tempFile.Close()
				return err
			}
		}
	}

	if err := w.Close(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Replace the output file
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to move temp file to output: %w", err)
	}

	fmt.Printf("Modified %d file(s)\n", modifiedCount)
	return nil
}

// copyZipFile copies a file from source to destination in zip
func copyZipFile(w *zip.Writer, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", f.Name, err)
	}
	defer func() {
		err := rc.Close()
		if err != nil {
			panic(err)
		}
	}()

	content, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", f.Name, err)
	}

	fh := &zip.FileHeader{
		Name:   f.Name,
		Method: f.Method,
	}
	fw, err := w.CreateHeader(fh)
	if err != nil {
		return fmt.Errorf("failed to create file header: %w", err)
	}

	if _, err := fw.Write(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// writeZipFile writes content to a zip file
func writeZipFile(w *zip.Writer, f *zip.File, content []byte) error {
	fh := &zip.FileHeader{
		Name:   f.Name,
		Method: zip.Store,
	}
	fw, err := w.CreateHeader(fh)
	if err != nil {
		return fmt.Errorf("failed to create file header: %w", err)
	}

	if _, err := fw.Write(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func modifyClsFileText(content []byte, allowedPercentage float64) ([]byte, bool, error) {
	text := string(content)
	modified := false

	// Pattern for the actual CLS format: allowedPercent = VALUE
	// This matches the format: allowedPercent   =   0.4
	// Pattern supports integers and decimals like 0.4, .5, 1.0, etc.
	re := regexp.MustCompile(`(?m)(allowedPercent\s*=\s*)([0-9]*\.?[0-9]+)`)

	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		// Replace each occurrence
		result := re.ReplaceAllString(text, fmt.Sprintf("${1}%.1f", allowedPercentage))
		if result != text {
			modified = true
			text = result
		}
	}

	if !modified {
		return content, false, nil
	}

	return []byte(text), true, nil
}
