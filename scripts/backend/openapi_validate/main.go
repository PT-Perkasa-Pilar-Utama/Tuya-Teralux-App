package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenAPI 3.1.0 meta schema URL
const openapi31SchemaURL = "https://spec.openapis.org/oas/3.1/dialect/base"

func main() {
	// Look for openapi.json in docs/openapi directory
	backendDir := findBackendDir()
	openapiDir := filepath.Join(backendDir, "docs", "openapi")
	jsonFile := filepath.Join(openapiDir, "openapi.json")
	yamlFile := filepath.Join(openapiDir, "openapi.yaml")

	// Check if files exist
	if !fileExists(jsonFile) && !fileExists(yamlFile) {
		fmt.Fprintf(os.Stderr, "❌ Error: No OpenAPI files found in %s\n", openapiDir)
		fmt.Fprintf(os.Stderr, "   Run 'make generate-openapi' first\n")
		os.Exit(1)
	}

	// Validate JSON file if exists
	if fileExists(jsonFile) {
		fmt.Printf("🔍 Validating %s...\n", jsonFile)
		if err := validateOpenAPI(jsonFile); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ JSON validation passed")
	}

	// Validate YAML file if exists
	if fileExists(yamlFile) {
		fmt.Printf("🔍 Validating %s...\n", yamlFile)
		if err := validateOpenAPI(yamlFile); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ YAML validation passed")
	}

	fmt.Println("\n✅ All OpenAPI validations passed!")
}

func findBackendDir() string {
	// Start from current directory and look for docs/openapi
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Check if we're already in backend dir
	if exists(filepath.Join(dir, "docs", "openapi")) {
		return dir
	}

	// Check if we're in backend/scripts
	if strings.HasSuffix(dir, "backend/scripts") {
		return filepath.Dir(dir)
	}

	// Try parent directories
	for i := 0; i < 5; i++ {
		if exists(filepath.Join(dir, "docs", "openapi")) {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	return "."
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateOpenAPI(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var spec map[string]interface{}

	// Determine file type and parse accordingly
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("error parsing YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}
	}

	// Validate OpenAPI version
	version, ok := spec["openapi"]
	if !ok {
		return fmt.Errorf("missing 'openapi' field")
	}

	versionStr, ok := version.(string)
	if !ok {
		return fmt.Errorf("'openapi' field must be a string")
	}

	if !strings.HasPrefix(versionStr, "3.1") {
		return fmt.Errorf("expected OpenAPI version 3.1.x, got %s", versionStr)
	}

	// Validate required fields
	if err := validateRequiredFields(spec); err != nil {
		return err
	}

	// Validate structure
	if err := validateStructure(spec); err != nil {
		return err
	}

	return nil
}

func validateRequiredFields(spec map[string]interface{}) error {
	// Check for required 'info' field
	info, ok := spec["info"]
	if !ok {
		return fmt.Errorf("missing required field: 'info'")
	}

	infoMap, ok := info.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'info' must be an object")
	}

	// Check for required 'title' in info
	if _, ok := infoMap["title"]; !ok {
		return fmt.Errorf("missing required field: 'info.title'")
	}

	// Check for required 'version' in info
	if _, ok := infoMap["version"]; !ok {
		return fmt.Errorf("missing required field: 'info.version'")
	}

	// Check for 'paths' field (can be empty object)
	if _, ok := spec["paths"]; !ok {
		return fmt.Errorf("missing required field: 'paths'")
	}

	return nil
}

func validateStructure(spec map[string]interface{}) error {
	// Validate paths structure
	if paths, ok := spec["paths"].(map[string]interface{}); ok {
		for path, pathItem := range paths {
			if !strings.HasPrefix(path, "/") {
				return fmt.Errorf("path must start with '/': %s", path)
			}

			pathItemMap, ok := pathItem.(map[string]interface{})
			if !ok {
				return fmt.Errorf("path item must be an object: %s", path)
			}

			// Validate HTTP methods
			validMethods := map[string]bool{
				"get": true, "put": true, "post": true,
				"delete": true, "options": true, "head": true,
				"patch": true, "trace": true,
			}

			for key := range pathItemMap {
				if validMethods[key] {
					operation, ok := pathItemMap[key].(map[string]interface{})
					if !ok {
						return fmt.Errorf("operation must be an object: %s %s", key, path)
					}
					if err := validateOperation(operation, key, path); err != nil {
						return err
					}
				}
			}
		}
	}

	// Validate components if present
	if components, ok := spec["components"].(map[string]interface{}); ok {
		if err := validateComponents(components); err != nil {
			return err
		}
	}

	return nil
}

func validateOperation(operation map[string]interface{}, method, path string) error {
	// Check for required 'responses' field
	if _, ok := operation["responses"]; !ok {
		return fmt.Errorf("missing 'responses' in %s %s", strings.ToUpper(method), path)
	}

	responses, ok := operation["responses"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("'responses' must be an object in %s %s", strings.ToUpper(method), path)
	}

	// Validate response codes
	for code := range responses {
		if code != "default" && !strings.HasPrefix(code, "2") &&
			!strings.HasPrefix(code, "3") && !strings.HasPrefix(code, "4") &&
			!strings.HasPrefix(code, "5") {
			return fmt.Errorf("invalid response code '%s' in %s %s", code, strings.ToUpper(method), path)
		}
	}

	return nil
}

func validateComponents(components map[string]interface{}) error {
	validKeys := map[string]bool{
		"schemas": true, "parameters": true, "responses": true,
		"securitySchemes": true, "examples": true, "requestBodies": true,
		"headers": true, "pathItems": true, "links": true, "callbacks": true,
	}

	for key := range components {
		if !validKeys[key] {
			return fmt.Errorf("invalid components key: %s", key)
		}
	}

	return nil
}
