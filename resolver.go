package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadSlots reads the JSON data used to fill in the profile templates
func LoadSlots(filepath string) (map[string]interface{}, error) {
	if filepath == "" {
		return make(map[string]interface{}), nil
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read slots file: %w", err)
	}

	var slots map[string]interface{}
	if err := json.Unmarshal(data, &slots); err != nil {
		return nil, fmt.Errorf("failed to parse slots JSON: %w", err)
	}
	return slots, nil
}

// getExpectedType resolves a local "#/components/schemas/..." reference and returns its "type"
func getExpectedType(doc map[string]interface{}, ref string) string {
	if !strings.HasPrefix(ref, "#/") {
		return "" // MVP: Only resolving local refs
	}

	parts := strings.Split(ref[2:], "/")
	var current interface{} = doc

	// Traverse the OpenAPI tree to find the schema
	for _, part := range parts {
		m, ok := toMap(current)
		if !ok {
			return ""
		}
		current, ok = m[part]
		if !ok {
			return ""
		}
	}

	schemaMap, ok := toMap(current)
	if ok {
		if t, isStr := schemaMap["type"].(string); isStr {
			return t
		}
	}
	return ""
}

// checkType compares the injected Go type against the expected OpenAPI type
func checkType(val interface{}, expectedType string) error {
	if expectedType == "" {
		return nil // Skip if schema type is undefined or complex
	}
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("expected string")
		}
	case "number", "integer":
		// Go's json.Unmarshal decodes all numbers as float64 by default
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("expected number/integer")
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected boolean")
		}
	case "object":
		if _, ok := toMap(val); !ok {
			return fmt.Errorf("expected object")
		}
	case "array":
		if _, ok := val.([]interface{}); !ok {
			return fmt.Errorf("expected array")
		}
	}
	return nil
}

// ResolveNode recursively walks the YAML/JSON tree and resolves $slot objects
func ResolveNode(node interface{}, slots map[string]interface{}, doc map[string]interface{}) (interface{}, error) {
	switch v := node.(type) {

	case map[string]interface{}:
		// Check if this map is a Typed Slot Placeholder
		if slotNameRaw, ok := v["$slot"]; ok {
			slotName := fmt.Sprintf("%v", slotNameRaw)

			isRequired := true
			if reqRaw, hasReq := v["required"]; hasReq {
				if reqBool, isBool := reqRaw.(bool); isBool {
					isRequired = reqBool
				}
			}

			resolvedValue, exists := slots[slotName]
			if !exists {
				if isRequired {
					return nil, fmt.Errorf("Required slot '%s' was not provided in the slots data", slotName)
				}
				return nil, nil // Omit non-required missing slots
			}

			// TYPE/SCHEMA CHECK
			if schemaRefRaw, hasRef := v["schemaRef"]; hasRef {
				schemaRef := fmt.Sprintf("%v", schemaRefRaw)
				expectedType := getExpectedType(doc, schemaRef)

				if err := checkType(resolvedValue, expectedType); err != nil {
					return nil, fmt.Errorf("Type validation failed for slot '%s': %v, but got %T", slotName, err, resolvedValue)
				}
			}

			return resolvedValue, nil
		}

		resolvedMap := make(map[string]interface{})
		for key, val := range v {
			resolvedVal, err := ResolveNode(val, slots, doc)
			if err != nil {
				return nil, err
			}
			if resolvedVal != nil {
				resolvedMap[key] = resolvedVal
			}
		}
		return resolvedMap, nil

	case []interface{}:
		var resolvedArray []interface{}
		for _, val := range v {
			resolvedVal, err := ResolveNode(val, slots, doc)
			if err != nil {
				return nil, err
			}
			if resolvedVal != nil {
				resolvedArray = append(resolvedArray, resolvedVal)
			}
		}
		return resolvedArray, nil

	default:
		return v, nil
	}
}
