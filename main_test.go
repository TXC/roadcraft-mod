package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModifyClsFileText tests the text-based modification function
func TestModifyClsFileText(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		percentage       float64
		expectedModified bool
		expectedContains string
	}{
		{
			name: "Actual CLS format with allowedPercent",
			input: `   prop_truck_mobile_sand_screen   =   {
      allowedPercent   =   0.4
      belt1SpeedCoef   =   0.3
   }`,
			percentage:       0.0,
			expectedModified: true,
			expectedContains: "allowedPercent   =   0.0",
		},
		{
			name: "CLS format with different spacing",
			input: `prop_truck_mobile_sand_screen = {
   allowedPercent = 0.5
}`,
			percentage:       0.0,
			expectedModified: true,
			expectedContains: "allowedPercent = 0.0",
		},
		{
			name: "XML with AllowedPercentage (backwards compatibility)",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<_templates>
  <Item>
    <AllowedPercentage>0.5</AllowedPercentage>
    <OtherProperty>test</OtherProperty>
  </Item>
</_templates>`,
			percentage:       0.0,
			expectedModified: true,
			expectedContains: "<AllowedPercentage>0.0</AllowedPercentage>",
		},
		{
			name: "XML with different percentage value",
			input: `<Item>
    <AllowedPercentage>100.0</AllowedPercentage>
</Item>`,
			percentage:       50.0,
			expectedModified: true,
			expectedContains: "<AllowedPercentage>50.0</AllowedPercentage>",
		},
		{
			name: "File without allowedPercent",
			input: `properties = {
   geom = {
      nameTpl = "test"
   }
}`,
			percentage:       0.0,
			expectedModified: false,
			expectedContains: "nameTpl",
		},
		{
			name: "Case insensitive allowedpercentage (XML)",
			input: `<item>
    <allowedpercentage>0.5</allowedpercentage>
</item>`,
			percentage:       0.0,
			expectedModified: true,
			expectedContains: "<allowedpercentage>0.0</allowedpercentage>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, modified, err := modifyClsFileText([]byte(tt.input), tt.percentage)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if modified != tt.expectedModified {
				t.Errorf("expected modified=%v, got %v", tt.expectedModified, modified)
			}

			resultStr := string(result)
			if !strings.Contains(resultStr, tt.expectedContains) {
				t.Errorf("expected result to contain %q, got:\n%s", tt.expectedContains, resultStr)
			}

			// Verify original content is preserved (except the modified field)
			if tt.expectedModified {
				// Check that other content from input is still present
				if strings.Contains(tt.input, "OtherProperty") && !strings.Contains(resultStr, "OtherProperty") {
					t.Error("expected OtherProperty to be preserved")
				}
				if strings.Contains(tt.input, "belt1SpeedCoef") && !strings.Contains(resultStr, "belt1SpeedCoef") {
					t.Error("expected belt1SpeedCoef to be preserved")
				}
				if strings.Contains(tt.input, "nameTpl") && !strings.Contains(resultStr, "nameTpl") {
					t.Error("expected nameTpl to be preserved")
				}
			}
		})
	}
}

// TestModifyClsFile tests the main modification function
func TestModifyClsFile(t *testing.T) {
	input := `properties = {
   prop_truck_mobile_sand_screen = {
      allowedPercent = 0.5
      belt1SpeedCoef = 0.3
   }
}`

	result, modified, err := modifyClsFile([]byte(input), 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !modified {
		t.Error("expected file to be modified")
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "allowedPercent = 0.0") {
		t.Errorf("expected allowedPercent to be 0.0, got:\n%s", resultStr)
	}

	// Verify belt1SpeedCoef field is preserved
	if !strings.Contains(resultStr, "belt1SpeedCoef") {
		t.Error("expected belt1SpeedCoef field to be preserved")
	}
}

// TestCreateAndModifyPakFile tests the full PAK file workflow
func TestCreateAndModifyPakFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create test .cls file content
	clsContent := `properties = {
   prop_truck_mobile_sand_screen = {
      allowedPercent = 0.5
      belt1SpeedCoef = 0.3
   }
}`

	// Create test PAK file
	inputPak := filepath.Join(tempDir, "test_input.pak")
	if err := createTestPak(inputPak, map[string]string{
		"vehicles/test_vehicle.cls": clsContent,
		"other/data.txt":            "some data",
	}); err != nil {
		t.Fatalf("failed to create test pak: %v", err)
	}

	// Test listing files
	t.Run("ListClsFiles", func(t *testing.T) {
		err := listClsFiles(inputPak)
		if err != nil {
			t.Errorf("listClsFiles failed: %v", err)
		}
	})

	// Test modifying the PAK file
	t.Run("ModifyPakFile", func(t *testing.T) {
		outputPak := filepath.Join(tempDir, "test_output.pak")
		err := modifyPakFile(inputPak, outputPak, "", 0.0)
		if err != nil {
			t.Fatalf("modifyPakFile failed: %v", err)
		}

		// Verify the output file exists
		if _, err := os.Stat(outputPak); os.IsNotExist(err) {
			t.Fatal("output pak file was not created")
		}

		// Extract and verify the modified content
		r, err := zip.OpenReader(outputPak)
		if err != nil {
			t.Fatalf("failed to open output pak: %v", err)
		}
		defer r.Close()

		var foundCls bool
		for _, f := range r.File {
			if f.Name == "vehicles/test_vehicle.cls" {
				foundCls = true
				rc, err := f.Open()
				if err != nil {
					t.Fatalf("failed to open cls file: %v", err)
				}
				content, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					t.Fatalf("failed to read cls file: %v", err)
				}

				contentStr := string(content)
				if !strings.Contains(contentStr, "allowedPercent = 0.0") {
					t.Errorf("expected allowedPercent to be 0.0, got:\n%s", contentStr)
				}
				if !strings.Contains(contentStr, "belt1SpeedCoef") {
					t.Error("expected belt1SpeedCoef field to be preserved")
				}

				// Verify Store compression is used
				if f.Method != zip.Store {
					t.Errorf("expected Store compression, got method %d", f.Method)
				}
			}
		}

		if !foundCls {
			t.Error("modified .cls file not found in output pak")
		}
	})

	// Test targeting specific file
	t.Run("ModifySpecificFile", func(t *testing.T) {
		// Create PAK with multiple .cls files
		multiPak := filepath.Join(tempDir, "test_multi.pak")
		if err := createTestPak(multiPak, map[string]string{
			"vehicles/vehicle1.cls": `properties = { prop_truck = { allowedPercent = 1.0 } }`,
			"vehicles/vehicle2.cls": `properties = { prop_truck = { allowedPercent = 2.0 } }`,
		}); err != nil {
			t.Fatalf("failed to create multi pak: %v", err)
		}

		outputPak := filepath.Join(tempDir, "test_multi_output.pak")
		err := modifyPakFile(multiPak, outputPak, "vehicles/vehicle1.cls", 10.0)
		if err != nil {
			t.Fatalf("modifyPakFile failed: %v", err)
		}

		// Verify only vehicle1.cls was modified
		r, err := zip.OpenReader(outputPak)
		if err != nil {
			t.Fatalf("failed to open output pak: %v", err)
		}
		defer r.Close()

		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			contentStr := string(content)
			if f.Name == "vehicles/vehicle1.cls" {
				if !strings.Contains(contentStr, "allowedPercent = 10.0") {
					t.Errorf("vehicle1.cls: expected 10.0, got:\n%s", contentStr)
				}
			} else if f.Name == "vehicles/vehicle2.cls" {
				if !strings.Contains(contentStr, "allowedPercent = 2.0") {
					t.Errorf("vehicle2.cls: expected 2.0 (unchanged), got:\n%s", contentStr)
				}
			}
		}
	})
}

// TestPreserveNonClsFiles verifies that non-.cls files are preserved unchanged
func TestPreserveNonClsFiles(t *testing.T) {
	tempDir := t.TempDir()

	inputPak := filepath.Join(tempDir, "test_preserve.pak")
	textContent := "This is a text file that should not be modified"
	if err := createTestPak(inputPak, map[string]string{
		"vehicles/test.cls": `properties = { prop_truck = { allowedPercent = 1.0 } }`,
		"data/readme.txt":   textContent,
		"data/config.xml":   "<config><value>123</value></config>",
	}); err != nil {
		t.Fatalf("failed to create test pak: %v", err)
	}

	outputPak := filepath.Join(tempDir, "test_preserve_output.pak")
	err := modifyPakFile(inputPak, outputPak, "", 0.0)
	if err != nil {
		t.Fatalf("modifyPakFile failed: %v", err)
	}

	// Verify non-.cls files are unchanged
	r, err := zip.OpenReader(outputPak)
	if err != nil {
		t.Fatalf("failed to open output pak: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "data/readme.txt" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			if string(content) != textContent {
				t.Errorf("readme.txt was modified, expected %q, got %q", textContent, string(content))
			}
		}
	}
}

// createTestPak creates a test PAK file with the given files
func createTestPak(path string, files map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range files {
		// Use Store compression (no compression)
		fh := &zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		}
		fw, err := w.CreateHeader(fh)
		if err != nil {
			return err
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}
