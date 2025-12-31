package parser

import (
	"testing"
)

func TestParseClsFile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name: "simple key-value pairs",
			input: `
				key1 = value1
				key2 = 123
				key3 = true
			`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["key1"] != "value1" {
					t.Errorf("key1 = %v, want value1", result["key1"])
				}
				if result["key2"] != "123" {
					t.Errorf("key2 = %v, want 123", result["key2"])
				}
				if result["key3"] != "true" {
					t.Errorf("key3 = %v, want true", result["key3"])
				}
			},
		},
		{
			name: "nested object",
			input: `
				properties = {
					name = "test"
					value = 42
				}
			`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				props, ok := result["properties"].(map[string]any)
				if !ok {
					t.Fatal("properties should be a map")
				}
				if props["name"] != "\"test\"" {
					t.Errorf("properties.name = %v, want \"test\"", props["name"])
				}
				if props["value"] != "42" {
					t.Errorf("properties.value = %v, want 42", props["value"])
				}
			},
		},
		{
			name: "array with simple values",
			input: `
				items = [
					"item1",
					"item2",
					"item3"
				]
			`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				items, ok := result["items"].([]any)
				if !ok {
					t.Fatal("items should be a slice")
				}
				if len(items) != 3 {
					t.Errorf("items length = %v, want 3", len(items))
				}
			},
		},
		{
			name: "array with objects",
			input: `
				allowedMaterials = [
					{
						material = "sand"
					},
					{
						material = "mud_sand"
					}
				]
			`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				materials, ok := result["allowedMaterials"].([]any)
				if !ok {
					t.Fatal("allowedMaterials should be a slice")
				}
				if len(materials) != 2 {
					t.Errorf("allowedMaterials length = %v, want 2", len(materials))
				}
				first, ok := materials[0].(map[string]any)
				if !ok {
					t.Fatal("first material should be a map")
				}
				if first["material"] != "\"sand\"" {
					t.Errorf("first material = %v, want \"sand\"", first["material"])
				}
			},
		},
		{
			name: "real-world Zikz example",
			input: `
properties = {
	geom = {
		nameTpl = "zikz_605e_mobile_scalper_res"
	}
	prop_truck_mobile_sand_screen = {
		chainName = "Ramp"
		allowedMaterials = [
			{
				material = "sand"
			},
			{
				material = "mud_sand"
			}
		]
		locator = "SandReciever"
		sandOnBelt = True
		materialCheckRadius = 6
		allowedPercent = 0.4
		belt1SpeedCoef = 0.3
		belt2SpeedCoef = 0.8
		sandDistance = 150
	}
	prop_usable = {
		smartsEntryPoints = {
			SandStorage = {
				checkers = {
					UsableCheckerDistance = {
						distance = 150
					}
				}
				focusDistance = 150
			}
		}
	}
}
			`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				props, ok := result["properties"].(map[string]any)
				if !ok {
					t.Fatal("properties should be a map")
				}

				// Check nested geom
				geom, ok := props["geom"].(map[string]any)
				if !ok {
					t.Fatal("geom should be a map")
				}
				if geom["nameTpl"] != "\"zikz_605e_mobile_scalper_res\"" {
					t.Errorf("geom.nameTpl = %v", geom["nameTpl"])
				}

				// Check mobile sand screen
				sandScreen, ok := props["prop_truck_mobile_sand_screen"].(map[string]any)
				if !ok {
					t.Fatal("prop_truck_mobile_sand_screen should be a map")
				}
				if sandScreen["allowedPercent"] != "0.4" {
					t.Errorf("allowedPercent = %v, want 0.4", sandScreen["allowedPercent"])
				}
				if sandScreen["sandDistance"] != "150" {
					t.Errorf("sandDistance = %v, want 150", sandScreen["sandDistance"])
				}

				// Check array
				materials, ok := sandScreen["allowedMaterials"].([]any)
				if !ok {
					t.Fatal("allowedMaterials should be a slice")
				}
				if len(materials) != 2 {
					t.Errorf("allowedMaterials length = %v, want 2", len(materials))
				}

				// Check deeply nested prop_usable
				propUsable, ok := props["prop_usable"].(map[string]any)
				if !ok {
					t.Fatal("prop_usable should be a map")
				}
				smartsEntryPoints, ok := propUsable["smartsEntryPoints"].(map[string]any)
				if !ok {
					t.Fatal("smartsEntryPoints should be a map")
				}
				sandStorage, ok := smartsEntryPoints["SandStorage"].(map[string]any)
				if !ok {
					t.Fatal("SandStorage should be a map")
				}
				if sandStorage["focusDistance"] != "150" {
					t.Errorf("focusDistance = %v, want 150", sandStorage["focusDistance"])
				}
			},
		},
		{
			name:    "unmatched brace",
			input:   `properties = { name = "test"`,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "unmatched bracket",
			input:   `items = [ "item1", "item2"`,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseClsFile([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestIsKeyChar(t *testing.T) {
	tests := []struct {
		char byte
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{'-', false},
		{'.', false},
		{' ', false},
		{'=', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			if got := isKeyChar(tt.char); got != tt.want {
				t.Errorf("isKeyChar(%c) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestFindMatchingBrace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "simple match",
			input: "test}",
			want:  4,
		},
		{
			name:  "nested braces",
			input: "outer { inner } }",
			want:  16,
		},
		{
			name:  "no match",
			input: "test",
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findMatchingBrace(tt.input); got != tt.want {
				t.Errorf("findMatchingBrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMatchingBracket(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "simple match",
			input: "test]",
			want:  4,
		},
		{
			name:  "nested brackets",
			input: "outer [ inner ] ]",
			want:  16,
		},
		{
			name:  "no match",
			input: "test",
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findMatchingBracket(tt.input); got != tt.want {
				t.Errorf("findMatchingBracket() = %v, want %v", got, tt.want)
			}
		})
	}
}
