package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"

	"github.com/aura-studio/lambda/event"
)

// --- Property-Based Tests for Task 4.2 ---

// TestProperty4_ContextSetGetRoundTrip_String verifies that for any key and string value,
// Set(key, value) followed by Get(key) returns (value, true) and GetString(key) returns value.
//
// **Validates: Requirements 5.3**
func TestProperty4_ContextSetGetRoundTrip_String(t *testing.T) {
	f := func(key, value string) bool {
		c := &event.Context{}
		c.Set(key, value)

		got, ok := c.Get(key)
		if !ok {
			return false
		}
		if got != value {
			return false
		}
		if c.GetString(key) != value {
			return false
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 4 (Context Set/Get string round-trip) failed: %v", err)
	}
}

// TestProperty4_ContextSetGetRoundTrip_Bool verifies that for any key and bool value,
// Set(key, value) followed by Get(key) returns (value, true) and GetBool(key) returns value.
//
// **Validates: Requirements 5.3**
func TestProperty4_ContextSetGetRoundTrip_Bool(t *testing.T) {
	f := func(key string, value bool) bool {
		c := &event.Context{}
		c.Set(key, value)

		got, ok := c.Get(key)
		if !ok {
			return false
		}
		if got != value {
			return false
		}
		if c.GetBool(key) != value {
			return false
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 4 (Context Set/Get bool round-trip) failed: %v", err)
	}
}

// TestProperty5_MatchPatternWildcard verifies that for any prefix and suffix strings,
// the pattern "<prefix>/*path" matches paths starting with "<prefix>/" and extracts
// the correct parameter. Paths not starting with "<prefix>/" should not match.
//
// **Validates: Requirements 5.5**
func TestProperty5_MatchPatternWildcard(t *testing.T) {
	// Generate safe path segments: non-empty, no slashes, no '*', printable ASCII
	genSegment := func(r *rand.Rand) string {
		const chars = "abcdefghijklmnopqrstuvwxyz0123456789-_"
		n := r.Intn(8) + 1
		b := make([]byte, n)
		for i := range b {
			b[i] = chars[r.Intn(len(chars))]
		}
		return string(b)
	}

	f := func(seed int64) bool {
		r := rand.New(rand.NewSource(seed))

		prefix := "/" + genSegment(r)
		suffix := genSegment(r)
		pattern := prefix + "/*path"
		path := prefix + "/" + suffix

		// Should match and extract "/<suffix>"
		param, matched := event.MatchPattern(pattern, path)
		if !matched {
			return false
		}
		if param != "/"+suffix {
			return false
		}

		// A path that does NOT start with prefix should not match
		otherPrefix := "/" + genSegment(r) + "x" // ensure different from prefix
		if strings.HasPrefix(otherPrefix+"/"+suffix, prefix+"/") {
			// In the unlikely case they collide, skip this check
			return true
		}
		_, matched2 := event.MatchPattern(pattern, otherPrefix+"/"+suffix)
		return !matched2
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 5 (MatchPattern wildcard) failed: %v", err)
	}
}

// TestProperty6_HandlerChainAbort verifies that when the Nth handler in a chain
// calls Abort(), subsequent handlers are not executed.
//
// **Validates: Requirements 5.6, 5.7**
func TestProperty6_HandlerChainAbort(t *testing.T) {
	// chainLen in [2, 10], abortAt in [0, chainLen-1)
	f := func(seed int64) bool {
		r := rand.New(rand.NewSource(seed))
		chainLen := r.Intn(9) + 2 // 2..10
		abortAt := r.Intn(chainLen - 1) // 0..chainLen-2, so there's always a handler after

		router := event.NewRouter()
		called := make([]bool, chainLen)

		handlers := make([]event.HandlerFunc, chainLen)
		for i := 0; i < chainLen; i++ {
			idx := i
			handlers[idx] = func(c *event.Context) {
				called[idx] = true
				if idx == abortAt {
					c.Abort()
				}
			}
		}

		router.Handle("/test", handlers...)

		ctx := &event.Context{}
		ctx.Set(event.ContextPath, "/test")
		router.Dispatch(ctx)

		// Handlers up to and including abortAt should have been called
		for i := 0; i <= abortAt; i++ {
			if !called[i] {
				return false
			}
		}
		// Handlers after abortAt should NOT have been called
		for i := abortAt + 1; i < chainLen; i++ {
			if called[i] {
				return false
			}
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 6 (handler chain abort) failed: %v", err)
	}
}

// TestProperty6_HandlerChainError verifies that when the Nth handler in a chain
// sets an error on the context, subsequent handlers are not executed.
//
// **Validates: Requirements 5.6, 5.7**
func TestProperty6_HandlerChainError(t *testing.T) {
	// chainLen in [2, 10], errorAt in [0, chainLen-1)
	f := func(seed int64) bool {
		r := rand.New(rand.NewSource(seed))
		chainLen := r.Intn(9) + 2 // 2..10
		errorAt := r.Intn(chainLen - 1) // 0..chainLen-2

		router := event.NewRouter()
		called := make([]bool, chainLen)

		handlers := make([]event.HandlerFunc, chainLen)
		for i := 0; i < chainLen; i++ {
			idx := i
			handlers[idx] = func(c *event.Context) {
				called[idx] = true
				if idx == errorAt {
					c.Set(event.ContextError, fmt.Errorf("error at handler %d", idx))
				}
			}
		}

		router.Handle("/test", handlers...)

		ctx := &event.Context{}
		ctx.Set(event.ContextPath, "/test")
		router.Dispatch(ctx)

		// Handlers up to and including errorAt should have been called
		for i := 0; i <= errorAt; i++ {
			if !called[i] {
				return false
			}
		}
		// Handlers after errorAt should NOT have been called
		for i := errorAt + 1; i < chainLen; i++ {
			if called[i] {
				return false
			}
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 6 (handler chain error) failed: %v", err)
	}
}

// --- Unit Tests for Task 4.3 ---

// newEventContext creates a new event.Context with the given path set.
func newEventContext(path string) *event.Context {
	c := &event.Context{}
	c.Set(event.ContextPath, path)
	return c
}

// TestEventRouterDispatch tests route registration and dispatch.
//
// **Validates: Requirements 5.4, 5.8**
func TestEventRouterDispatch(t *testing.T) {
	t.Run("dispatch to registered route", func(t *testing.T) {
		r := event.NewRouter()
		called := false
		r.Handle("/test", func(c *event.Context) {
			called = true
			c.Set("response", "test-response")
		})

		ctx := newEventContext("/test")
		r.Dispatch(ctx)

		if !called {
			t.Error("Handler was not called")
		}
		if ctx.GetString("response") != "test-response" {
			t.Errorf("Response = %q, want %q", ctx.GetString("response"), "test-response")
		}
	})

	t.Run("dispatch to NoRoute handler", func(t *testing.T) {
		r := event.NewRouter()
		noRouteCalled := false
		r.NoRoute(func(c *event.Context) {
			noRouteCalled = true
		})

		ctx := newEventContext("/nonexistent")
		r.Dispatch(ctx)

		if !noRouteCalled {
			t.Error("NoRoute handler was not called")
		}
	})

	t.Run("no route and no NoRoute handler sets error", func(t *testing.T) {
		r := event.NewRouter()

		ctx := newEventContext("/nonexistent")
		r.Dispatch(ctx)

		err := ctx.GetError()
		if err == nil {
			t.Fatal("Expected error to be set in context, got nil")
		}
		if !strings.Contains(err.Error(), "no route") {
			t.Errorf("Error = %q, want it to contain %q", err.Error(), "no route")
		}
	})

	t.Run("exact match", func(t *testing.T) {
		r := event.NewRouter()
		var matchedPath string
		r.Handle("/health-check", func(c *event.Context) {
			matchedPath = "health-check"
		})
		r.Handle("/other", func(c *event.Context) {
			matchedPath = "other"
		})

		ctx := newEventContext("/health-check")
		r.Dispatch(ctx)

		if matchedPath != "health-check" {
			t.Errorf("matchedPath = %q, want %q", matchedPath, "health-check")
		}
	})

	t.Run("wildcard match extracts param", func(t *testing.T) {
		r := event.NewRouter()
		var capturedPath string
		r.Handle("/api/*path", func(c *event.Context) {
			capturedPath = c.GetString(event.ContextPath)
		})

		ctx := newEventContext("/api/pkg/v1/route")
		r.Dispatch(ctx)

		if capturedPath != "/pkg/v1/route" {
			t.Errorf("capturedPath = %q, want %q", capturedPath, "/pkg/v1/route")
		}
	})
}
