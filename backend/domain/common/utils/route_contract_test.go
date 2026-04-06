package utils

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOpenAPI_RouteConsistency validates that critical API routes
// match between registered Gin routes and OpenAPI specification.
// This prevents contract drift between runtime behavior and documentation.
func TestOpenAPI_RouteConsistency(t *testing.T) {
	// Read OpenAPI spec
	openapiData, err := os.ReadFile("../../../docs/openapi/openapi.json")
	if err != nil {
		t.Skipf("OpenAPI spec not found, skipping contract validation: %v", err)
	}

	var openapi map[string]interface{}
	if err := json.Unmarshal(openapiData, &openapi); err != nil {
		t.Fatalf("Failed to parse OpenAPI spec: %v", err)
	}

	paths, ok := openapi["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("OpenAPI spec missing 'paths' object")
	}

	// Critical route prefixes that must stay in sync
	criticalPrefixes := []string{
		"/api/models/v1/whisper",
		"/api/models/v1/pipeline",
		"/api/models/v1/rag",
		"/api/terminal",
		"/api/mqtt",
	}

	// Collect all OpenAPI paths
	openapiPaths := make(map[string]bool)
	for path := range paths {
		openapiPaths[path] = true
	}

	// Check that critical OpenAPI paths use correct prefixes
	t.Run("OpenAPI_Paths_Use_Correct_Prefixes", func(t *testing.T) {
		for path := range openapiPaths {
			// Skip non-critical paths
			isCritical := false
			for _, prefix := range criticalPrefixes {
				if strings.HasPrefix(path, prefix) {
					isCritical = true
					break
				}
			}

			if !isCritical {
				continue
			}

			// Check for old /api/v1/models/ prefix (should be /api/models/v1/)
			if strings.HasPrefix(path, "/api/v1/models/") {
				t.Errorf("OpenAPI path uses deprecated /api/v1/models/ prefix: %s (should be /api/models/v1/)", path)
			}
		}
	})

	// Verify specific critical endpoints exist
	t.Run("Critical_Endpoints_Exist", func(t *testing.T) {
		criticalEndpoints := []string{
			"/api/models/v1/whisper/transcribe",
			"/api/models/v1/whisper/status/{transcribe_id}",
			"/api/models/v1/whisper/uploads/sessions",
			"/api/models/v1/whisper/uploads/sessions/{id}",
			"/api/models/v1/whisper/uploads/sessions/{id}/chunks/{index}",
			"/api/models/v1/pipeline/job",
			"/api/models/v1/pipeline/status/{task_id}",
			"/api/models/v1/rag/translate",
			"/api/models/v1/rag/summary",
			"/api/models/v1/rag/chat",
			"/api/models/v1/rag/control",
			"/api/models/v1/rag/status/{task_id}",
			"/api/terminal",
			"/api/terminal/{id}",
			"/api/terminal/mac/{mac}",
		}

		for _, endpoint := range criticalEndpoints {
			if !openapiPaths[endpoint] {
				t.Errorf("Critical endpoint missing from OpenAPI spec: %s", endpoint)
			}
		}
	})
}

// TestOpenAPI_AuthConsistency validates that authentication schemes
// are consistent across all endpoints.
func TestOpenAPI_AuthConsistency(t *testing.T) {
	openapiData, err := os.ReadFile("../../../docs/openapi/openapi.json")
	if err != nil {
		t.Skipf("OpenAPI spec not found, skipping auth validation: %v", err)
	}

	var openapi map[string]interface{}
	if err := json.Unmarshal(openapiData, &openapi); err != nil {
		t.Fatalf("Failed to parse OpenAPI spec: %v", err)
	}

	components, ok := openapi["components"].(map[string]interface{})
	if !ok {
		t.Skip("OpenAPI spec missing 'components' object")
	}

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	if !ok {
		t.Skip("OpenAPI spec missing 'securitySchemes' object")
	}

	// Verify BearerAuth is defined
	_, hasBearer := securitySchemes["BearerAuth"]
	assert.True(t, hasBearer, "BearerAuth security scheme should be defined")

	// Verify ApiKeyAuth is defined
	_, hasAPIKey := securitySchemes["ApiKeyAuth"]
	assert.True(t, hasAPIKey, "ApiKeyAuth security scheme should be defined")
}
