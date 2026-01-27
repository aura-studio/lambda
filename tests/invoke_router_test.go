package tests

import (
	"testing"

	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: invoke-lambda-handler, Property 2: 路由匹配正确性**
// **Validates: Requirements 2.1, 2.2, 2.4**
//
// Property 2: 路由匹配正确性
// For any 注册的路由模式和请求路径，如果路径与模式匹配（精确匹配或通配符匹配），
// 则路由器 SHALL 调用对应的处理器；如果路径不匹配任何模式，则路由器 SHALL 调用 NoRoute 处理器。

// genPathSegment generates a valid path segment (alphanumeric, non-empty)
// Uses Identifier which always generates non-empty strings
func genPathSegment() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		// Limit length to avoid overly long paths
		if len(s) > 15 {
			return s[:15]
		}
		return s
	})
}

// genSimplePath generates a simple path like "/foo"
func genSimplePath() gopter.Gen {
	return genPathSegment().Map(func(segment string) string {
		return "/" + segment
	})
}

// genTwoSegmentPath generates a path like "/foo/bar"
func genTwoSegmentPath() gopter.Gen {
	return gopter.CombineGens(
		genPathSegment(),
		genPathSegment(),
	).Map(func(values []interface{}) string {
		return "/" + values[0].(string) + "/" + values[1].(string)
	})
}

// TestExactPathMatching tests that exact path patterns match correctly
// **Validates: Requirements 2.1**
func TestExactPathMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Exact path matching: pattern matches identical path", prop.ForAll(
		func(path string) bool {
			// Pattern equals path should match
			param, ok := invoke.MatchPattern(path, path)
			if !ok {
				t.Logf("Exact match failed: pattern=%q, path=%q", path, path)
				return false
			}
			// For exact matches, param should be empty
			if param != "" {
				t.Logf("Exact match should have empty param: got %q", param)
				return false
			}
			return true
		},
		genSimplePath(),
	))

	properties.TestingRun(t)
}

// TestExactPathNonMatching tests that exact path patterns don't match different paths
// **Validates: Requirements 2.1**
func TestExactPathNonMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Exact path non-matching: pattern does not match different path", prop.ForAll(
		func(segment1, segment2 string) bool {
			// Skip if they happen to be equal
			if segment1 == segment2 {
				return true
			}
			pattern := "/" + segment1
			path := "/" + segment2
			// Different paths should not match
			_, ok := invoke.MatchPattern(pattern, path)
			if ok {
				t.Logf("Unexpected match: pattern=%q, path=%q", pattern, path)
				return false
			}
			return true
		},
		genPathSegment(),
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestWildcardPathMatching tests that wildcard patterns match paths with the correct prefix
// **Validates: Requirements 2.2**
func TestWildcardPathMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Wildcard path matching: pattern /prefix/*path matches /prefix/rest", prop.ForAll(
		func(prefix string, rest string) bool {
			pattern := "/" + prefix + "/*path"
			path := "/" + prefix + "/" + rest

			param, ok := invoke.MatchPattern(pattern, path)
			if !ok {
				t.Logf("Wildcard match failed: pattern=%q, path=%q", pattern, path)
				return false
			}

			// Param should be "/" + rest (with leading slash)
			expectedParam := "/" + rest
			if param != expectedParam {
				t.Logf("Wildcard param mismatch: expected %q, got %q", expectedParam, param)
				return false
			}
			return true
		},
		genPathSegment(),
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestWildcardPathNonMatching tests that wildcard patterns don't match paths with different prefixes
// **Validates: Requirements 2.2**
func TestWildcardPathNonMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Wildcard path non-matching: pattern /prefix/*path does not match /other/rest", prop.ForAll(
		func(prefix, other, rest string) bool {
			// Skip if prefixes happen to be equal
			if prefix == other {
				return true
			}
			pattern := "/" + prefix + "/*path"
			path := "/" + other + "/" + rest

			_, ok := invoke.MatchPattern(pattern, path)
			if ok {
				t.Logf("Unexpected wildcard match: pattern=%q, path=%q", pattern, path)
				return false
			}
			return true
		},
		genPathSegment(),
		genPathSegment(),
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestWildcardEmptyRest tests that wildcard patterns match paths with empty rest (just the prefix)
// **Validates: Requirements 2.2**
func TestWildcardEmptyRest(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Wildcard empty rest: pattern /prefix/*path matches /prefix/ with param /", prop.ForAll(
		func(prefix string) bool {
			pattern := "/" + prefix + "/*path"
			path := "/" + prefix + "/"

			param, ok := invoke.MatchPattern(pattern, path)
			if !ok {
				t.Logf("Wildcard empty rest match failed: pattern=%q, path=%q", pattern, path)
				return false
			}

			// Param should be "/" for empty rest
			if param != "/" {
				t.Logf("Wildcard empty rest param mismatch: expected %q, got %q", "/", param)
				return false
			}
			return true
		},
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestRouterDispatchMatching tests that router dispatches to correct handler when path matches
// **Validates: Requirements 2.1, 2.4**
func TestRouterDispatchMatching(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Router dispatch: calls matching handler when path matches registered pattern", prop.ForAll(
		func(segment string) bool {
			pattern := "/" + segment
			requestPath := "/" + segment

			// Create a new router
			router := invoke.NewRouter()

			// Track which handler was called
			handlerCalled := false
			noRouteCalled := false

			// Register handler for the pattern
			router.Handle(pattern, func(c *invoke.Context) {
				handlerCalled = true
			})

			// Register NoRoute handler
			router.NoRoute(func(c *invoke.Context) {
				noRouteCalled = true
			})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: requestPath,
			}
			router.Dispatch(ctx)

			// Handler should have been called
			if !handlerCalled {
				t.Logf("Handler should be called for matching path: pattern=%q, path=%q", pattern, requestPath)
				return false
			}
			if noRouteCalled {
				t.Logf("NoRoute should not be called when path matches")
				return false
			}

			return true
		},
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestRouterNoRouteHandler tests that NoRoute handler is called when no pattern matches
// **Validates: Requirements 2.4**
func TestRouterNoRouteHandler(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Router NoRoute: NoRoute handler is called when path matches no pattern", prop.ForAll(
		func(patternSegment, pathSegment string) bool {
			// Skip if they happen to match
			if patternSegment == pathSegment {
				return true
			}

			pattern := "/" + patternSegment
			requestPath := "/" + pathSegment

			// Create a new router
			router := invoke.NewRouter()

			// Track which handler was called
			handlerCalled := false
			noRouteCalled := false

			// Register a handler for the pattern
			router.Handle(pattern, func(c *invoke.Context) {
				handlerCalled = true
			})

			// Register NoRoute handler
			router.NoRoute(func(c *invoke.Context) {
				noRouteCalled = true
			})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: requestPath,
			}
			router.Dispatch(ctx)

			// NoRoute should be called, handler should not
			if handlerCalled {
				t.Logf("Handler should not be called for non-matching path: pattern=%q, path=%q", pattern, requestPath)
				return false
			}
			if !noRouteCalled {
				t.Logf("NoRoute should be called for non-matching path: pattern=%q, path=%q", pattern, requestPath)
				return false
			}

			return true
		},
		genPathSegment(),
		genPathSegment(),
	))

	properties.TestingRun(t)
}

// TestWildcardRouterDispatch tests router dispatch with wildcard patterns
// **Validates: Requirements 2.2, 2.4**
func TestWildcardRouterDispatch(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Wildcard router dispatch: calls handler and sets ParamPath correctly", prop.ForAll(
		func(prefix, rest string) bool {
			pattern := "/" + prefix + "/*path"
			requestPath := "/" + prefix + "/" + rest

			// Create a new router
			router := invoke.NewRouter()

			// Track handler call and param
			handlerCalled := false
			capturedParam := ""

			// Register wildcard handler
			router.Handle(pattern, func(c *invoke.Context) {
				handlerCalled = true
				capturedParam = c.ParamPath
			})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: requestPath,
			}
			router.Dispatch(ctx)

			// Handler should be called
			if !handlerCalled {
				t.Logf("Handler should be called for matching wildcard: pattern=%q, path=%q", pattern, requestPath)
				return false
			}

			// ParamPath should be set correctly
			expectedParam := "/" + rest
			if capturedParam != expectedParam {
				t.Logf("ParamPath mismatch: expected %q, got %q", expectedParam, capturedParam)
				return false
			}

			return true
		},
		genPathSegment(),
		genPathSegment(),
	))

	properties.TestingRun(t)
}


// **Feature: invoke-lambda-handler, Property 4: 中间件链执行顺序**
// **Validates: Requirements 2.3**
//
// Property 4: 中间件链执行顺序
// For any 注册的中间件链，路由器 SHALL 按注册顺序依次执行每个中间件，
// 直到某个中间件调用 Abort 或所有中间件执行完成。

// TestMiddlewareChainExecutionOrder tests that middleware executes in registration order
// **Validates: Requirements 2.3**
func TestMiddlewareChainExecutionOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Middleware chain: executes in registration order", prop.ForAll(
		func(numMiddleware int) bool {
			// Create a new router
			router := invoke.NewRouter()

			// Track execution order
			executionOrder := make([]int, 0, numMiddleware)

			// Register middleware in order 0, 1, 2, ...
			for i := 0; i < numMiddleware; i++ {
				idx := i // Capture loop variable
				router.Use(func(c *invoke.Context) {
					executionOrder = append(executionOrder, idx)
				})
			}

			// Register a handler for the path
			router.Handle("/test", func(c *invoke.Context) {})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: "/test",
			}
			router.Dispatch(ctx)

			// Verify all middleware executed
			if len(executionOrder) != numMiddleware {
				t.Logf("Expected %d middleware executions, got %d", numMiddleware, len(executionOrder))
				return false
			}

			// Verify execution order is 0, 1, 2, ...
			for i, order := range executionOrder {
				if order != i {
					t.Logf("Middleware execution order mismatch at position %d: expected %d, got %d", i, i, order)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10), // Test with 1-10 middleware
	))

	properties.TestingRun(t)
}

// TestMiddlewareChainAbortStopsExecution tests that Abort() stops the middleware chain
// **Validates: Requirements 2.3**
func TestMiddlewareChainAbortStopsExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Middleware chain abort: Abort() stops execution at the aborting middleware", prop.ForAll(
		func(totalMiddleware, abortAt int) bool {
			// Ensure abortAt is within valid range
			if abortAt >= totalMiddleware {
				abortAt = totalMiddleware - 1
			}

			// Create a new router
			router := invoke.NewRouter()

			// Track execution order
			executionOrder := make([]int, 0, totalMiddleware)

			// Register middleware, one of which will abort
			for i := 0; i < totalMiddleware; i++ {
				idx := i // Capture loop variable
				router.Use(func(c *invoke.Context) {
					executionOrder = append(executionOrder, idx)
					if idx == abortAt {
						c.Abort()
					}
				})
			}

			// Register a handler for the path
			handlerCalled := false
			router.Handle("/test", func(c *invoke.Context) {
				handlerCalled = true
			})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: "/test",
			}
			router.Dispatch(ctx)

			// Verify only middleware up to and including abortAt executed
			expectedExecutions := abortAt + 1
			if len(executionOrder) != expectedExecutions {
				t.Logf("Expected %d middleware executions (abort at %d), got %d",
					expectedExecutions, abortAt, len(executionOrder))
				return false
			}

			// Verify execution order is 0, 1, 2, ..., abortAt
			for i, order := range executionOrder {
				if order != i {
					t.Logf("Middleware execution order mismatch at position %d: expected %d, got %d", i, i, order)
					return false
				}
			}

			// Verify handler was NOT called (abort happened in middleware)
			if handlerCalled {
				t.Logf("Handler should not be called when middleware aborts")
				return false
			}

			return true
		},
		gen.IntRange(2, 10), // Total middleware (at least 2 to test abort)
		gen.IntRange(0, 9),  // Abort position (will be clamped to valid range)
	))

	properties.TestingRun(t)
}

// TestMiddlewareChainWithHandlers tests that middleware executes before route handlers
// **Validates: Requirements 2.3**
func TestMiddlewareChainWithHandlers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Middleware chain with handlers: middleware executes before route handlers", prop.ForAll(
		func(numMiddleware int) bool {
			// Create a new router
			router := invoke.NewRouter()

			// Track execution order with markers
			// Middleware will add negative numbers (-1, -2, ...), handler adds positive (1)
			executionOrder := make([]int, 0, numMiddleware+1)

			// Register middleware
			for i := 0; i < numMiddleware; i++ {
				idx := i + 1 // Use 1-based index for middleware markers
				router.Use(func(c *invoke.Context) {
					executionOrder = append(executionOrder, -idx)
				})
			}

			// Register a handler for the path
			router.Handle("/test", func(c *invoke.Context) {
				executionOrder = append(executionOrder, 1) // Handler marker
			})

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: "/test",
			}
			router.Dispatch(ctx)

			// Verify total executions (middleware + handler)
			expectedTotal := numMiddleware + 1
			if len(executionOrder) != expectedTotal {
				t.Logf("Expected %d total executions, got %d", expectedTotal, len(executionOrder))
				return false
			}

			// Verify middleware executed first (negative markers)
			for i := 0; i < numMiddleware; i++ {
				expectedMarker := -(i + 1)
				if executionOrder[i] != expectedMarker {
					t.Logf("Middleware marker mismatch at position %d: expected %d, got %d",
						i, expectedMarker, executionOrder[i])
					return false
				}
			}

			// Verify handler executed last (positive marker)
			if executionOrder[numMiddleware] != 1 {
				t.Logf("Handler should execute after all middleware, got marker %d at position %d",
					executionOrder[numMiddleware], numMiddleware)
				return false
			}

			return true
		},
		gen.IntRange(1, 10), // Test with 1-10 middleware
	))

	properties.TestingRun(t)
}

// TestMiddlewareChainMultipleHandlers tests middleware with multiple route handlers
// **Validates: Requirements 2.3**
func TestMiddlewareChainMultipleHandlers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Middleware chain multiple handlers: route handlers execute in order after middleware", prop.ForAll(
		func(numMiddleware, numHandlers int) bool {
			// Create a new router
			router := invoke.NewRouter()

			// Track execution order
			// Middleware: negative (-1, -2, ...)
			// Handlers: positive (1, 2, ...)
			executionOrder := make([]int, 0, numMiddleware+numHandlers)

			// Register middleware
			for i := 0; i < numMiddleware; i++ {
				idx := i + 1
				router.Use(func(c *invoke.Context) {
					executionOrder = append(executionOrder, -idx)
				})
			}

			// Create handlers slice
			handlers := make([]invoke.HandlerFunc, numHandlers)
			for i := 0; i < numHandlers; i++ {
				idx := i + 1
				handlers[i] = func(c *invoke.Context) {
					executionOrder = append(executionOrder, idx)
				}
			}

			// Register handlers for the path
			router.Handle("/test", handlers...)

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: "/test",
			}
			router.Dispatch(ctx)

			// Verify total executions
			expectedTotal := numMiddleware + numHandlers
			if len(executionOrder) != expectedTotal {
				t.Logf("Expected %d total executions, got %d", expectedTotal, len(executionOrder))
				return false
			}

			// Verify middleware executed first in order
			for i := 0; i < numMiddleware; i++ {
				expectedMarker := -(i + 1)
				if executionOrder[i] != expectedMarker {
					t.Logf("Middleware marker mismatch at position %d: expected %d, got %d",
						i, expectedMarker, executionOrder[i])
					return false
				}
			}

			// Verify handlers executed after middleware in order
			for i := 0; i < numHandlers; i++ {
				expectedMarker := i + 1
				actualPos := numMiddleware + i
				if executionOrder[actualPos] != expectedMarker {
					t.Logf("Handler marker mismatch at position %d: expected %d, got %d",
						actualPos, expectedMarker, executionOrder[actualPos])
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 5), // Test with 1-5 middleware
		gen.IntRange(1, 5), // Test with 1-5 handlers
	))

	properties.TestingRun(t)
}

// TestMiddlewareChainAbortInHandler tests that Abort() in a handler stops subsequent handlers
// **Validates: Requirements 2.3**
func TestMiddlewareChainAbortInHandler(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Handler chain abort: Abort() in handler stops subsequent handlers", prop.ForAll(
		func(totalHandlers, abortAt int) bool {
			// Ensure abortAt is within valid range
			if abortAt >= totalHandlers {
				abortAt = totalHandlers - 1
			}

			// Create a new router
			router := invoke.NewRouter()

			// Track execution order
			executionOrder := make([]int, 0, totalHandlers)

			// Create handlers, one of which will abort
			handlers := make([]invoke.HandlerFunc, totalHandlers)
			for i := 0; i < totalHandlers; i++ {
				idx := i
				handlers[i] = func(c *invoke.Context) {
					executionOrder = append(executionOrder, idx)
					if idx == abortAt {
						c.Abort()
					}
				}
			}

			// Register handlers for the path
			router.Handle("/test", handlers...)

			// Create context and dispatch
			ctx := &invoke.Context{
				Path: "/test",
			}
			router.Dispatch(ctx)

			// Verify only handlers up to and including abortAt executed
			expectedExecutions := abortAt + 1
			if len(executionOrder) != expectedExecutions {
				t.Logf("Expected %d handler executions (abort at %d), got %d",
					expectedExecutions, abortAt, len(executionOrder))
				return false
			}

			// Verify execution order is 0, 1, 2, ..., abortAt
			for i, order := range executionOrder {
				if order != i {
					t.Logf("Handler execution order mismatch at position %d: expected %d, got %d", i, i, order)
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 10), // Total handlers (at least 2 to test abort)
		gen.IntRange(0, 9),  // Abort position (will be clamped to valid range)
	))

	properties.TestingRun(t)
}
