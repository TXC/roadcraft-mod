package parser

import (
	"strings"
	"testing"
)

func TestSerializeClsFile(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		expect  string
		wantErr bool
	}{
		{
			name: "simple key-value",
			input: map[string]any{
				"name": "test",
			},
			expect:  "name = test\n",
			wantErr: false,
		},
		{
			name: "nested object",
			input: map[string]any{
				"properties": map[string]any{
					"prop1": "value1",
					"prop2": "value2",
				},
			},
			expect:  "properties = {\n   prop1 = value1\n   prop2 = value2\n}\n",
			wantErr: false,
		},
		{
			name: "array with simple values",
			input: map[string]any{
				"items": []any{"item1", "item2", "item3"},
			},
			expect:  "items = [\n   item1,\n   item2,\n   item3\n]\n",
			wantErr: false,
		},
		{
			name: "array with objects",
			input: map[string]any{
				"gears": []any{
					map[string]any{"speed": "10"},
					map[string]any{"speed": "20"},
				},
			},
			expect:  "gears = [\n   {\n      speed = 10\n   },\n   {\n      speed = 20\n   }\n]\n",
			wantErr: false,
		},
		{
			name: "complex nested structure",
			input: map[string]any{
				"properties": map[string]any{
					"geom": map[string]any{
						"nameTpl": "test_vehicle",
					},
					"prop_truck": map[string]any{
						"wheels": []any{
							map[string]any{
								"diameter": "1.2",
								"width":    "0.5",
							},
						},
					},
				},
			},
			expect:  "properties = {\n   geom = {\n      nameTpl = test_vehicle\n   }\n   prop_truck = {\n      wheels = [\n         {\n            diameter = 1.2\n            width = 0.5\n         }\n      ]\n   }\n}\n",
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]any{},
			expect:  "",
			wantErr: false,
		},
		{
			name: "quoted strings",
			input: map[string]any{
				"text": "\"quoted value\"",
			},
			expect:  "text = \"quoted value\"\n",
			wantErr: false,
		},
		{
			name: "boolean values",
			input: map[string]any{
				"enabled": "True",
				"visible": "False",
			},
			expect:  "enabled = True\nvisible = False\n",
			wantErr: false,
		},
		{
			name: "numeric values",
			input: map[string]any{
				"count":   "42",
				"ratio":   "1.5",
				"percent": "0.4",
			},
			expect:  "count = 42\npercent = 0.4\nratio = 1.5\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SerializeClsFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeClsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != tt.expect {
				t.Errorf("SerializeClsFile() failed\n  actual:\n%s\n  expected:\n%s", string(result), tt.expect)
			}
		})
	}
}

func TestSerializeObject(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]any
		expect  string
		wantErr bool
	}{
		{
			name: "set simple value",
			initial: map[string]any{
				"key1": "value1",
			},
			expect:  "key1 = value1\n",
			wantErr: false,
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
			expect:  "properties = {\n   prop1 = {\n      value = old\n   }\n}\n",
			wantErr: false,
		},
		{
			name: "set array element",
			initial: map[string]any{
				"items": []any{"old1", "old2", "old3"},
			},
			expect:  "items = [\n   old1,\n   old2,\n   old3\n]\n",
			wantErr: false,
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
			expect:  "gears = [\n   {\n      ratio = 1.0\n      speed = 10\n   },\n   {\n      ratio = 2.0\n      speed = 20\n   },\n   {\n      ratio = 3.0\n      speed = 30\n   }\n]\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			err := serializeObject(&sb, tt.initial, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("serializeObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sb.String() != tt.expect {
				t.Errorf("serializeObject() failed\n  actual: '%v'\n expected: '%v'", sb.String(), tt.expect)
			}
		})
	}
}

func TestSerializeValueIndentation(t *testing.T) {
	tests := []struct {
		name   string
		indent int
		expect string
	}{
		{
			name:   "no indentation",
			indent: 0,
			expect: "",
		},
		{
			name:   "one level",
			indent: 1,
			expect: "   ",
		},
		{
			name:   "two levels",
			indent: 2,
			expect: "      ",
		},
		{
			name:   "three levels",
			indent: 3,
			expect: "         ",
		},
		{
			name:   "five levels",
			indent: 5,
			expect: "               ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			writeIndent(&sb, tt.indent)
			if sb.String() != tt.expect {
				t.Errorf("writeIndent() = '%v', want '%v' (len: %d vs %d)", sb.String(), tt.expect, len(sb.String()), len(tt.expect))
			}
		})
	}
}

func TestSerializeValue(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		indent  int
		expect  string
		wantErr bool
	}{
		{
			name:    "simple string",
			value:   "test",
			indent:  0,
			expect:  "test",
			wantErr: false,
		},
		{
			name:    "quoted string",
			value:   "\"quoted text\"",
			indent:  0,
			expect:  "\"quoted text\"",
			wantErr: false,
		},
		{
			name:    "numeric string",
			value:   "123.45",
			indent:  0,
			expect:  "123.45",
			wantErr: false,
		},
		{
			name: "simple object",
			value: map[string]any{
				"key": "value",
			},
			indent:  0,
			expect:  "{\n   key = value\n}",
			wantErr: false,
		},
		{
			name: "nested object",
			value: map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
			},
			indent:  0,
			expect:  "{\n   outer = {\n      inner = value\n   }\n}",
			wantErr: false,
		},
		{
			name: "object with indent",
			value: map[string]any{
				"key": "value",
			},
			indent:  1,
			expect:  "{\n      key = value\n   }",
			wantErr: false,
		},
		{
			name:    "simple array",
			value:   []any{"a", "b", "c"},
			indent:  0,
			expect:  "[\n   a,\n   b,\n   c\n]",
			wantErr: false,
		},
		{
			name:    "array with one element",
			value:   []any{"single"},
			indent:  0,
			expect:  "[\n   single\n]",
			wantErr: false,
		},
		{
			name:    "empty array",
			value:   []any{},
			indent:  0,
			expect:  "[\n]",
			wantErr: false,
		},
		{
			name: "array with objects",
			value: []any{
				map[string]any{"id": "1"},
				map[string]any{"id": "2"},
			},
			indent:  0,
			expect:  "[\n   {\n      id = 1\n   },\n   {\n      id = 2\n   }\n]",
			wantErr: false,
		},
		{
			name: "array with nested arrays",
			value: []any{
				[]any{"a", "b"},
				[]any{"c", "d"},
			},
			indent:  0,
			expect:  "[\n   [\n      a,\n      b\n   ],\n   [\n      c,\n      d\n   ]\n]",
			wantErr: false,
		},
		{
			name:    "unsupported type - int",
			value:   123,
			indent:  0,
			expect:  "",
			wantErr: true,
		},
		{
			name:    "unsupported type - bool",
			value:   true,
			indent:  0,
			expect:  "",
			wantErr: true,
		},
		{
			name:    "unsupported type - nil",
			value:   nil,
			indent:  0,
			expect:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			err := serializeValue(&sb, tt.value, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("serializeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sb.String() != tt.expect {
				t.Errorf("serializeValue() failed\n  actual: '%v'\n expected: '%v'", sb.String(), tt.expect)
			}
		})
	}
}

func TestWriteIndent(t *testing.T) {
	tests := []struct {
		name   string
		indent int
		expect string
	}{
		{
			name:   "no indentation",
			indent: 0,
			expect: "",
		},
		{
			name:   "one level",
			indent: 1,
			expect: "   ",
		},
		{
			name:   "two levels",
			indent: 2,
			expect: "      ",
		},
		{
			name:   "three levels",
			indent: 3,
			expect: "         ",
		},
		{
			name:   "five levels",
			indent: 5,
			expect: "               ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			writeIndent(&sb, tt.indent)
			if sb.String() != tt.expect {
				t.Errorf("writeIndent() = '%v', want '%v' (len: %d vs %d)", sb.String(), tt.expect, len(sb.String()), len(tt.expect))
			}
		})
	}
}

// TestRoundTrip tests that parsing and serializing produces consistent results
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple properties",
			input: `properties = {
   geom = {
      nameTpl = "test_vehicle"
   }
   prop_truck = {
      strength = 0.8
   }
}`,
		},
		{
			name: "array with objects",
			input: `gears = [
   {
      angularVelocity = 12
      gearRatio = 1.5
   },
   {
      angularVelocity = 16
      gearRatio = 1.0
   }
]`,
		},
		{
			name: "mixed structure",
			input: `properties = {
   enabled = True
   items = [
      "item1",
      "item2",
      "item3"
   ]
   nested = {
      value = 42
   }
}`,
		},
		{
			name: "deeply nested",
			input: `root = {
   level1 = {
      level2 = {
         level3 = {
            value = "deep"
         }
      }
   }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			parsed, err := ParseClsFile([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseClsFile() error = %v", err)
			}

			// Serialize it back
			serialized, err := SerializeClsFile(parsed)
			if err != nil {
				t.Fatalf("SerializeClsFile() error = %v", err)
			}

			// Parse the serialized version again
			reparsed, err := ParseClsFile(serialized)
			if err != nil {
				t.Fatalf("ParseClsFile() second time error = %v", err)
			}

			// The two parsed versions should be functionally equivalent
			// We'll do a deep comparison of the structure
			if !deepEqual(parsed, reparsed) {
				t.Errorf("Round-trip failed: structures don't match\nOriginal:\n%s\nSerialized:\n%s", tt.input, string(serialized))
			}
		})
	}
}

// deepEqual recursively compares two values
func deepEqual(a, b any) bool {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	case []any:
		vb, ok := b.([]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	default:
		return false
	}
}

// TestSerializeEdgeCases tests edge cases and error conditions
func TestSerializeEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
	}{
		{
			name: "very deep nesting",
			input: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": map[string]any{
							"l4": map[string]any{
								"l5": map[string]any{
									"value": "deep",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "keys with special characters",
			input: map[string]any{
				"prop_truck_mobile_sand_screen": "value",
				"__type":                        "Vector2dRnd",
			},
			wantErr: false,
		},
		{
			name: "mixed array types",
			input: map[string]any{
				"mixed": []any{
					"string",
					map[string]any{"nested": "object"},
					[]any{"nested", "array"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty nested structures",
			input: map[string]any{
				"empty_obj":   map[string]any{},
				"empty_array": []any{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SerializeClsFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeClsFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
