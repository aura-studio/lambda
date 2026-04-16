package tests

import (
	"context"
	"strings"
	"testing"
	"testing/quick"

	"github.com/aura-studio/lambda/event"
)

// --- Property-Based Tests for Task 6.2 ---

// TestProperty9_EngineInvokeRouteDispatch verifies that "/" and "/health-check"
// return nil error, and unregistered paths return non-nil error.
//
// **Validates: Requirements 7.4, 7.5**
func TestProperty9_EngineInvokeRouteDispatch(t *testing.T) {
	engine := event.NewEngine(nil, nil)

	// Sub-property: "/" returns nil error
	t.Run("root_path_returns_nil_error", func(t *testing.T) {
		err := engine.Invoke(context.Background(), &event.Request{Path: "/"})
		if err != nil {
			t.Errorf("Invoke(\"/\") returned error: %v, want nil", err)
		}
	})

	// Sub-property: "/health-check" returns nil error
	t.Run("health_check_returns_nil_error", func(t *testing.T) {
		err := engine.Invoke(context.Background(), &event.Request{Path: "/health-check"})
		if err != nil {
			t.Errorf("Invoke(\"/health-check\") returned error: %v, want nil", err)
		}
	})

	// Sub-property: random unregistered paths return non-nil error
	t.Run("unregistered_paths_return_error", func(t *testing.T) {
		f := func(suffix string) bool {
			// Filter out paths that match registered routes
			path := "/" + suffix

			// Skip paths that match registered routes:
			// "/" is registered (exact match)
			if path == "/" {
				return true
			}
			// "/health-check" is registered
			if path == "/health-check" {
				return true
			}
			// "/api/*path" matches anything starting with "/api/"
			if strings.HasPrefix(path, "/api/") {
				return true
			}
			// "/_/api/*path" matches anything starting with "/_/api/"
			if strings.HasPrefix(path, "/_/api/") {
				return true
			}
			// "/meta/*path" matches anything starting with "/meta/"
			if strings.HasPrefix(path, "/meta/") {
				return true
			}

			err := engine.Invoke(context.Background(), &event.Request{Path: path})
			return err != nil
		}

		cfg := &quick.Config{MaxCount: 100}
		if err := quick.Check(f, cfg); err != nil {
			t.Errorf("Property 9 (unregistered paths return error) failed: %v", err)
		}
	})
}

// --- Unit Tests for Task 6.3 ---

// TestEventEngineCreation verifies that NewEngine returns a non-nil Engine
// with Options, Router, and Dynamic properly initialized.
//
// **Validates: Requirements 7.1, 7.2, 7.3**
func TestEventEngineCreation(t *testing.T) {
	engine := event.NewEngine(nil, nil)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.Options == nil {
		t.Error("Engine.Options should not be nil")
	}

	if engine.Router == nil {
		t.Error("Engine.Router should not be nil")
	}

	if engine.Dynamic == nil {
		t.Error("Engine.Dynamic should not be nil")
	}
}

// TestEventEngineCreationWithOptions verifies that NewEngine applies the
// provided options correctly.
//
// **Validates: Requirements 7.1, 7.2**
func TestEventEngineCreationWithOptions(t *testing.T) {
	engine := event.NewEngine([]event.Option{
		event.WithDebugMode(true),
	}, nil)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if !engine.DebugMode {
		t.Error("DebugMode should be true")
	}
}

// TestEventEngineInvokeRootNoResponse verifies that Invoke on "/" returns
// only nil error, reflecting the fire-and-forget semantics (no Response).
//
// **Validates: Requirements 7.5**
func TestEventEngineInvokeRootNoResponse(t *testing.T) {
	engine := event.NewEngine(nil, nil)

	err := engine.Invoke(context.Background(), &event.Request{Path: "/"})
	if err != nil {
		t.Fatalf("Invoke(\"/\") returned error: %v, want nil", err)
	}
}

// TestEventEngineInvokeHealthCheck verifies that Invoke on "/health-check"
// returns nil error.
//
// **Validates: Requirements 7.5**
func TestEventEngineInvokeHealthCheck(t *testing.T) {
	engine := event.NewEngine(nil, nil)

	err := engine.Invoke(context.Background(), &event.Request{Path: "/health-check"})
	if err != nil {
		t.Fatalf("Invoke(\"/health-check\") returned error: %v, want nil", err)
	}
}

// TestEventEngineInvokeDebugMode verifies that the engine works correctly
// when DebugMode is enabled. The engine should log request/response info
// and still return nil error for valid paths.
//
// **Validates: Requirements 7.5, 7.3**
func TestEventEngineInvokeDebugMode(t *testing.T) {
	engine := event.NewEngine([]event.Option{
		event.WithDebugMode(true),
	}, nil)

	err := engine.Invoke(context.Background(), &event.Request{
		Path:    "/",
		Payload: []byte("test-payload"),
	})
	if err != nil {
		t.Fatalf("Invoke with DebugMode=true returned error: %v, want nil", err)
	}
}

// TestEventEngineInvokeUnregisteredPath verifies that Invoke on an
// unregistered path returns a non-nil error.
//
// **Validates: Requirements 7.5**
func TestEventEngineInvokeUnregisteredPath(t *testing.T) {
	engine := event.NewEngine(nil, nil)

	err := engine.Invoke(context.Background(), &event.Request{
		Path: "/nonexistent/path",
	})
	if err == nil {
		t.Error("Invoke on unregistered path should return non-nil error")
	}
}
