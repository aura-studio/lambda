package tests

import (
	"encoding/json"
	"testing"

	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: invoke-lambda-handler, Property 10: 调试模式响应格式**
// **Validates: Requirements 5.6**
//
// Property 10: 调试模式响应格式
// For any 启用调试模式的请求，响应 SHALL 包含 mode、raw_path、path、param、request、response、error 等调试字段。

// debugResponse represents the expected structure of a debug mode response
type debugResponse struct {
	Mode     string `json:"mode"`
	RawPath  string `json:"raw_path"`
	Path     string `json:"path"`
	Param    string `json:"param"`
	Request  string `json:"request"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error"`
}

// genDebugPathSegment generates a valid path segment for debug mode tests
// Uses alphanumeric characters only to ensure valid UTF-8 and URL-safe paths
func genDebugPathSegment() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		// Limit length to avoid overly long paths
		if len(s) > 12 {
			return s[:12]
		}
		if len(s) == 0 {
			return "test"
		}
		return s
	})
}

// genDebugAPIPath generates a debug API path like "/_/api/pkg/version/route"
func genDebugAPIPath() gopter.Gen {
	return gopter.CombineGens(
		genDebugPathSegment(), // pkg
		genDebugPathSegment(), // version
		genDebugPathSegment(), // route
	).Map(func(values []interface{}) string {
		pkg := values[0].(string)
		version := values[1].(string)
		route := values[2].(string)
		return "/_/api/" + pkg + "/" + version + "/" + route
	})
}

// genValidUTF8String generates a valid UTF-8 string using alphanumeric characters
// This ensures the string can be safely JSON encoded and decoded without modification
func genValidUTF8String() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		// Limit length to avoid overly long strings
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// genNonEmptyValidUTF8String generates a non-empty valid UTF-8 string
func genNonEmptyValidUTF8String() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) < 50
	})
}

// TestDebugModeResponseContainsRequiredFields tests that debug mode responses contain all required fields
// **Validates: Requirements 5.6**
func TestDebugModeResponseContainsRequiredFields(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode response contains mode, raw_path, path, param, request, error fields", prop.ForAll(
		func(pkg, version, route, request string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create debug API path
			rawPath := "/_/api/" + pkg + "/" + version + "/" + route

			// Create context with debug mode enabled
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: rawPath,
				Path:    rawPath,
				Request: request,
			}

			// Enable debug mode
			engine.Debug(ctx)

			if !ctx.DebugMode {
				t.Logf("Debug mode should be enabled after calling Debug handler")
				return false
			}

			// Simulate API handler behavior with debug mode
			// The API handler will call FormatDebug when there's an error
			ctx.ParamPath = "/" + pkg + "/" + version + "/" + route
			ctx.Err = nil // No error case

			// Call FormatDebugWithResponse to get debug output
			debugOutput := engine.FormatDebugWithResponse(ctx, "api", "test_response")

			// Parse the debug response
			var resp debugResponse
			if err := json.Unmarshal([]byte(debugOutput), &resp); err != nil {
				t.Logf("Failed to parse debug response: %v, output: %s", err, debugOutput)
				return false
			}

			// Verify all required fields are present
			if resp.Mode != "api" {
				t.Logf("Expected mode='api', got=%q", resp.Mode)
				return false
			}

			if resp.RawPath != rawPath {
				t.Logf("Expected raw_path=%q, got=%q", rawPath, resp.RawPath)
				return false
			}

			if resp.Path != rawPath {
				t.Logf("Expected path=%q, got=%q", rawPath, resp.Path)
				return false
			}

			if resp.Param != ctx.ParamPath {
				t.Logf("Expected param=%q, got=%q", ctx.ParamPath, resp.Param)
				return false
			}

			if resp.Request != request {
				t.Logf("Expected request=%q, got=%q", request, resp.Request)
				return false
			}

			return true
		},
		genDebugPathSegment(),
		genDebugPathSegment(),
		genDebugPathSegment(),
		genValidUTF8String(),
	))

	properties.TestingRun(t)
}

// TestDebugModeResponseWithError tests that debug mode responses include error information
// **Validates: Requirements 5.6**
func TestDebugModeResponseWithError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode response includes error field when error occurs", prop.ForAll(
		func(pkg, version, route, request, errorMsg string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create debug API path
			rawPath := "/_/api/" + pkg + "/" + version + "/" + route

			// Create context with debug mode enabled and an error
			ctx := &invoke.Context{
				Engine:    engine,
				RawPath:   rawPath,
				Path:      rawPath,
				Request:   request,
				ParamPath: "/" + pkg + "/" + version + "/" + route,
				DebugMode: true,
				Err:       &testError{msg: errorMsg},
			}

			// Call FormatDebug to get debug output with error
			debugOutput := engine.FormatDebug(ctx, "api")

			// Parse the debug response
			var resp debugResponse
			if err := json.Unmarshal([]byte(debugOutput), &resp); err != nil {
				t.Logf("Failed to parse debug response: %v, output: %s", err, debugOutput)
				return false
			}

			// Verify error field contains the error message
			if resp.Error != errorMsg {
				t.Logf("Expected error=%q, got=%q", errorMsg, resp.Error)
				return false
			}

			// Verify other required fields are still present
			if resp.Mode != "api" {
				t.Logf("Expected mode='api', got=%q", resp.Mode)
				return false
			}

			if resp.RawPath != rawPath {
				t.Logf("Expected raw_path=%q, got=%q", rawPath, resp.RawPath)
				return false
			}

			return true
		},
		genDebugPathSegment(),
		genDebugPathSegment(),
		genDebugPathSegment(),
		genValidUTF8String(),
		genNonEmptyValidUTF8String(),
	))

	properties.TestingRun(t)
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestDebugModeResponseWithResponse tests that debug mode responses include response field
// **Validates: Requirements 5.6**
func TestDebugModeResponseWithResponse(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode response includes response field when response is present", prop.ForAll(
		func(pkg, version, route, request, response string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create debug API path
			rawPath := "/_/api/" + pkg + "/" + version + "/" + route

			// Create context with debug mode enabled
			ctx := &invoke.Context{
				Engine:    engine,
				RawPath:   rawPath,
				Path:      rawPath,
				Request:   request,
				ParamPath: "/" + pkg + "/" + version + "/" + route,
				DebugMode: true,
			}

			// Call FormatDebugWithResponse to get debug output with response
			debugOutput := engine.FormatDebugWithResponse(ctx, "api", response)

			// Parse the debug response
			var resp debugResponse
			if err := json.Unmarshal([]byte(debugOutput), &resp); err != nil {
				t.Logf("Failed to parse debug response: %v, output: %s", err, debugOutput)
				return false
			}

			// Verify response field contains the response
			if resp.Response != response {
				t.Logf("Expected response=%q, got=%q", response, resp.Response)
				return false
			}

			// Verify all other required fields are present
			if resp.Mode != "api" {
				t.Logf("Expected mode='api', got=%q", resp.Mode)
				return false
			}

			if resp.RawPath != rawPath {
				t.Logf("Expected raw_path=%q, got=%q", rawPath, resp.RawPath)
				return false
			}

			if resp.Path != rawPath {
				t.Logf("Expected path=%q, got=%q", rawPath, resp.Path)
				return false
			}

			if resp.Param != ctx.ParamPath {
				t.Logf("Expected param=%q, got=%q", ctx.ParamPath, resp.Param)
				return false
			}

			if resp.Request != request {
				t.Logf("Expected request=%q, got=%q", request, resp.Request)
				return false
			}

			return true
		},
		genDebugPathSegment(),
		genDebugPathSegment(),
		genDebugPathSegment(),
		genValidUTF8String(),
		genValidUTF8String(),
	))

	properties.TestingRun(t)
}

// TestDebugModeHandlerEnablesDebugMode tests that the Debug handler enables debug mode
// **Validates: Requirements 5.6**
func TestDebugModeHandlerEnablesDebugMode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug handler enables DebugMode on context", prop.ForAll(
		func(rawPath, request string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create context with debug mode initially disabled
			ctx := &invoke.Context{
				Engine:    engine,
				RawPath:   rawPath,
				Path:      rawPath,
				Request:   request,
				DebugMode: false,
			}

			// Verify debug mode is initially disabled
			if ctx.DebugMode {
				t.Logf("DebugMode should be initially false")
				return false
			}

			// Call Debug handler
			engine.Debug(ctx)

			// Verify debug mode is now enabled
			if !ctx.DebugMode {
				t.Logf("DebugMode should be true after calling Debug handler")
				return false
			}

			return true
		},
		genDebugAPIPath(),
		genValidUTF8String(),
	))

	properties.TestingRun(t)
}

// TestDebugModeResponseIsValidJSON tests that debug mode responses are valid JSON
// **Validates: Requirements 5.6**
func TestDebugModeResponseIsValidJSON(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode response is valid JSON", prop.ForAll(
		func(pkg, version, route, request, response string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create debug API path
			rawPath := "/_/api/" + pkg + "/" + version + "/" + route

			// Create context
			ctx := &invoke.Context{
				Engine:    engine,
				RawPath:   rawPath,
				Path:      rawPath,
				Request:   request,
				ParamPath: "/" + pkg + "/" + version + "/" + route,
				DebugMode: true,
			}

			// Test FormatDebug output
			debugOutput := engine.FormatDebug(ctx, "api")
			var parsed1 map[string]interface{}
			if err := json.Unmarshal([]byte(debugOutput), &parsed1); err != nil {
				t.Logf("FormatDebug output is not valid JSON: %v, output: %s", err, debugOutput)
				return false
			}

			// Test FormatDebugWithResponse output
			debugOutputWithResponse := engine.FormatDebugWithResponse(ctx, "api", response)
			var parsed2 map[string]interface{}
			if err := json.Unmarshal([]byte(debugOutputWithResponse), &parsed2); err != nil {
				t.Logf("FormatDebugWithResponse output is not valid JSON: %v, output: %s", err, debugOutputWithResponse)
				return false
			}

			return true
		},
		genDebugPathSegment(),
		genDebugPathSegment(),
		genDebugPathSegment(),
		genValidUTF8String(),
		genValidUTF8String(),
	))

	properties.TestingRun(t)
}

// TestDebugModePreservesContextValues tests that debug mode response preserves all context values
// **Validates: Requirements 5.6**
func TestDebugModePreservesContextValues(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode response preserves all context values accurately", prop.ForAll(
		func(rawPath, path, paramPath, request, mode string) bool {
			// Create engine
			engine := invoke.NewEngine(nil, nil)

			// Create context with specific values
			ctx := &invoke.Context{
				Engine:    engine,
				RawPath:   rawPath,
				Path:      path,
				ParamPath: paramPath,
				Request:   request,
				DebugMode: true,
			}

			// Get debug output
			debugOutput := engine.FormatDebug(ctx, mode)

			// Parse the debug response
			var resp map[string]interface{}
			if err := json.Unmarshal([]byte(debugOutput), &resp); err != nil {
				t.Logf("Failed to parse debug response: %v", err)
				return false
			}

			// Verify each field matches the context value
			if resp["mode"] != mode {
				t.Logf("mode mismatch: expected=%q, got=%v", mode, resp["mode"])
				return false
			}

			if resp["raw_path"] != rawPath {
				t.Logf("raw_path mismatch: expected=%q, got=%v", rawPath, resp["raw_path"])
				return false
			}

			if resp["path"] != path {
				t.Logf("path mismatch: expected=%q, got=%v", path, resp["path"])
				return false
			}

			if resp["param"] != paramPath {
				t.Logf("param mismatch: expected=%q, got=%v", paramPath, resp["param"])
				return false
			}

			if resp["request"] != request {
				t.Logf("request mismatch: expected=%q, got=%v", request, resp["request"])
				return false
			}

			return true
		},
		genDebugAPIPath(),
		genDebugAPIPath(),
		genDebugPathSegment(),
		genValidUTF8String(),
		gen.OneConstOf("api", "wapi", "debug"),
	))

	properties.TestingRun(t)
}
