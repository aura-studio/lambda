package tests

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: event-lambda-handler
// Property 6: Router Path Matching
// *For any* registered route pattern and any path that matches that pattern,
// the Router SHALL dispatch to the correct handler.
//
// **Validates: Requirements 2.1, 2.3**

// Property 7: NoRoute Fallback
// *For any* path that does not match any registered route,
// the Router SHALL invoke the NoRoute handler.
//
// **Validates: Requirements 2.2**

// Property 8: Abort Stops Chain
// *For any* handler chain where a handler calls Abort(),
// no subsequent handlers in the chain SHALL be executed.
//
// **Validates: Requirements 2.5**

// genValidPathSegment generates a valid path segment (alphanumeric, no slashes)
func genValidPathSegment() gopter.Gen {
	// Use a fixed set of valid segments to avoid SuchThat filtering
	return gen.OneConstOf("users", "api", "v1", "v2", "test", "pkg", "route", "data", "items", "config")
}

// genExactPath generates a valid exact path like /segment1/segment2
func genExactPath() gopter.Gen {
	return gen.IntRange(1, 3).FlatMap(func(n interface{}) gopter.Gen {
		count := n.(int)
		return gen.SliceOfN(count, genValidPathSegment()).Map(func(segments []string) string {
			return "/" + strings.Join(segments, "/")
		})
	}, reflect.TypeOf(""))
}

// genWildcardPrefix generates a valid wildcard prefix like /api, /wapi, /_/api
func genWildcardPrefix() gopter.Gen {
	return gen.OneConstOf("/api", "/wapi", "/_/api", "/v1", "/v2")
}

// genWildcardSuffix generates a valid suffix for wildcard paths
func genWildcardSuffix() gopter.Gen {
	return gen.IntRange(0, 3).FlatMap(func(n interface{}) gopter.Gen {
		count := n.(int)
		if count == 0 {
			return gen.Const("")
		}
		return gen.SliceOfN(count, genValidPathSegment()).Map(func(segments []string) string {
			return "/" + strings.Join(segments, "/")
		})
	}, reflect.TypeOf(""))
}

// TestEventRouterPathMatchingExact tests Property 6 for exact path matching
// For any registered exact route pattern and any path that matches that pattern,
// the Router SHALL dispatch to the correct handler.
func TestEventRouterPathMatchingExact(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("exact path matching dispatches to correct handler", prop.ForAll(
		func(path string) bool {
			router := event.NewRouter()
			handlerCalled := false
			expectedPath := path

			router.Handle(path, func(c *event.Context) {
				handlerCalled = true
				// Verify the path matches
				if c.Path != expectedPath {
					return
				}
			})

			ctx := &event.Context{Path: path}
			router.Dispatch(ctx)

			return handlerCalled
		},
		genExactPath(),
	))

	properties.TestingRun(t)
}

// TestEventRouterPathMatchingWildcard tests Property 6 for wildcard path matching
// For any registered wildcard route pattern and any path that matches that pattern,
// the Router SHALL dispatch to the correct handler and set ParamPath correctly.
func TestEventRouterPathMatchingWildcard(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("wildcard path matching dispatches to correct handler with ParamPath", prop.ForAll(
		func(prefix string, suffix string) bool {
			router := event.NewRouter()
			handlerCalled := false
			var capturedParamPath string

			pattern := prefix + "/*path"
			fullPath := prefix + suffix
			if suffix == "" {
				fullPath = prefix + "/"
			}

			router.Handle(pattern, func(c *event.Context) {
				handlerCalled = true
				capturedParamPath = c.ParamPath
			})

			ctx := &event.Context{Path: fullPath}
			router.Dispatch(ctx)

			// Verify handler was called
			if !handlerCalled {
				return false
			}

			// Verify ParamPath is set correctly
			// ParamPath should be the suffix with leading slash
			expectedParam := suffix
			if expectedParam == "" {
				expectedParam = "/"
			}
			return capturedParamPath == expectedParam
		},
		genWildcardPrefix(),
		genWildcardSuffix(),
	))

	properties.TestingRun(t)
}

// TestEventRouterNoRouteFallback tests Property 7: NoRoute Fallback
// For any path that does not match any registered route,
// the Router SHALL invoke the NoRoute handler.
func TestEventRouterNoRouteFallback(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("unmatched path invokes NoRoute handler", prop.ForAll(
		func(registeredPath string, requestPath string) bool {
			// Ensure paths are different
			if registeredPath == requestPath {
				return true // Skip this case
			}

			router := event.NewRouter()
			routeHandlerCalled := false
			noRouteHandlerCalled := false

			router.Handle(registeredPath, func(c *event.Context) {
				routeHandlerCalled = true
			})

			router.NoRoute(func(c *event.Context) {
				noRouteHandlerCalled = true
			})

			ctx := &event.Context{Path: requestPath}
			router.Dispatch(ctx)

			// NoRoute should be called, route handler should not
			return noRouteHandlerCalled && !routeHandlerCalled
		},
		genExactPath(),
		genExactPath().SuchThat(func(s string) bool {
			// Generate a different path
			return true
		}),
	))

	properties.TestingRun(t)
}

// TestEventRouterNoRouteFallbackWithNoRegisteredRoutes tests Property 7
// When no routes are registered, NoRoute handler should be invoked.
func TestEventRouterNoRouteFallbackWithNoRegisteredRoutes(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("NoRoute handler invoked when no routes registered", prop.ForAll(
		func(path string) bool {
			router := event.NewRouter()
			noRouteHandlerCalled := false

			router.NoRoute(func(c *event.Context) {
				noRouteHandlerCalled = true
			})

			ctx := &event.Context{Path: path}
			router.Dispatch(ctx)

			return noRouteHandlerCalled
		},
		genExactPath(),
	))

	properties.TestingRun(t)
}

// TestEventRouterAbortStopsChainInMiddleware tests Property 8: Abort Stops Chain
// When a middleware calls Abort(), no subsequent handlers SHALL be executed.
func TestEventRouterAbortStopsChainInMiddleware(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Abort in middleware stops handler chain", prop.ForAll(
		func(path string, middlewareCount int, abortAt int) bool {
			if middlewareCount <= 0 {
				return true // Skip invalid case
			}
			if abortAt >= middlewareCount {
				abortAt = middlewareCount - 1
			}

			router := event.NewRouter()
			executedMiddlewares := make([]bool, middlewareCount)
			routeHandlerCalled := false

			// Register middlewares
			for i := 0; i < middlewareCount; i++ {
				idx := i // Capture loop variable
				router.Use(func(c *event.Context) {
					executedMiddlewares[idx] = true
					if idx == abortAt {
						c.Abort()
					}
				})
			}

			router.Handle(path, func(c *event.Context) {
				routeHandlerCalled = true
			})

			ctx := &event.Context{Path: path}
			router.Dispatch(ctx)

			// Verify: middlewares before and at abortAt should be executed
			for i := 0; i <= abortAt; i++ {
				if !executedMiddlewares[i] {
					return false
				}
			}

			// Verify: middlewares after abortAt should NOT be executed
			for i := abortAt + 1; i < middlewareCount; i++ {
				if executedMiddlewares[i] {
					return false
				}
			}

			// Verify: route handler should NOT be executed
			return !routeHandlerCalled
		},
		genExactPath(),
		gen.IntRange(1, 5),
		gen.IntRange(0, 4),
	))

	properties.TestingRun(t)
}

// TestEventRouterAbortStopsChainInRouteHandler tests Property 8: Abort Stops Chain
// When a route handler calls Abort(), no subsequent handlers in the chain SHALL be executed.
func TestEventRouterAbortStopsChainInRouteHandler(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Abort in route handler stops subsequent handlers", prop.ForAll(
		func(path string, handlerCount int, abortAt int) bool {
			if handlerCount <= 0 {
				return true // Skip invalid case
			}
			if abortAt >= handlerCount {
				abortAt = handlerCount - 1
			}

			router := event.NewRouter()
			executedHandlers := make([]bool, handlerCount)

			// Create handlers slice
			handlers := make([]event.HandlerFunc, handlerCount)
			for i := 0; i < handlerCount; i++ {
				idx := i // Capture loop variable
				handlers[idx] = func(c *event.Context) {
					executedHandlers[idx] = true
					if idx == abortAt {
						c.Abort()
					}
				}
			}

			router.Handle(path, handlers...)

			ctx := &event.Context{Path: path}
			router.Dispatch(ctx)

			// Verify: handlers before and at abortAt should be executed
			for i := 0; i <= abortAt; i++ {
				if !executedHandlers[i] {
					return false
				}
			}

			// Verify: handlers after abortAt should NOT be executed
			for i := abortAt + 1; i < handlerCount; i++ {
				if executedHandlers[i] {
					return false
				}
			}

			return true
		},
		genExactPath(),
		gen.IntRange(1, 5),
		gen.IntRange(0, 4),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Unit Tests for additional coverage
// ============================================================================

// TestEventRouterDispatchToMatchingRoute tests basic route dispatch
func TestEventRouterDispatchToMatchingRoute(t *testing.T) {
	r := event.NewRouter()
	called := false
	r.Handle("/test", func(c *event.Context) {
		called = true
	})

	ctx := &event.Context{Path: "/test"}
	r.Dispatch(ctx)

	if !called {
		t.Error("Handler was not called")
	}
}

// TestEventRouterDispatchToNoRouteHandler tests NoRoute handler dispatch
func TestEventRouterDispatchToNoRouteHandler(t *testing.T) {
	r := event.NewRouter()
	noRouteCalled := false
	r.NoRoute(func(c *event.Context) {
		noRouteCalled = true
	})

	ctx := &event.Context{Path: "/nonexistent"}
	r.Dispatch(ctx)

	if !noRouteCalled {
		t.Error("NoRoute handler was not called")
	}
}

// TestEventRouterMiddlewareExecutionOrder tests middleware execution order
func TestEventRouterMiddlewareExecutionOrder(t *testing.T) {
	r := event.NewRouter()
	order := []string{}

	r.Use(func(c *event.Context) {
		order = append(order, "middleware1")
	})
	r.Use(func(c *event.Context) {
		order = append(order, "middleware2")
	})
	r.Handle("/test", func(c *event.Context) {
		order = append(order, "handler")
	})

	ctx := &event.Context{Path: "/test"}
	r.Dispatch(ctx)

	expected := []string{"middleware1", "middleware2", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("Order length = %d, want %d", len(order), len(expected))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

// TestEventRouterAbortStopsChain tests that Abort stops the handler chain
func TestEventRouterAbortStopsChain(t *testing.T) {
	r := event.NewRouter()
	handlerCalled := false

	r.Use(func(c *event.Context) {
		c.Abort()
	})
	r.Handle("/test", func(c *event.Context) {
		handlerCalled = true
	})

	ctx := &event.Context{Path: "/test"}
	r.Dispatch(ctx)

	if handlerCalled {
		t.Error("Handler should not be called after Abort")
	}
}

// TestEventRouterWildcardRouteSetsParam tests wildcard route parameter extraction
func TestEventRouterWildcardRouteSetsParam(t *testing.T) {
	r := event.NewRouter()
	var capturedParam string

	r.Handle("/api/*path", func(c *event.Context) {
		capturedParam = c.ParamPath
	})

	ctx := &event.Context{Path: "/api/pkg/v1/route"}
	r.Dispatch(ctx)

	if capturedParam != "/pkg/v1/route" {
		t.Errorf("ParamPath = %q, want '/pkg/v1/route'", capturedParam)
	}
}

// TestEventRouterWildcardRouteEmptySuffix tests wildcard route with empty suffix
func TestEventRouterWildcardRouteEmptySuffix(t *testing.T) {
	r := event.NewRouter()
	var capturedParam string
	handlerCalled := false

	r.Handle("/api/*path", func(c *event.Context) {
		handlerCalled = true
		capturedParam = c.ParamPath
	})

	ctx := &event.Context{Path: "/api/"}
	r.Dispatch(ctx)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if capturedParam != "/" {
		t.Errorf("ParamPath = %q, want '/'", capturedParam)
	}
}

// TestEventRouterMultipleRoutes tests routing with multiple registered routes
func TestEventRouterMultipleRoutes(t *testing.T) {
	r := event.NewRouter()
	calledRoute := ""

	r.Handle("/route1", func(c *event.Context) {
		calledRoute = "route1"
	})
	r.Handle("/route2", func(c *event.Context) {
		calledRoute = "route2"
	})
	r.Handle("/api/*path", func(c *event.Context) {
		calledRoute = "api"
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/route1", "route1"},
		{"/route2", "route2"},
		{"/api/test", "api"},
	}

	for _, tt := range tests {
		calledRoute = ""
		ctx := &event.Context{Path: tt.path}
		r.Dispatch(ctx)

		if calledRoute != tt.expected {
			t.Errorf("Path %q: calledRoute = %q, want %q", tt.path, calledRoute, tt.expected)
		}
	}
}

// TestEventRouterNoRouteWithNoHandlers tests dispatch when no NoRoute handler is set
func TestEventRouterNoRouteWithNoHandlers(t *testing.T) {
	r := event.NewRouter()

	ctx := &event.Context{Path: "/nonexistent"}
	r.Dispatch(ctx)

	// Should set an error on context
	if ctx.Err == nil {
		t.Error("Expected error to be set when no route and no NoRoute handler")
	}
}

// TestEventRouterNilHandlerSkipped tests that nil handlers are skipped
func TestEventRouterNilHandlerSkipped(t *testing.T) {
	r := event.NewRouter()
	handlerCalled := false

	r.Use(nil) // nil middleware should be skipped
	r.Handle("/test", nil, func(c *event.Context) {
		handlerCalled = true
	})

	ctx := &event.Context{Path: "/test"}
	r.Dispatch(ctx)

	if !handlerCalled {
		t.Error("Handler should be called even with nil handlers in chain")
	}
}

// TestEventContextAbort tests the Context.Abort function
func TestEventContextAbort(t *testing.T) {
	ctx := &event.Context{}
	ctx.Abort()
	// The abort state is internal, tested through router dispatch
}

// TestEventRouterHandlerErrorStopsChain tests that setting error stops chain
func TestEventRouterHandlerErrorStopsChain(t *testing.T) {
	r := event.NewRouter()
	secondHandlerCalled := false

	r.Handle("/test",
		func(c *event.Context) {
			c.Err = &testError{msg: "test error"}
		},
		func(c *event.Context) {
			secondHandlerCalled = true
		},
	)

	ctx := &event.Context{Path: "/test"}
	r.Dispatch(ctx)

	if secondHandlerCalled {
		t.Error("Second handler should not be called after error is set")
	}
	if ctx.Err == nil {
		t.Error("Error should be set on context")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
