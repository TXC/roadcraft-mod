package patch

import (
	"fmt"
)

// Patch represents a modification to be applied to a CLS file
type Patch struct {
	// Name is a descriptive name for this patch
	Name string
	// Path is the dotted path to the value to modify (e.g., "properties.prop_truck_mobile_sand_screen.allowedPercent")
	Path string
	// Value is the new value to set (can be a string, number, map[string]any for objects, or []any for arrays)
	Value any
	// Operation defines how to apply the patch (set, delete, etc.)
	Operation PatchOperation
}

// PatchOperation defines the type of modification
type PatchOperation string

const (
	// OpSet sets or replaces a value at the given path
	OpSet PatchOperation = "set"
	// OpDelete removes a value at the given path
	OpDelete PatchOperation = "delete"
	// OpAdd adds to an array at the given path
	OpAdd PatchOperation = "add"
)

// VehiclePatch represents all patches for a specific vehicle
type VehiclePatch struct {
	// Name of the vehicle (for display purposes)
	VehicleName string
	// FilePattern is the pattern to match files (e.g., "*zikz_605e*.cls")
	FilePattern string
	// Patches to apply to this vehicle
	Patches []Patch
}

// Apply applies a patch to a parsed CLS map
func (p *Patch) Apply(clsMap map[string]any) error {
	switch p.Operation {
	case OpSet:
		return p.applySet(clsMap)
	case OpDelete:
		return p.applyDelete(clsMap)
	case OpAdd:
		return p.applyAdd(clsMap)
	default:
		return fmt.Errorf("unknown operation: %s", p.Operation)
	}
}

// applySet sets a value at the given path, supporting array indexing
func (p *Patch) applySet(clsMap map[string]any) error {
	keys := parsePath(p.Path)
	if len(keys) == 0 {
		return fmt.Errorf("empty path")
	}

	// Navigate to the parent
	current := any(clsMap)
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]

		// Check if this is an array index
		if len(key) > 2 && key[0] == '[' && key[len(key)-1] == ']' {
			// Parse array index
			indexStr := key[1 : len(key)-1]
			var index int
			if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
				return fmt.Errorf("invalid array index: %s", key)
			}

			// Current must be an array
			arr, ok := current.([]any)
			if !ok {
				return fmt.Errorf("path segment %s is not an array", keys[i-1])
			}
			if index < 0 || index >= len(arr) {
				return fmt.Errorf("array index %d out of bounds (length %d)", index, len(arr))
			}
			current = arr[index]
		} else {
			// Regular map access
			currentMap, ok := current.(map[string]any)
			if !ok {
				return fmt.Errorf("path segment is not a map at key %s", key)
			}

			next, ok := currentMap[key]
			if !ok {
				// Create missing intermediate maps
				newMap := make(map[string]any)
				currentMap[key] = newMap
				current = newMap
				continue
			}
			current = next
		}
	}

	// Set the final value
	finalKey := keys[len(keys)-1]

	// Check if final key is an array index
	if len(finalKey) > 2 && finalKey[0] == '[' && finalKey[len(finalKey)-1] == ']' {
		// Parse array index
		indexStr := finalKey[1 : len(finalKey)-1]
		var index int
		if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
			return fmt.Errorf("invalid array index: %s", finalKey)
		}

		// Current must be an array
		arr, ok := current.([]any)
		if !ok {
			return fmt.Errorf("cannot index non-array with %s", finalKey)
		}
		if index < 0 || index >= len(arr) {
			return fmt.Errorf("array index %d out of bounds (length %d)", index, len(arr))
		}
		arr[index] = p.Value
	} else {
		// Regular map key
		currentMap, ok := current.(map[string]any)
		if !ok {
			return fmt.Errorf("cannot set key %s on non-map", finalKey)
		}
		currentMap[finalKey] = p.Value
	}

	return nil
}

// applyDelete removes a value at the given path
func (p *Patch) applyDelete(clsMap map[string]any) error {
	keys := parsePath(p.Path)
	if len(keys) == 0 {
		return fmt.Errorf("empty path")
	}

	// Navigate to the parent
	current := clsMap
	for i := 0; i < len(keys)-1; i++ {
		next, ok := current[keys[i]]
		if !ok {
			return fmt.Errorf("path not found: %s", p.Path)
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("path segment %s is not a map", keys[i])
		}
		current = nextMap
	}

	// Delete the final key
	finalKey := keys[len(keys)-1]
	delete(current, finalKey)
	return nil
}

// applyAdd adds a value to an array at the given path
func (p *Patch) applyAdd(clsMap map[string]any) error {
	keys := parsePath(p.Path)
	if len(keys) == 0 {
		return fmt.Errorf("empty path")
	}

	// Navigate to the parent
	current := clsMap
	for i := 0; i < len(keys)-1; i++ {
		next, ok := current[keys[i]]
		if !ok {
			return fmt.Errorf("path not found: %s", p.Path)
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("path segment %s is not a map", keys[i])
		}
		current = nextMap
	}

	// Add to the array
	finalKey := keys[len(keys)-1]
	arr, ok := current[finalKey].([]any)
	if !ok {
		return fmt.Errorf("path %s is not an array", p.Path)
	}
	current[finalKey] = append(arr, p.Value)
	return nil
}

// parsePath splits a dotted path into keys, supporting array indexing with [n] notation
// Examples:
//   - "key1.key2.key3" -> ["key1", "key2", "key3"]
//   - "gears[3].speed" -> ["gears", "[3]", "speed"]
//   - "items[0]" -> ["items", "[0]"]
func parsePath(path string) []string {
	if path == "" {
		return nil
	}
	keys := []string{}
	current := ""
	inBracket := false

	for _, c := range path {
		if c == '[' {
			// Save the key before the bracket
			if current != "" {
				keys = append(keys, current)
				//current = ""
			}
			inBracket = true
			current = "["
		} else if c == ']' {
			current += "]"
			keys = append(keys, current)
			current = ""
			inBracket = false
		} else if c == '.' && !inBracket {
			if current != "" {
				keys = append(keys, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		keys = append(keys, current)
	}
	return keys
}
