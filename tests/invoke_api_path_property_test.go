package tests

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// **Feature: invoke-lambda-handler, Property 7: API 路径解析正确性**
// **Validates: Requirements 8.2, 8.3**
//
// Property 7: API 路径解析正确性
// For any 格式为 `/api/{pkg}/{version}/{route}` 的请求路径，引擎 SHALL 正确解析出包名、版本号和路由路径，
// 并通过 Dynamic 组件调用对应的业务包。

// mockAPITunnel implements dynamic.Tunnel for testing API path parsing
type mockAPITunnel struct {
	pkg     string
	version string
	invoked bool
	route   string
	req     string
}

func (m *mockAPITunnel) Init() {}

func (m *mockAPITunnel) Invoke(route string, req string) string {
	m.invoked = true
	m.route = route
	m.req = req
	return fmt.Sprintf("response:pkg=%s,version=%s,route=%s,req=%s", m.pkg, m.version, route, req)
}

func (m *mockAPITunnel) Meta() string {
	return fmt.Sprintf("meta:%s_%s", m.pkg, m.version)
}

func (m *mockAPITunnel) Close() {}

// apiPathInput represents the input data for generating API paths
type apiPathInput struct {
	Pkg     string
	Version string
	Route   string
}

// genLowercaseAlphaNum generates lowercase alphanumeric strings starting with a letter
// This matches the dynamic package's allowed keyword pattern: ^[a-z0-9][a-z0-9-]*$
func genLowercaseAlphaNum(minLen, maxLen int) gopter.Gen {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	firstChars := "abcdefghijklmnopqrstuvwxyz"

	return gen.IntRange(minLen, maxLen).FlatMap(func(length interface{}) gopter.Gen {
		l := length.(int)
		return gopter.CombineGens(
			gen.IntRange(0, len(firstChars)-1),
			gen.SliceOfN(l-1, gen.IntRange(0, len(chars)-1)),
		).Map(func(values []any) string {
			firstIdx := values[0].(int)
			restIdxs := values[1].([]int)
			result := string(firstChars[firstIdx])
			for _, idx := range restIdxs {
				result += string(chars[idx])
			}
			return result
		})
	}, reflect.TypeOf(""))
}

func genValidPkgName() gopter.Gen {
	return genLowercaseAlphaNum(2, 10)
}

func genValidVersionName() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("v1"),
		gen.Const("v2"),
		gen.Const("latest"),
		genLowercaseAlphaNum(2, 8),
	)
}

func genValidRoutePath() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(""),
		genLowercaseAlphaNum(1, 20),
		gopter.CombineGens(
			genLowercaseAlphaNum(1, 10),
			genLowercaseAlphaNum(1, 10),
		).Map(func(values []any) string {
			return values[0].(string) + "/" + values[1].(string)
		}),
	)
}

func genAPIPathInputData() gopter.Gen {
	return gopter.CombineGens(
		genValidPkgName(),
		genValidVersionName(),
		genValidRoutePath(),
	).Map(func(values []any) apiPathInput {
		return apiPathInput{
			Pkg:     values[0].(string),
			Version: values[1].(string),
			Route:   values[2].(string),
		}
	})
}

func buildFullAPIPath(input apiPathInput) string {
	if input.Route == "" {
		return fmt.Sprintf("/api/%s/%s", input.Pkg, input.Version)
	}
	return fmt.Sprintf("/api/%s/%s/%s", input.Pkg, input.Version, input.Route)
}

func expectedRoutePath(input apiPathInput) string {
	if input.Route == "" {
		return "/"
	}
	return "/" + input.Route
}

// TestAPIPathParsingCorrectness tests that valid API paths are correctly parsed
// and the correct package, version, and route are extracted.
// **Validates: Requirements 8.2, 8.3**
func TestAPIPathParsingCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("API path parsing: valid paths are correctly parsed into pkg, version, and route", prop.ForAll(
		func(input apiPathInput, requestPayload string) bool {
			tunnel := &mockAPITunnel{
				pkg:     input.Pkg,
				version: input.Version,
			}

			dynamic.RegisterPackage(input.Pkg, input.Version, tunnel)

			engine := invoke.NewEngine(nil, []dynamicpkg.Option{
				dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
					Package: input.Pkg,
					Version: input.Version,
					Tunnel:  tunnel,
				}),
			})

			apiPath := buildFullAPIPath(input)

			req := &invoke.Request{
				CorrelationId: "test-api-parsing",
				Path:          apiPath,
				Payload:       []byte(requestPayload),
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

			if resp.Error != "" {
				t.Logf("Unexpected error: %s for path: %s", resp.Error, apiPath)
				return false
			}

			if !tunnel.invoked {
				t.Logf("Tunnel was not invoked for path: %s", apiPath)
				return false
			}

			expected := expectedRoutePath(input)
			if tunnel.route != expected {
				t.Logf("Route mismatch: expected %q, got %q", expected, tunnel.route)
				return false
			}

			if tunnel.req != requestPayload {
				t.Logf("Request mismatch: expected %q, got %q", requestPayload, tunnel.req)
				return false
			}

			expectedResponse := fmt.Sprintf("response:pkg=%s,version=%s,route=%s,req=%s",
				input.Pkg, input.Version, expected, requestPayload)
			if string(resp.Payload) != expectedResponse {
				t.Logf("Response mismatch: expected %q, got %q", expectedResponse, string(resp.Payload))
				return false
			}

			return true
		},
		genAPIPathInputData(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) <= 50 }),
	))

	properties.TestingRun(t)
}

// TestAPIPathParsingInvalidPaths tests that invalid API paths return errors
// **Validates: Requirements 8.2, 8.4**
func TestAPIPathParsingInvalidPaths(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("API path parsing: paths with missing pkg or version return errors", prop.ForAll(
		func(singleSegment string) bool {
			engine := invoke.NewEngine(nil, nil)

			apiPath := "/api/" + singleSegment

			req := &invoke.Request{
				CorrelationId: "test-invalid-path",
				Path:          apiPath,
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

			if resp.Error == "" {
				t.Logf("Expected error for path with single segment: %s", apiPath)
				return false
			}

			if !strings.Contains(resp.Error, "invalid path") {
				t.Logf("Expected 'invalid path' error, got: %s", resp.Error)
				return false
			}

			return true
		},
		genValidPkgName(),
	))

	properties.TestingRun(t)
}

// TestAPIPathParsingEmptyPath tests that empty API paths return errors
// **Validates: Requirements 8.2**
func TestAPIPathParsingEmptyPath(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("API path parsing: empty paths return error", prop.ForAll(
		func(_ bool) bool {
			engine := invoke.NewEngine(nil, nil)

			apiPath := "/api/"

			req := &invoke.Request{
				CorrelationId: "test-empty-path",
				Path:          apiPath,
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

			if resp.Error == "" {
				t.Logf("Expected error for empty path")
				return false
			}

			if !strings.Contains(resp.Error, "invalid path") && !strings.Contains(resp.Error, "missing api path") {
				t.Logf("Expected path-related error, got: %s", resp.Error)
				return false
			}

			return true
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// TestAPIPathParsingPackageNotFound tests that non-existent packages return errors
// **Validates: Requirements 8.4**
func TestAPIPathParsingPackageNotFound(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("API path parsing: non-existent packages return errors", prop.ForAll(
		func(input apiPathInput) bool {
			engine := invoke.NewEngine(nil, nil)

			uniquePkg := "notfound" + input.Pkg
			uniqueInput := apiPathInput{
				Pkg:     uniquePkg,
				Version: input.Version,
				Route:   input.Route,
			}
			apiPath := buildFullAPIPath(uniqueInput)

			req := &invoke.Request{
				CorrelationId: "test-pkg-not-found",
				Path:          apiPath,
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

			if resp.Error == "" {
				t.Logf("Expected error for non-existent package: %s/%s", uniquePkg, input.Version)
				return false
			}

			return true
		},
		genAPIPathInputData(),
	))

	properties.TestingRun(t)
}

// TestAPIPathParsingRouteExtraction tests that route paths are correctly extracted
// **Validates: Requirements 8.2**
func TestAPIPathParsingRouteExtraction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("API path parsing: route paths with multiple segments are correctly extracted", prop.ForAll(
		func(pkg string, version string, routeSegments []string) bool {
			var validSegments []string
			for _, s := range routeSegments {
				if s != "" {
					validSegments = append(validSegments, s)
				}
			}

			tunnel := &mockAPITunnel{
				pkg:     pkg,
				version: version,
			}

			dynamic.RegisterPackage(pkg, version, tunnel)

			engine := invoke.NewEngine(nil, []dynamicpkg.Option{
				dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
					Package: pkg,
					Version: version,
					Tunnel:  tunnel,
				}),
			})

			route := strings.Join(validSegments, "/")

			var apiPath string
			if route == "" {
				apiPath = fmt.Sprintf("/api/%s/%s", pkg, version)
			} else {
				apiPath = fmt.Sprintf("/api/%s/%s/%s", pkg, version, route)
			}

			req := &invoke.Request{
				CorrelationId: "test-route-extraction",
				Path:          apiPath,
				Payload:       []byte("test"),
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

			if resp.Error != "" {
				t.Logf("Unexpected error: %s", resp.Error)
				return false
			}

			expectedRoute := "/"
			if route != "" {
				expectedRoute = "/" + route
			}
			if tunnel.route != expectedRoute {
				t.Logf("Route mismatch: expected %q, got %q", expectedRoute, tunnel.route)
				return false
			}

			return true
		},
		genValidPkgName(),
		genValidVersionName(),
		gen.SliceOfN(3, genLowercaseAlphaNum(1, 10)),
	))

	properties.TestingRun(t)
}
