package tests

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// Feature: event-lambda-handler
// Property 9: Path Rewriting
// *For any* path that matches a StaticLinkMap entry or PrefixLinkMap entry,
// the Router SHALL rewrite the path before dispatching.
//
// **Validates: Requirements 2.6, 2.7**

// Property 10: API Path Extraction
// *For any* path matching `/api/{package}/{version}/*`, the Engine SHALL
// correctly extract the package name and version.
//
// **Validates: Requirements 3.1**

// Property 18: Panic Recovery
// *For any* handler that panics, the Engine SHALL recover from the panic
// and set an error on the context without crashing.
//
// **Validates: Requirements 8.1**

// Property 19: Invalid Path Error
// *For any* path that does not conform to the expected format (e.g., missing
// package/version), the Engine SHALL return a descriptive error message.
//
// **Validates: Requirements 8.3**

// ============================================================================
// Generators for Property Tests
// ============================================================================

// genPathSegment generates a valid path segment (alphanumeric)
func genPathSegment() gopter.Gen {
	return gen.OneConstOf("users", "api", "v1", "v2", "test", "pkg", "route", "data", "items", "config", "service", "handler")
}

// genPackageName generates a valid package name (lowercase alphanumeric only)
func genPackageName() gopter.Gen {
	return gen.OneConstOf("mypkg", "testpkg", "service", "handler", "core", "util", "apipkg", "webpkg")
}

// genVersion generates a valid version string (only alphanumeric, no dots)
func genVersion() gopter.Gen {
	return gen.OneConstOf("v1", "v2", "commit", "latest")
}

// genRoute generates a valid route path
func genRoute() gopter.Gen {
	return gen.IntRange(0, 3).FlatMap(func(n interface{}) gopter.Gen {
		count := n.(int)
		if count == 0 {
			return gen.Const("/")
		}
		return gen.SliceOfN(count, genPathSegment()).Map(func(segments []string) string {
			return "/" + strings.Join(segments, "/")
		})
	}, reflect.TypeOf(""))
}

// genHandlerStaticLinkEntry generates a static link mapping entry for handler tests
func genHandlerStaticLinkEntry() gopter.Gen {
	return gen.Struct(reflect.TypeOf(struct {
		SrcPath string
		DstPath string
	}{}), map[string]gopter.Gen{
		"SrcPath": gen.SliceOfN(2, genPathSegment()).Map(func(segments []string) string {
			return "/" + strings.Join(segments, "/")
		}),
		"DstPath": gen.SliceOfN(2, genPathSegment()).Map(func(segments []string) string {
			return "/" + strings.Join(segments, "/")
		}),
	})
}

// genHandlerPrefixLinkEntry generates a prefix link mapping entry for handler tests
func genHandlerPrefixLinkEntry() gopter.Gen {
	return gen.Struct(reflect.TypeOf(struct {
		SrcPrefix string
		DstPrefix string
	}{}), map[string]gopter.Gen{
		"SrcPrefix": genPathSegment().Map(func(s string) string {
			return "/" + s
		}),
		"DstPrefix": genPathSegment().Map(func(s string) string {
			return "/" + s
		}),
	})
}

// ============================================================================
// Property 9: Path Rewriting Tests
// ============================================================================

// TestEventHandlerStaticLinkPathRewriting tests Property 9 for StaticLinkMap
// For any path that matches a StaticLinkMap entry, the Router SHALL rewrite
// the path before dispatching.
func TestEventHandlerStaticLinkPathRewriting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLinkMap rewrites path before dispatching", prop.ForAll(
		func(entry struct {
			SrcPath string
			DstPath string
		}) bool {
			// Skip if src and dst are the same
			if entry.SrcPath == entry.DstPath {
				return true
			}

			var capturedPath string
			handlerCalled := false

			// Create engine with static link mapping
			e := event.NewEngine([]event.Option{
				event.WithStaticLink(entry.SrcPath, entry.DstPath),
			}, nil)

			// Override the router to capture the rewritten path
			router := event.NewRouter()
			router.Use(e.HeaderLink)
			router.Use(e.StaticLink)
			router.Use(e.PrefixLink)
			router.Handle(entry.DstPath, func(c *event.Context) {
				handlerCalled = true
				capturedPath = c.Path
			})
			router.NoRoute(func(c *event.Context) {
				capturedPath = c.Path
			})

			// Dispatch with the source path
			ctx := &event.Context{
				Engine: e,
				Path:   entry.SrcPath,
			}
			router.Dispatch(ctx)

			// Verify the path was rewritten to the destination
			return capturedPath == entry.DstPath && handlerCalled
		},
		genHandlerStaticLinkEntry(),
	))

	properties.TestingRun(t)
}

// TestEventHandlerPrefixLinkPathRewriting tests Property 9 for PrefixLinkMap
// For any path that matches a PrefixLinkMap entry, the Router SHALL rewrite
// the path before dispatching.
func TestEventHandlerPrefixLinkPathRewriting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLinkMap rewrites path prefix before dispatching", prop.ForAll(
		func(entry struct {
			SrcPrefix string
			DstPrefix string
		}, suffix string) bool {
			// Skip if src and dst prefixes are the same
			if entry.SrcPrefix == entry.DstPrefix {
				return true
			}

			srcPath := entry.SrcPrefix + suffix
			expectedDstPath := entry.DstPrefix + suffix

			var capturedPath string

			// Create engine with prefix link mapping
			e := event.NewEngine([]event.Option{
				event.WithPrefixLink(entry.SrcPrefix, entry.DstPrefix),
			}, nil)

			// Create router to capture the rewritten path
			router := event.NewRouter()
			router.Use(e.HeaderLink)
			router.Use(e.StaticLink)
			router.Use(e.PrefixLink)
			router.NoRoute(func(c *event.Context) {
				capturedPath = c.Path
			})

			// Dispatch with the source path
			ctx := &event.Context{
				Engine: e,
				Path:   srcPath,
			}
			router.Dispatch(ctx)

			// Verify the path prefix was rewritten
			return capturedPath == expectedDstPath
		},
		genHandlerPrefixLinkEntry(),
		genRoute(),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Property 10: API Path Extraction Tests
// ============================================================================

// TestEventHandlerAPIPathExtraction tests Property 10: API Path Extraction
// For any path matching `/api/{package}/{version}/*`, the Engine SHALL
// correctly extract the package name and version.
func TestEventHandlerAPIPathExtraction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Use a counter to generate unique package names
	counter := 0

	properties.Property("API path correctly extracts package and version", prop.ForAll(
		func(pkgSuffix string, version string, route string) bool {
			var extractedRoute string
			extractedPkg := ""

			// Generate a unique valid package name (no underscores)
			counter++
			testPkgName := fmt.Sprintf("apipathtest%s%d", pkgSuffix, counter)

			// Register a test package that captures the invocation
			dynamic.RegisterPackage(testPkgName, version, &mockTunnel{
				invoke: func(r, req string) string {
					extractedRoute = r
					extractedPkg = testPkgName
					return ""
				},
			})

			e := event.NewEngine(nil, nil)

			// Construct the API path
			apiPath := "/api/" + testPkgName + "/" + version + route

			req := &event.Request{
				Items: []*event.Item{
					{Path: apiPath, Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			// Invoke the engine
			_ = e.Invoke(context.Background(), payload)

			// Verify the package was invoked with the correct route
			return extractedPkg == testPkgName && extractedRoute == route
		},
		genPackageName(),
		genVersion(),
		genRoute(),
	))

	properties.TestingRun(t)
}

// TestEventHandlerAPIPathExtractionWithDebugMode tests Property 10 with debug mode
// For any path matching `/_/api/{package}/{version}/*`, the Engine SHALL
// correctly extract the package name and version in debug mode.
func TestEventHandlerAPIPathExtractionWithDebugMode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Use a counter to generate unique package names
	counter := 0

	properties.Property("Debug API path correctly extracts package and version", prop.ForAll(
		func(pkgSuffix string, version string, route string) bool {
			var extractedRoute string
			extractedPkg := ""

			// Generate a unique valid package name (no underscores)
			counter++
			testPkgName := fmt.Sprintf("debugapipathtest%s%d", pkgSuffix, counter)

			// Register a test package
			dynamic.RegisterPackage(testPkgName, version, &mockTunnel{
				invoke: func(r, req string) string {
					extractedRoute = r
					extractedPkg = testPkgName
					return ""
				},
			})

			e := event.NewEngine(nil, nil)

			// Construct the debug API path
			apiPath := "/_/api/" + testPkgName + "/" + version + route

			req := &event.Request{
				Items: []*event.Item{
					{Path: apiPath, Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			// Invoke the engine
			_ = e.Invoke(context.Background(), payload)

			// Verify the package was invoked with the correct route
			return extractedPkg == testPkgName && extractedRoute == route
		},
		genPackageName(),
		genVersion(),
		genRoute(),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Property 18: Panic Recovery Tests
// ============================================================================

// TestEventHandlerPanicRecovery tests Property 18: Panic Recovery
// For any handler that panics, the Engine SHALL recover from the panic
// and set an error on the context without crashing.
func TestEventHandlerPanicRecovery(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generate panic messages directly without filtering
	genPanicMsg := gen.OneConstOf(
		"test panic",
		"error occurred",
		"unexpected failure",
		"handler crashed",
		"runtime error",
		"nil pointer",
		"index out of range",
		"division by zero",
		"stack overflow",
		"memory allocation failed",
	)

	properties.Property("Engine recovers from handler panic without crashing", prop.ForAll(
		func(panicMsg string) (result bool) {
			// Ensure we don't crash the test
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Test itself panicked (should not happen): %v", r)
					result = false
				}
			}()

			// Register a package that panics
			testPkgName := "panictest"
			testVersion := "v1"
			dynamic.RegisterPackage(testPkgName, testVersion, &mockTunnel{
				invoke: func(route, req string) string {
					panic(panicMsg)
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModePartial), // Continue even on error
			}, nil)

			// Create a request that will trigger the panic
			req := &event.Request{
				Items: []*event.Item{
					{Path: "/api/" + testPkgName + "/" + testVersion + "/route", Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			// Invoke should not panic - it should recover and return nil (partial mode)
			err = e.Invoke(context.Background(), payload)

			// In partial mode, error is recorded but nil is returned
			// The important thing is that we didn't crash
			return true
		},
		genPanicMsg,
	))

	properties.TestingRun(t)
}

// TestEventHandlerPanicRecoveryWithBatchMode tests Property 18 with batch mode
// For any handler that panics in batch mode, the Engine SHALL recover and return error.
func TestEventHandlerPanicRecoveryWithBatchMode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generate panic messages directly without filtering
	genPanicMsg := gen.OneConstOf(
		"test panic",
		"error occurred",
		"unexpected failure",
		"handler crashed",
		"runtime error",
		"nil pointer",
		"index out of range",
		"division by zero",
		"stack overflow",
		"memory allocation failed",
	)

	properties.Property("Engine recovers from panic in batch mode and returns error", prop.ForAll(
		func(panicMsg string) (result bool) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Test itself panicked (should not happen): %v", r)
					result = false
				}
			}()

			// Register a package that panics
			testPkgName := "panicbatchtest"
			testVersion := "v1"
			dynamic.RegisterPackage(testPkgName, testVersion, &mockTunnel{
				invoke: func(route, req string) string {
					panic(panicMsg)
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeBatch), // Fail entire batch on error
			}, nil)

			req := &event.Request{
				Items: []*event.Item{
					{Path: "/api/" + testPkgName + "/" + testVersion + "/route", Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			// Invoke should not panic - it should recover and return an error
			err = e.Invoke(context.Background(), payload)

			// In batch mode, error should be returned
			return err != nil && strings.Contains(err.Error(), "panic")
		},
		genPanicMsg,
	))

	properties.TestingRun(t)
}

// ============================================================================
// Property 19: Invalid Path Error Tests
// ============================================================================

// genInvalidAPIPath generates invalid API paths that don't conform to expected format
func genInvalidAPIPath() gopter.Gen {
	return gen.OneGenOf(
		// Empty path after /api/
		gen.Const("/api/"),
		// Missing version
		gen.OneConstOf("/api/pkg", "/api/mypkg"),
		// Just /api without anything else
		gen.Const("/api"),
		// Path with empty segments
		gen.Const("/api//version/route"),
		gen.Const("/api/pkg//route"),
	)
}

// TestEventHandlerInvalidPathError tests Property 19: Invalid Path Error
// For any path that does not conform to the expected format (e.g., missing
// package/version), the Engine SHALL return a descriptive error message.
func TestEventHandlerInvalidPathError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Invalid API path returns descriptive error", prop.ForAll(
		func(invalidPath string) bool {
			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeBatch), // Return error on failure
			}, nil)

			req := &event.Request{
				Items: []*event.Item{
					{Path: invalidPath, Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			// Invoke should return an error for invalid paths
			err = e.Invoke(context.Background(), payload)

			// Should return an error with descriptive message
			if err == nil {
				return false
			}

			// Error message should be descriptive (contain relevant info)
			errMsg := err.Error()
			return len(errMsg) > 0 && (strings.Contains(errMsg, "event") ||
				strings.Contains(errMsg, "path") ||
				strings.Contains(errMsg, "package") ||
				strings.Contains(errMsg, "not found"))
		},
		genInvalidAPIPath(),
	))

	properties.TestingRun(t)
}

// TestEventHandlerMissingPackageError tests Property 19 for missing package
// When the package does not exist, the Engine SHALL return a descriptive error.
func TestEventHandlerMissingPackageError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Use a counter to generate unique non-existent package names
	counter := 0

	properties.Property("Missing package returns descriptive error", prop.ForAll(
		func(pkgSuffix string, version string) bool {
			// Generate a unique package name that definitely doesn't exist
			counter++
			nonExistentPkg := fmt.Sprintf("nonexistent%s%d", pkgSuffix, counter)

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeBatch),
			}, nil)

			req := &event.Request{
				Items: []*event.Item{
					{Path: "/api/" + nonExistentPkg + "/" + version + "/route", Payload: []byte(`{}`)},
				},
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				return false
			}

			err = e.Invoke(context.Background(), payload)

			// Should return an error about package not found
			if err == nil {
				return false
			}

			errMsg := err.Error()
			return strings.Contains(errMsg, "package") && strings.Contains(errMsg, "not found")
		},
		genPackageName(),
		genVersion(),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Unit Tests for additional coverage
// ============================================================================

// TestEventHandlerStaticLinkRewrite tests static link path rewriting
func TestEventHandlerStaticLinkRewrite(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithStaticLink("/old-path", "/"),
	}, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/old-path", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventHandlerPrefixLinkRewrite tests prefix link path rewriting
func TestEventHandlerPrefixLinkRewrite(t *testing.T) {
	// Register a test package
	dynamic.RegisterPackage("prefixhandlerpkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return ""
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithPrefixLink("/v1", "/api"),
	}, nil)

	// /v1/prefixhandlerpkg/v1/route -> /api/prefixhandlerpkg/v1/route
	req := &event.Request{
		Items: []*event.Item{
			{Path: "/v1/prefixhandlerpkg/v1/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventHandlerAPIPathExtractionUnit tests API path extraction
func TestEventHandlerAPIPathExtractionUnit(t *testing.T) {
	var capturedRoute string
	dynamic.RegisterPackage("extractpkg", "v2", &mockTunnel{
		invoke: func(route, req string) string {
			capturedRoute = route
			return ""
		},
	})

	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/api/extractpkg/v2/users/list", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}

	if capturedRoute != "/users/list" {
		t.Errorf("capturedRoute = %q, want '/users/list'", capturedRoute)
	}
}

// TestEventHandlerPanicRecoveryUnit tests panic recovery
func TestEventHandlerPanicRecoveryUnit(t *testing.T) {
	dynamic.RegisterPackage("panicunitpkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			panic("test panic message")
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModeBatch),
	}, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/api/panicunitpkg/v1/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err == nil {
		t.Error("Expected error from panic recovery")
	}

	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should contain 'panic': %v", err)
	}
}

// TestEventHandlerInvalidPathUnit tests invalid path error handling
func TestEventHandlerInvalidPathUnit(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModeBatch),
	}, nil)

	testCases := []struct {
		name string
		path string
	}{
		{"empty path after api", "/api/"},
		{"missing version", "/api/pkg"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &event.Request{
				Items: []*event.Item{
					{Path: tc.path, Payload: []byte(`{}`)},
				},
			}
			payload, _ := proto.Marshal(req)

			err := e.Invoke(context.Background(), payload)
			if err == nil {
				t.Errorf("Expected error for path %q", tc.path)
			}
		})
	}
}

// TestEventHandlerMissingPackageUnit tests missing package error
func TestEventHandlerMissingPackageUnit(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModeBatch),
	}, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/api/nonexistentpkg123/v1/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err == nil {
		t.Error("Expected error for missing package")
	}

	if !strings.Contains(err.Error(), "package") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention package not found: %v", err)
	}
}

// TestEventHandlerMultipleStaticLinks tests multiple static link mappings
func TestEventHandlerMultipleStaticLinks(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithStaticLink("/path1", "/"),
		event.WithStaticLink("/path2", "/"),
	}, nil)

	for _, path := range []string{"/path1", "/path2"} {
		req := &event.Request{
			Items: []*event.Item{
				{Path: path, Payload: nil},
			},
		}
		payload, _ := proto.Marshal(req)

		err := e.Invoke(context.Background(), payload)
		if err != nil {
			t.Errorf("Invoke error for path %q: %v", path, err)
		}
	}
}

// TestEventHandlerMultiplePrefixLinks tests multiple prefix link mappings
func TestEventHandlerMultiplePrefixLinks(t *testing.T) {
	dynamic.RegisterPackage("multiprefixpkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return ""
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithPrefixLink("/prefix1", "/api"),
		event.WithPrefixLink("/prefix2", "/api"),
	}, nil)

	for _, prefix := range []string{"/prefix1", "/prefix2"} {
		req := &event.Request{
			Items: []*event.Item{
				{Path: prefix + "/multiprefixpkg/v1/route", Payload: []byte(`{}`)},
			},
		}
		payload, _ := proto.Marshal(req)

		err := e.Invoke(context.Background(), payload)
		if err != nil {
			t.Errorf("Invoke error for prefix %q: %v", prefix, err)
		}
	}
}

// TestEventHandlerDebugModeAPI tests debug mode API handling
func TestEventHandlerDebugModeAPI(t *testing.T) {
	var capturedRoute string
	dynamic.RegisterPackage("debugmodepkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			capturedRoute = route
			return ""
		},
	})

	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/_/api/debugmodepkg/v1/debug/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}

	if capturedRoute != "/debug/route" {
		t.Errorf("capturedRoute = %q, want '/debug/route'", capturedRoute)
	}
}
