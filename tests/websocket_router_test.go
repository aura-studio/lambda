package tests

import (
	"testing"

	lambdaws "github.com/aura-studio/lambda/websocket"
)

// TestWebSocketRouterMatch tests exact route key matching
func TestWebSocketRouterMatch(t *testing.T) {
	r := lambdaws.NewRouter()

	var called string
	r.Handle("$connect", func(c *lambdaws.Context) { called = "connect" })
	r.Handle("subscribe", func(c *lambdaws.Context) { called = "subscribe" })

	c := &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "$connect")
	r.Dispatch(c)
	if called != "connect" {
		t.Errorf("called = %q, want 'connect'", called)
	}

	called = ""
	c = &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "subscribe")
	r.Dispatch(c)
	if called != "subscribe" {
		t.Errorf("called = %q, want 'subscribe'", called)
	}
}

// TestWebSocketRouterNoRoute tests the NoRoute fallback for unmatched keys
func TestWebSocketRouterNoRoute(t *testing.T) {
	r := lambdaws.NewRouter()
	r.Handle("$connect", func(c *lambdaws.Context) {})

	var noRouteCalled bool
	r.NoRoute(func(c *lambdaws.Context) { noRouteCalled = true })

	c := &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "unknown")
	r.Dispatch(c)
	if !noRouteCalled {
		t.Error("NoRoute handler should be called for unmatched key")
	}
}

// TestWebSocketRouterNoRouteMissing tests that an unmatched key without NoRoute sets an error
func TestWebSocketRouterNoRouteMissing(t *testing.T) {
	r := lambdaws.NewRouter()

	c := &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "unknown")
	r.Dispatch(c)
	if c.GetError() == nil {
		t.Error("Dispatch should set an error when no route matches and NoRoute is empty")
	}
}

// TestWebSocketRouterUseAbort tests that a pre handler can abort the chain
func TestWebSocketRouterUseAbort(t *testing.T) {
	r := lambdaws.NewRouter()

	r.Use(func(c *lambdaws.Context) { c.Abort() })

	var called bool
	r.Handle("$connect", func(c *lambdaws.Context) { called = true })

	c := &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "$connect")
	r.Dispatch(c)
	if called {
		t.Error("route handler should not run after abort")
	}
}

// TestWebSocketRouterErrorStopsChain tests that a handler error stops the remaining chain
func TestWebSocketRouterErrorStopsChain(t *testing.T) {
	r := lambdaws.NewRouter()

	var second bool
	r.Handle("$connect",
		func(c *lambdaws.Context) { c.Set(lambdaws.ContextError, errTestUnauthorized) },
		func(c *lambdaws.Context) { second = true },
	)

	c := &lambdaws.Context{}
	c.Set(lambdaws.ContextRouteKey, "$connect")
	r.Dispatch(c)
	if second {
		t.Error("second handler should not run after an error")
	}
}
