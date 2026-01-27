package tests

import (
	"testing"

	"github.com/aura-studio/lambda/reqresp"
)

// TestMatchPattern tests the MatchPattern function
func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		path        string
		wantParam   string
		wantMatched bool
	}{
		{
			name:        "exact match root",
			pattern:     "/",
			path:        "/",
			wantParam:   "",
			wantMatched: true,
		},
		{
			name:        "exact match health-check",
			pattern:     "/health-check",
			path:        "/health-check",
			wantParam:   "",
			wantMatched: true,
		},
		{
			name:        "exact match no match",
			pattern:     "/health-check",
			path:        "/other",
			wantParam:   "",
			wantMatched: false,
		},
		{
			name:        "wildcard api with route",
			pattern:     "/api/*path",
			path:        "/api/pkg/v1/route",
			wantParam:   "/pkg/v1/route",
			wantMatched: true,
		},
		{
			name:        "wildcard api no route",
			pattern:     "/api/*path",
			path:        "/api/",
			wantParam:   "/",
			wantMatched: true,
		},
		{
			name:        "wildcard wapi with route",
			pattern:     "/wapi/*path",
			path:        "/wapi/pkg/v1/route",
			wantParam:   "/pkg/v1/route",
			wantMatched: true,
		},
		{
			name:        "wildcard debug api",
			pattern:     "/_/api/*path",
			path:        "/_/api/pkg/v1/route",
			wantParam:   "/pkg/v1/route",
			wantMatched: true,
		},
		{
			name:        "wildcard no match different prefix",
			pattern:     "/api/*path",
			path:        "/wapi/pkg/v1/route",
			wantParam:   "",
			wantMatched: false,
		},
		{
			name:        "wildcard deep path",
			pattern:     "/api/*path",
			path:        "/api/pkg/v1/users/123/profile/settings",
			wantParam:   "/pkg/v1/users/123/profile/settings",
			wantMatched: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParam, gotMatched := reqresp.MatchPattern(tt.pattern, tt.path)
			if gotMatched != tt.wantMatched {
				t.Errorf("MatchPattern() matched = %v, want %v", gotMatched, tt.wantMatched)
			}
			if gotParam != tt.wantParam {
				t.Errorf("MatchPattern() param = %q, want %q", gotParam, tt.wantParam)
			}
		})
	}
}

// TestRouterDispatch tests the Router.Dispatch function
func TestRouterDispatch(t *testing.T) {
	t.Run("dispatch to matching route", func(t *testing.T) {
		r := reqresp.NewRouter()
		called := false
		r.Handle("/test", func(c *reqresp.Context) {
			called = true
			c.Response = "test-response"
		})

		ctx := &reqresp.Context{Path: "/test"}
		r.Dispatch(ctx)

		if !called {
			t.Error("Handler was not called")
		}
		if ctx.Response != "test-response" {
			t.Errorf("Response = %q, want 'test-response'", ctx.Response)
		}
	})

	t.Run("dispatch to no route handler", func(t *testing.T) {
		r := reqresp.NewRouter()
		noRouteCalled := false
		r.NoRoute(func(c *reqresp.Context) {
			noRouteCalled = true
			c.Err = nil // Clear error
		})

		ctx := &reqresp.Context{Path: "/nonexistent"}
		r.Dispatch(ctx)

		if !noRouteCalled {
			t.Error("NoRoute handler was not called")
		}
	})

	t.Run("middleware execution order", func(t *testing.T) {
		r := reqresp.NewRouter()
		order := []string{}

		r.Use(func(c *reqresp.Context) {
			order = append(order, "middleware1")
		})
		r.Use(func(c *reqresp.Context) {
			order = append(order, "middleware2")
		})
		r.Handle("/test", func(c *reqresp.Context) {
			order = append(order, "handler")
		})

		ctx := &reqresp.Context{Path: "/test"}
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
	})

	t.Run("abort stops chain", func(t *testing.T) {
		r := reqresp.NewRouter()
		handlerCalled := false

		r.Use(func(c *reqresp.Context) {
			c.Abort()
		})
		r.Handle("/test", func(c *reqresp.Context) {
			handlerCalled = true
		})

		ctx := &reqresp.Context{Path: "/test"}
		r.Dispatch(ctx)

		if handlerCalled {
			t.Error("Handler should not be called after Abort")
		}
	})

	t.Run("wildcard route sets param", func(t *testing.T) {
		r := reqresp.NewRouter()
		var capturedParam string

		r.Handle("/api/*path", func(c *reqresp.Context) {
			capturedParam = c.ParamPath
		})

		ctx := &reqresp.Context{Path: "/api/pkg/v1/route"}
		r.Dispatch(ctx)

		if capturedParam != "/pkg/v1/route" {
			t.Errorf("ParamPath = %q, want '/pkg/v1/route'", capturedParam)
		}
	})
}

// TestContextAbort tests the Context.Abort function
func TestContextAbort(t *testing.T) {
	ctx := &reqresp.Context{}

	ctx.Abort()

	// After abort, the context should be marked as aborted
	// This is tested indirectly through router dispatch tests
}
