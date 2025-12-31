package parser

import (
	"fmt"
	"strings"
)

// ParseClsFile parses a CLS file into a nested map structure.
// Example: parsed["properties"]["prop_truck_mobile_sand_screen"]["allowedPercent"] -> "0.4"
// This preserves all values as strings without modification.
func ParseClsFile(content []byte) (map[string]any, error) {
	text := string(content)
	result := make(map[string]any)

	// Start parsing from the beginning
	if err := parseClsObject(text, result); err != nil {
		return nil, err
	}

	return result, nil
}

// parseClsObject recursively parses a CLS object structure into a nested map
func parseClsObject(text string, result map[string]any) error {
	// Remove leading/trailing whitespace
	text = strings.TrimSpace(text)

	// Track position in text
	pos := 0

	for pos < len(text) {
		// Skip whitespace
		for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r') {
			pos++
		}

		if pos >= len(text) {
			break
		}

		// Check for closing brace or bracket
		if text[pos] == '}' || text[pos] == ']' {
			break
		}

		// Find the key name (alphanumeric, underscore, or dot)
		keyStart := pos
		for pos < len(text) && (isKeyChar(text[pos])) {
			pos++
		}

		if pos == keyStart {
			// No key found, skip character
			pos++
			continue
		}

		key := strings.TrimSpace(text[keyStart:pos])

		// Skip whitespace
		for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
			pos++
		}

		// Expect '=' sign
		if pos >= len(text) || text[pos] != '=' {
			// Skip to next line if no equals sign
			for pos < len(text) && text[pos] != '\n' {
				pos++
			}
			continue
		}
		pos++ // skip '='

		// Skip whitespace after '='
		for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
			pos++
		}

		if pos >= len(text) {
			break
		}

		// Check what kind of value we have
		switch text[pos] {
		case '{':
			// Nested object
			pos++ // skip '{'
			nestedEnd := findMatchingBrace(text[pos:])
			if nestedEnd == -1 {
				return fmt.Errorf("unmatched opening brace at position %d", pos)
			}
			nestedContent := text[pos : pos+nestedEnd]
			pos += nestedEnd + 1 // skip past '}'

			// Create nested map and recursively parse
			nestedMap := make(map[string]any)
			if err := parseClsObject(nestedContent, nestedMap); err != nil {
				return err
			}
			result[key] = nestedMap
		case '[':
			// Array - parse as slice of items
			pos++ // skip '['
			arrayEnd := findMatchingBracket(text[pos:])
			if arrayEnd == -1 {
				return fmt.Errorf("unmatched opening bracket at position %d", pos)
			}
			arrayContent := text[pos : pos+arrayEnd]
			pos += arrayEnd + 1 // skip past ']'

			// Parse array items
			arrayItems, err := parseClsArray(arrayContent)
			if err != nil {
				return err
			}
			result[key] = arrayItems
		default:
			// Simple value (string, number, boolean)
			valueStart := pos

			// Check if it's a quoted string
			if text[pos] == '"' {
				pos++ // skip opening quote
				for pos < len(text) && text[pos] != '"' {
					if text[pos] == '\\' {
						pos++ // skip escaped character
					}
					pos++
				}
				if pos < len(text) {
					pos++ // skip closing quote
				}
			} else {
				// Read until newline, comma, or closing brace/bracket
				for pos < len(text) && text[pos] != '\n' && text[pos] != '\r' && text[pos] != ',' && text[pos] != '}' && text[pos] != ']' {
					pos++
				}
			}

			value := strings.TrimSpace(text[valueStart:pos])
			result[key] = value
		}

		// Skip optional comma
		for pos < len(text) && (text[pos] == ',' || text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r') {
			pos++
		}
	}

	return nil
}

// parseClsArray parses array content into a slice
func parseClsArray(text string) ([]any, error) {
	text = strings.TrimSpace(text)
	var result []any
	pos := 0

	for pos < len(text) {
		// Skip whitespace and commas
		for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r' || text[pos] == ',') {
			pos++
		}

		if pos >= len(text) {
			break
		}

		// Check for nested object in array
		switch text[pos] {
		case '{':
			pos++ // skip '{'
			nestedEnd := findMatchingBrace(text[pos:])
			if nestedEnd == -1 {
				return nil, fmt.Errorf("unmatched opening brace in array at position %d", pos)
			}
			nestedContent := text[pos : pos+nestedEnd]
			pos += nestedEnd + 1 // skip past '}'

			nestedMap := make(map[string]any)
			if err := parseClsObject(nestedContent, nestedMap); err != nil {
				return nil, err
			}
			result = append(result, nestedMap)
		case '[':
			// Nested array
			pos++ // skip '['
			arrayEnd := findMatchingBracket(text[pos:])
			if arrayEnd == -1 {
				return nil, fmt.Errorf("unmatched opening bracket in array at position %d", pos)
			}
			arrayContent := text[pos : pos+arrayEnd]
			pos += arrayEnd + 1 // skip past ']'

			nestedArray, err := parseClsArray(arrayContent)
			if err != nil {
				return nil, err
			}
			result = append(result, nestedArray)
		default:
			// Simple value
			valueStart := pos

			if text[pos] == '"' {
				pos++ // skip opening quote
				for pos < len(text) && text[pos] != '"' {
					if text[pos] == '\\' {
						pos++ // skip escaped character
					}
					pos++
				}
				if pos < len(text) {
					pos++ // skip closing quote
				}
			} else {
				// Read until comma, newline, or closing bracket
				for pos < len(text) && text[pos] != ',' && text[pos] != '\n' && text[pos] != '\r' && text[pos] != ']' && text[pos] != '}' {
					pos++
				}
			}

			value := strings.TrimSpace(text[valueStart:pos])
			if value != "" {
				result = append(result, value)
			}
		}
	}

	return result, nil
}

// isKeyChar returns true if the character is valid in a key name
func isKeyChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// findMatchingBrace finds the position of the matching closing brace
func findMatchingBrace(text string) int {
	depth := 1
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// findMatchingBracket finds the position of the matching closing bracket
func findMatchingBracket(text string) int {
	depth := 1
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
