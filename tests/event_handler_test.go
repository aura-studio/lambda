package tests

import (
	"strings"
	"testing"

	"github.com/aura-studio/lambda/event"
)

// --- Unit Tests for Task 5.3 ---

// TestEventHandler_HealthCheck_RootPath tests that dispatching "/" returns "OK".
//
// **Validates: Requirements 6.1, 6.2**
func TestEventHandler_HealthCheck_RootPath(t *testing.T) {
	engine := &event.Engine{
		Options: event.NewOptions(),
		Router:  event.NewRouter(),
	}
	engine.InstallHandlers()

	ctx := &event.Context{}
	ctx.Set(event.ContextPath, "/")
	engine.Dispatch(ctx)

	resp := ctx.GetString(event.ContextResponse)
	if resp != "OK" {
		t.Errorf("Response = %q, want %q", resp, "OK")
	}
	if ctx.GetError() != nil {
		t.Errorf("Unexpected error: %v", ctx.GetError())
	}
}

// TestEventHandler_HealthCheck_HealthCheckPath tests that dispatching "/health-check" returns "OK".
//
// **Validates: Requirements 6.1, 6.2**
func TestEventHandler_HealthCheck_HealthCheckPath(t *testing.T) {
	engine := &event.Engine{
		Options: event.NewOptions(),
		Router:  event.NewRouter(),
	}
	engine.InstallHandlers()

	ctx := &event.Context{}
	ctx.Set(event.ContextPath, "/health-check")
	engine.Dispatch(ctx)

	resp := ctx.GetString(event.ContextResponse)
	if resp != "OK" {
		t.Errorf("Response = %q, want %q", resp, "OK")
	}
	if ctx.GetError() != nil {
		t.Errorf("Unexpected error: %v", ctx.GetError())
	}
}

// TestEventHandler_API_EmptyPath tests that the API handler sets "missing api path"
// error when ContextPath is empty.
//
// **Validates: Requirements 6.5**
func TestEventHandler_API_EmptyPath(t *testing.T) {
	engine := &event.Engine{
		Options: event.NewOptions(),
		Router:  event.NewRouter(),
	}

	ctx := &event.Context{}
	ctx.Set(event.ContextPath, "")
	engine.API(ctx)

	err := ctx.GetError()
	if err == nil {
		t.Fatal("Expected error for empty API path, got nil")
	}
	if !strings.Contains(err.Error(), "missing api path") {
		t.Errorf("Error = %q, want it to contain %q", err.Error(), "missing api path")
	}
}

// TestEventHandler_PageNotFound tests that dispatching an unregistered path
// triggers the PageNotFound handler with a "404 page not found" error.
//
// **Validates: Requirements 6.5**
func TestEventHandler_PageNotFound(t *testing.T) {
	engine := &event.Engine{
		Options: event.NewOptions(),
		Router:  event.NewRouter(),
	}
	engine.InstallHandlers()

	ctx := &event.Context{}
	ctx.Set(event.ContextPath, "/nonexistent/path")
	engine.Dispatch(ctx)

	err := ctx.GetError()
	if err == nil {
		t.Fatal("Expected error for unregistered path, got nil")
	}
	if !strings.Contains(err.Error(), "404 page not found") {
		t.Errorf("Error = %q, want it to contain %q", err.Error(), "404 page not found")
	}
}
