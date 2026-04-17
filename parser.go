package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// toMap safely casts interface{} to map[string]interface{}
func toMap(v interface{}) (map[string]interface{}, bool) {
	// Standard map
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	// Generic interface map (fallback)
	if m, ok := v.(map[interface{}]interface{}); ok {
		res := make(map[string]interface{})
		for k, val := range m {
			res[fmt.Sprintf("%v", k)] = val
		}
		return res, true
	}
	return nil, false
}

// LoadSpec reads the YAML file from disk and unmarshals it natively
func LoadSpec(filepath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return doc, nil
}

// FindOperationSecurity scans the paths object to find the target operationId
func FindOperationSecurity(doc map[string]interface{}, targetOpID string) ([]interface{}, error) {
	pathsRaw, exists := doc["paths"]
	if !exists {
		return nil, fmt.Errorf("the 'paths' key does not exist at the root of the document")
	}

	paths, ok := toMap(pathsRaw)
	if !ok {
		return nil, fmt.Errorf("the 'paths' object is not a valid map, it is type %T", pathsRaw)
	}

	for pathKey, pathItemRaw := range paths {
		pathItem, ok := toMap(pathItemRaw)
		if !ok {
			continue
		}

		// Check all HTTP methods under this path (get, post and others)
		for method, operationRaw := range pathItem {
			operation, ok := toMap(operationRaw)
			if !ok {
				continue
			}

			// Check if this operation has the ID we are looking for
			if opID, exists := operation["operationId"]; exists && opID == targetOpID {
				// extract its security array
				securityRaw, hasSec := operation["security"]
				if !hasSec {
					return nil, fmt.Errorf("operation '%s' found at %s %s, but it has no security requirements", targetOpID, method, pathKey)
				}

				securityArray, ok := securityRaw.([]interface{})
				if !ok {
					return nil, fmt.Errorf("security requirement for '%s' is not an array", targetOpID)
				}

				return securityArray, nil
			}
		}
	}

	return nil, fmt.Errorf("operationId '%s' not found in specification", targetOpID)
}

// FindGNAPProfile fetches a specific profile array from the x-gnap-access-profiles catalog
func FindGNAPProfile(doc map[string]interface{}, profileName string) ([]interface{}, error) {
	componentsRaw, ok := doc["components"]
	if !ok {
		return nil, fmt.Errorf("no 'components' object found")
	}
	components, _ := toMap(componentsRaw)

	secSchemesRaw, ok := components["securitySchemes"]
	if !ok {
		return nil, fmt.Errorf("no 'securitySchemes' object found")
	}
	secSchemes, _ := toMap(secSchemesRaw)

	gnapRaw, ok := secSchemes["GNAP"]
	if !ok {
		return nil, fmt.Errorf("GNAP security scheme not found")
	}
	gnap, _ := toMap(gnapRaw)

	profilesCatalogRaw, ok := gnap["x-gnap-access-profiles"]
	if !ok {
		return nil, fmt.Errorf("x-gnap-access-profiles catalog not found in GNAP scheme")
	}
	profilesCatalog, _ := toMap(profilesCatalogRaw)

	profileRaw, ok := profilesCatalog[profileName]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found in catalog", profileName)
	}

	profileArr, ok := profileRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("profile '%s' is not an array", profileName)
	}

	return profileArr, nil
}
