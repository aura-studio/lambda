package tests

import (
	"context"
	"testing"

	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// **Feature: invoke-lambda-handler, Property 8: 错误响应一致性**
// **Validates: Requirements 1.5, 8.4**
//
// Property 8: 错误响应一致性
// For any 导致错误的请求（无效路径、不存在的包、处理器异常等），
// 引擎 SHALL 返回包含错误信息的 Response，且 error 字段非空。

// TestErrorResponseInvalidProtobuf tests that invalid protobuf payloads return error responses
// **Validates: Requirements 1.5**
func TestErrorResponseInvalidProtobuf(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: invalid protobuf payload returns non-empty error", prop.ForAll(
		func(invalidPayload []byte) bool {
			// Skip empty payloads as they might be valid edge cases
			if len(invalidPayload) == 0 {
				return true
			}

			engine := invoke.NewEngine(nil, nil)

			// Invoke with invalid protobuf data
			respBytes, err := engine.Invoke(context.Background(), invalidPayload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for invalid protobuf
			if resp.Error == "" {
				t.Logf("Expected non-empty error for invalid protobuf payload")
				return false
			}

			return true
		},
		// Generate random bytes that are unlikely to be valid protobuf
		gen.SliceOfN(50, gen.UInt8()).SuchThat(func(b []byte) bool {
			// Ensure it's not accidentally valid protobuf by checking if it fails to unmarshal
			var req invoke.Request
			return proto.Unmarshal(b, &req) != nil
		}),
	))

	properties.TestingRun(t)
}

// TestErrorResponseNonExistentRoute tests that non-existent routes return 404 error responses
// **Validates: Requirements 1.5**
func TestErrorResponseNonExistentRoute(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: non-existent routes return non-empty error", prop.ForAll(
		func(randomPath string) bool {
			engine := invoke.NewEngine(nil, nil)

			// Create a request with a path that doesn't match any route
			// Avoid paths that match existing routes like /health-check, /api/*, etc.
			path := "/nonexistent/" + randomPath

			req := &invoke.Request{
				CorrelationId: "test-404",
				Path:          path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for 404
			if resp.Error == "" {
				t.Logf("Expected non-empty error for non-existent route: %s", path)
				return false
			}

			return true
		},
		genLowercaseAlphaNum(1, 20),
	))

	properties.TestingRun(t)
}

// TestErrorResponseNonExistentPackage tests that non-existent packages return error responses
// **Validates: Requirements 8.4**
func TestErrorResponseNonExistentPackage(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: non-existent packages return non-empty error", prop.ForAll(
		func(pkg string, version string) bool {
			engine := invoke.NewEngine(nil, nil)

			// Create a request for a package that doesn't exist
			path := "/api/" + pkg + "/" + version + "/route"

			req := &invoke.Request{
				CorrelationId: "test-pkg-not-found",
				Path:          path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for non-existent package
			if resp.Error == "" {
				t.Logf("Expected non-empty error for non-existent package: %s/%s", pkg, version)
				return false
			}

			return true
		},
		genLowercaseAlphaNum(3, 10).Map(func(s string) string { return "notfound" + s }),
		gen.OneConstOf("v1", "v2", "latest"),
	))

	properties.TestingRun(t)
}


// TestErrorResponseInvalidAPIPath tests that invalid API path formats return error responses
// **Validates: Requirements 8.4**
func TestErrorResponseInvalidAPIPath(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: invalid API path format returns non-empty error", prop.ForAll(
		func(singleSegment string) bool {
			engine := invoke.NewEngine(nil, nil)

			// Create a request with invalid API path (missing version)
			path := "/api/" + singleSegment

			req := &invoke.Request{
				CorrelationId: "test-invalid-api-path",
				Path:          path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for invalid API path
			if resp.Error == "" {
				t.Logf("Expected non-empty error for invalid API path: %s", path)
				return false
			}

			return true
		},
		genLowercaseAlphaNum(2, 10),
	))

	properties.TestingRun(t)
}

// TestErrorResponseEngineStopped tests that stopped engine returns error responses
// **Validates: Requirements 1.5**
func TestErrorResponseEngineStopped(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: stopped engine returns non-empty error", prop.ForAll(
		func(path string, correlationId string) bool {
			engine := invoke.NewEngine(nil, nil)

			// Stop the engine
			engine.Stop()

			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty when engine is stopped
			if resp.Error == "" {
				t.Logf("Expected non-empty error when engine is stopped")
				return false
			}

			// Error should specifically indicate engine is stopped
			if resp.Error != "engine is stopped" {
				t.Logf("Expected 'engine is stopped' error, got: %s", resp.Error)
				return false
			}

			return true
		},
		gen.OneConstOf("/health-check", "/api/pkg/v1/route", "/nonexistent"),
		genLowercaseAlphaNum(1, 20),
	))

	properties.TestingRun(t)
}

// TestErrorResponseMissingAPIPath tests that empty API path returns error responses
// **Validates: Requirements 8.4**
func TestErrorResponseMissingAPIPath(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: empty API path returns non-empty error", prop.ForAll(
		func(_ bool) bool {
			engine := invoke.NewEngine(nil, nil)

			// Create a request with empty API path
			path := "/api/"

			req := &invoke.Request{
				CorrelationId: "test-empty-api-path",
				Path:          path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for empty API path
			if resp.Error == "" {
				t.Logf("Expected non-empty error for empty API path")
				return false
			}

			return true
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// TestErrorResponseCorrelationIdPreserved tests that correlation ID is preserved in error responses
// **Validates: Requirements 1.5**
func TestErrorResponseCorrelationIdPreserved(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Error response: correlation ID is preserved in error responses", prop.ForAll(
		func(correlationId string) bool {
			engine := invoke.NewEngine(nil, nil)

			// Create a request that will cause an error (non-existent route)
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          "/nonexistent/path",
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Correlation ID should be preserved
			if resp.CorrelationId != correlationId {
				t.Logf("Correlation ID mismatch: expected %q, got %q", correlationId, resp.CorrelationId)
				return false
			}

			// Error field should be non-empty
			if resp.Error == "" {
				t.Logf("Expected non-empty error for non-existent route")
				return false
			}

			return true
		},
		genLowercaseAlphaNum(1, 50),
	))

	properties.TestingRun(t)
}

// TestErrorResponseAllErrorScenarios tests that all error scenarios return non-empty error
// **Validates: Requirements 1.5, 8.4**
func TestErrorResponseAllErrorScenarios(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	// errorScenario represents different error scenarios
	type errorScenario struct {
		name        string
		path        string
		stopEngine  bool
		description string
	}

	// Define error scenarios as a slice
	errorScenarios := []errorScenario{
		{
			name:        "non-existent-route",
			path:        "/nonexistent/path",
			stopEngine:  false,
			description: "non-existent route",
		},
		{
			name:        "non-existent-package",
			path:        "/api/notfoundpkg/v1/route",
			stopEngine:  false,
			description: "non-existent package",
		},
		{
			name:        "invalid-api-path",
			path:        "/api/onlypkg",
			stopEngine:  false,
			description: "invalid API path format",
		},
		{
			name:        "empty-api-path",
			path:        "/api/",
			stopEngine:  false,
			description: "empty API path",
		},
		{
			name:        "engine-stopped",
			path:        "/health-check",
			stopEngine:  true,
			description: "engine stopped",
		},
	}

	genErrorScenario := func() gopter.Gen {
		return gen.IntRange(0, len(errorScenarios)-1).Map(func(i int) errorScenario {
			return errorScenarios[i]
		})
	}

	// Generate simple correlation IDs without filtering
	genSimpleCorrelationId := func() gopter.Gen {
		return gen.Identifier().Map(func(s string) string {
			if len(s) > 20 {
				return s[:20]
			}
			return s
		})
	}

	properties.Property("Error response: all error scenarios return non-empty error field", prop.ForAll(
		func(scenario errorScenario, correlationId string) bool {
			engine := invoke.NewEngine(nil, nil)

			if scenario.stopEngine {
				engine.Stop()
			}

			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          scenario.path,
				Payload:       []byte("{}"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Error field should be non-empty for all error scenarios
			if resp.Error == "" {
				t.Logf("Expected non-empty error for scenario %q (%s)", scenario.name, scenario.description)
				return false
			}

			return true
		},
		genErrorScenario(),
		genSimpleCorrelationId(),
	))

	properties.TestingRun(t)
}

// Ensure dynamic package is imported
var _ = dynamicpkg.Option(nil)
