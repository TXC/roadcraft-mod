package parser

import (
	"fmt"
	"sort"
	"strings"
)

// SerializeClsFile converts a parsed CLS map back to CLS file format
func SerializeClsFile(clsMap map[string]any) ([]byte, error) {
	var sb strings.Builder
	if err := serializeObject(&sb, clsMap, 0); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// serializeObject serializes a map with proper indentation
func serializeObject(sb *strings.Builder, obj map[string]any, indent int) error {
	// Sort keys for consistent output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]
		writeIndent(sb, indent)
		sb.WriteString(key)
		sb.WriteString(" = ")

		if err := serializeValue(sb, value, indent); err != nil {
			return err
		}
		sb.WriteString("\n")
	}

	return nil
}

// serializeValue serializes a value based on its type
func serializeValue(sb *strings.Builder, value any, indent int) error {
	switch v := value.(type) {
	case map[string]any:
		sb.WriteString("{\n")
		if err := serializeObject(sb, v, indent+1); err != nil {
			return err
		}
		writeIndent(sb, indent)
		sb.WriteString("}")

	case []any:
		sb.WriteString("[\n")
		for i, item := range v {
			writeIndent(sb, indent+1)
			if err := serializeValue(sb, item, indent+1); err != nil {
				return err
			}
			if i < len(v)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		writeIndent(sb, indent)
		sb.WriteString("]")

	case string:
		// Write string as-is (already includes quotes if needed)
		_, err := sb.WriteString(v)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}

	return nil
}

// writeIndent writes the appropriate indentation
func writeIndent(sb *strings.Builder, indent int) {
	for range indent {
		sb.WriteString("   ")
	}
}
