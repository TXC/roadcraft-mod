package patch

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "simple path",
			path: "key1.key2.key3",
			want: []string{"key1", "key2", "key3"},
		},
		{
			name: "single key",
			path: "key",
			want: []string{"key"},
		},
		{
			name: "empty path",
			path: "",
			want: nil,
		},
		{
			name: "array index",
			path: "items[0]",
			want: []string{"items", "[0]"},
		},
		{
			name: "nested with array",
			path: "gears[3].speed",
			want: []string{"gears", "[3]", "speed"},
		},
		{
			name: "multiple arrays",
			path: "data[0].items[5].value",
			want: []string{"data", "[0]", "items", "[5]", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePath(tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("parsePath() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parsePath()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestPatchApplySet(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]any
		patch   Patch
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name: "set simple value",
			initial: map[string]any{
				"key1": "value1",
			},
			patch: Patch{
				Path:      "key1",
				Value:     "newvalue",
				Operation: OpSet,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["key1"] != "newvalue" {
					t.Errorf("key1 = %v, want newvalue", result["key1"])
				}
			},
		},
		{
			name: "set nested value",
			initial: map[string]any{
				"properties": map[string]any{
					"prop1": map[string]any{
						"value": "old",
					},
				},
			},
			patch: Patch{
				Path:      "properties.prop1.value",
				Value:     "new",
				Operation: OpSet,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				props := result["properties"].(map[string]any)
				prop1 := props["prop1"].(map[string]any)
				if prop1["value"] != "new" {
					t.Errorf("properties.prop1.value = %v, want new", prop1["value"])
				}
			},
		},
		{
			name: "create missing intermediate maps",
			initial: map[string]any{
				"key1": "value1",
			},
			patch: Patch{
				Path:      "new.nested.value",
				Value:     "test",
				Operation: OpSet,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				new := result["new"].(map[string]any)
				nested := new["nested"].(map[string]any)
				if nested["value"] != "test" {
					t.Errorf("new.nested.value = %v, want test", nested["value"])
				}
			},
		},
		{
			name: "set array element",
			initial: map[string]any{
				"items": []any{"old1", "old2", "old3"},
			},
			patch: Patch{
				Path:      "items[1]",
				Value:     "new2",
				Operation: OpSet,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				items := result["items"].([]any)
				if items[1] != "new2" {
					t.Errorf("items[1] = %v, want new2", items[1])
				}
				// Verify other items unchanged
				if items[0] != "old1" {
					t.Errorf("items[0] = %v, want old1", items[0])
				}
			},
		},
		{
			name: "set nested array element property",
			initial: map[string]any{
				"gears": []any{
					map[string]any{"speed": "10", "ratio": "1.0"},
					map[string]any{"speed": "20", "ratio": "2.0"},
					map[string]any{"speed": "30", "ratio": "3.0"},
				},
			},
			patch: Patch{
				Path:      "gears[1].speed",
				Value:     "25",
				Operation: OpSet,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				gears := result["gears"].([]any)
				gear1 := gears[1].(map[string]any)
				if gear1["speed"] != "25" {
					t.Errorf("gears[1].speed = %v, want 25", gear1["speed"])
				}
				// Verify ratio unchanged
				if gear1["ratio"] != "2.0" {
					t.Errorf("gears[1].ratio = %v, want 2.0", gear1["ratio"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Apply(tt.initial)
			if (err != nil) != tt.wantErr {
				t.Errorf("Patch.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, tt.initial)
			}
		})
	}
}

func TestPatchApplyDelete(t *testing.T) {
	initial := map[string]any{
		"key1": "value1",
		"key2": map[string]any{
			"nested": "value",
		},
	}

	patch := Patch{
		Path:      "key2.nested",
		Operation: OpDelete,
	}

	err := patch.Apply(initial)
	if err != nil {
		t.Errorf("Patch.Apply() error = %v", err)
		return
	}

	key2 := initial["key2"].(map[string]any)
	if _, exists := key2["nested"]; exists {
		t.Errorf("key2.nested should be deleted")
	}
}

func TestPatchApplyAdd(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]any
		patch   Patch
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name: "add to array",
			initial: map[string]any{
				"items": []any{"item1", "item2"},
			},
			patch: Patch{
				Path:      "items",
				Value:     "item3",
				Operation: OpAdd,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				items := result["items"].([]any)
				if len(items) != 3 {
					t.Errorf("expected 3 items, got %d", len(items))
				}
				if items[2] != "item3" {
					t.Errorf("expected last item to be 'item3', got %v", items[2])
				}
			},
		},
		{
			name: "add to nested array",
			initial: map[string]any{
				"properties": map[string]any{
					"wheels": []any{"wheel1", "wheel2"},
				},
			},
			patch: Patch{
				Path:      "properties.wheels",
				Value:     "wheel3",
				Operation: OpAdd,
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				props := result["properties"].(map[string]any)
				wheels := props["wheels"].([]any)
				if len(wheels) != 3 {
					t.Errorf("expected 3 wheels, got %d", len(wheels))
				}
				if wheels[2] != "wheel3" {
					t.Errorf("expected last wheel to be 'wheel3', got %v", wheels[2])
				}
			},
		},
		{
			name: "error on non-array path",
			initial: map[string]any{
				"value": "string",
			},
			patch: Patch{
				Path:      "value",
				Value:     "item",
				Operation: OpAdd,
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "error on missing path",
			initial: map[string]any{
				"key1": "value1",
			},
			patch: Patch{
				Path:      "nonexistent.array",
				Value:     "item",
				Operation: OpAdd,
			},
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Apply(tt.initial)
			if (err != nil) != tt.wantErr {
				t.Errorf("Patch.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, tt.initial)
			}
		})
	}
}

func TestGetPresetPatches(t *testing.T) {
	presets := GetPresetPatches()
	if len(presets) == 0 {
		t.Error("GetPresetPatches() returned no presets")
	}

	for _, preset := range presets {
		if preset.VehicleName == "" {
			t.Error("preset has empty VehicleName")
		}
		if preset.FilePattern == "" {
			t.Error("preset has empty FilePattern")
		}
		if len(preset.Patches) == 0 {
			t.Errorf("preset %s has no patches", preset.VehicleName)
		}
	}
}
