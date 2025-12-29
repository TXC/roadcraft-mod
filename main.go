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
)

// Note: We use text-based parsing to preserve all file content

func main() {
	pakPath := flag.String("pak", "", "Path to the default_other.pak file")
	outputPath := flag.String("output", "", "Output path for modified pak file (default: overwrites input)")
	targetFile := flag.String("file", "", "Target .cls file within pak to modify (optional, modifies all if not specified)")
	allowedPercentage := flag.Float64("percentage", 0.0, "New allowedpercentage value (default: 0.0)")
	listFiles := flag.Bool("list", false, "List all .cls files in the pak")
	flag.Parse()

	if *pakPath == "" {
		fmt.Println("Usage: roadcraft-mod -pak <path-to-pak> [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  roadcraft-mod -pak default_other.pak -percentage 0")
		fmt.Println("  roadcraft-mod -pak default_other.pak -file ssl/autogen_designer_wizard/trucks/auto_ziks605e_mobile_scalper_res/auto_zikz_605e_mobile_scalper_res.cls -percentage 0")
		os.Exit(1)
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

	if err := modifyPakFile(*pakPath, output, *targetFile, *allowedPercentage); err != nil {
		fmt.Fprintf(os.Stderr, "Error modifying pak file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully modified pak file: %s\n", output)
}

func listClsFiles(pakPath string) error {
	r, err := zip.OpenReader(pakPath)
	if err != nil {
		return fmt.Errorf("failed to open pak file: %w", err)
	}
	defer r.Close()

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
	defer r.Close()

	// Create a temporary file for output
	tempFile, err := os.CreateTemp("", "roadcraft-mod-*.pak")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

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
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to open file %s in pak: %w", f.Name, err)
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				w.Close()
				tempFile.Close()
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
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to create file header: %w", err)
			}

			if _, err := fw.Write(modifiedContent); err != nil {
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
		} else {
			// Copy file as-is
			rc, err := f.Open()
			if err != nil {
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to open file %s in pak: %w", f.Name, err)
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to read file %s: %w", f.Name, err)
			}

			// Preserve original compression method
			fh := &zip.FileHeader{
				Name:   f.Name,
				Method: f.Method,
			}
			fw, err := w.CreateHeader(fh)
			if err != nil {
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to create file header: %w", err)
			}

			if _, err := fw.Write(content); err != nil {
				w.Close()
				tempFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	if err := w.Close(); err != nil {
		tempFile.Close()
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

	// Also try XML format for backwards compatibility (AllowedPercentage)
	if !modified {
		patterns := []string{
			`(<AllowedPercentage[^>]*>)([^<]+)(</AllowedPercentage>)`,
			`(<allowedpercentage[^>]*>)([^<]+)(</allowedpercentage>)`,
			`("allowedpercentage"\s*:\s*)([0-9.]+)`,
			`("AllowedPercentage"\s*:\s*)([0-9.]+)`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(`(?i)` + pattern)
			if re.MatchString(text) {
				// For XML-style tags
				if strings.Contains(pattern, "<") {
					result := re.ReplaceAllString(text, fmt.Sprintf("${1}%.1f${3}", allowedPercentage))
					if result != text {
						modified = true
						text = result
						break
					}
				} else {
					// For JSON-style
					result := re.ReplaceAllString(text, fmt.Sprintf("${1}%.1f", allowedPercentage))
					if result != text {
						modified = true
						text = result
						break
					}
				}
			}
		}
	}

	if !modified {
		return content, false, nil
	}

	return []byte(text), true, nil
}
