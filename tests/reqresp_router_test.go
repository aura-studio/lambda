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
		{"exact match root", "/", "/", "", true},
		{"exact match health-check", "/health-check", "/health-check", "", true},
		{"exact match no match", "/health-check", "/other", "", false},
		{"wildcard api with route", "/api/*path", "/api/pkg/v1/route", "/pkg/v1/route", true},
		{"wildcard api no route", "/api/*path", "/api/", "/", true},
		{"wildcard debug api", "/_/api/*path", "/_/api/pkg/v1/route", "/pkg/v1/route", true},
		{"wildcard deep path", "/api/*path", "/api/pkg/v1/users/123/profile/settings", "/pkg/v1/users/123/profile/settings", true},
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

func newContext(path string) *reqresp.Context {
	c := &reqresp.Context{}
	c.Set(reqresp.ContextPath, path)
	return c
}

// TestRouterDispatch tests the Router.Dispatch function
func TestRouterDispatch(t *testing.T) {
	t.Run("dispatch to matching route", func(t *testing.T) {
		r := reqresp.NewRouter()
		called := false
		r.Handle("/test", func(c *reqresp.Context) {
			called = true
			c.Set(reqresp.ContextResponse, "test-response")
		})

		ctx := newContext("/test")
		r.Dispatch(ctx)

		if !called {
			t.Error("Handler was not called")
		}
		if ctx.GetString(reqresp.ContextResponse) != "test-response" {
			t.Errorf("Response = %q, want 'test-response'", ctx.GetString(reqresp.ContextResponse))
		}
	})

	t.Run("dispatch to no route handler", func(t *testing.T) {
		r := reqresp.NewRouter()
		noRouteCalled := false
		r.NoRoute(func(c *reqresp.Context) {
			noRouteCalled = true
		})

		ctx := newContext("/nonexistent")
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

		ctx := newContext("/test")
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

		ctx := newContext("/test")
		r.Dispatch(ctx)

		if handlerCalled {
			t.Error("Handler should not be called after Abort")
		}
	})

	t.Run("wildcard route sets param", func(t *testing.T) {
		r := reqresp.NewRouter()
		var capturedParam string

		r.Handle("/api/*path", func(c *reqresp.Context) {
			capturedParam = c.GetString(reqresp.ContextPath)
		})

		ctx := newContext("/api/pkg/v1/route")
		r.Dispatch(ctx)

		if capturedParam != "/pkg/v1/route" {
			t.Errorf("ParamPath = %q, want '/pkg/v1/route'", capturedParam)
		}
	})
}
