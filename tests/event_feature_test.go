package tests

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/aura-studio/lambda/event"
	"github.com/aura-studio/lambda/server"
)

// =============================================================================
// Property 11: Server 配置解析 event YAML
// Feature: lambda-event-module
// **Validates: Requirements 10.4**
// =============================================================================

func TestProperty11_ServerConfigParsesEventYAML(t *testing.T) {
	f := func(debug bool) bool {
		// Construct a server YAML string with embedded event config
		yaml := []byte(fmt.Sprintf(
			"lambda: event\nevent:\n  mode:\n    debug: %v\n", debug,
		))

		// Parse it using server.WithServeConfig
		opt := server.WithServeConfig(yaml)

		// Apply the resulting option to a server.Options
		options := &server.Options{}
		opt.Apply(options)

		// Verify the lambda type is correctly parsed
		if options.Lambda != "event" {
			t.Logf("Lambda = %q, want 'event'", options.Lambda)
			return false
		}

		// Verify the Event slice contains the parsed event option
		if len(options.Event) == 0 {
			t.Log("Event options slice is empty, expected at least one option")
			return false
		}

		// Apply the event options to a new event.Options and check DebugMode
		eventOpts := event.NewOptions(options.Event...)
		if eventOpts.DebugMode != debug {
			t.Logf("DebugMode = %v, want %v", eventOpts.DebugMode, debug)
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 11 failed: %v", err)
	}
}

// =============================================================================
// Unit Tests: Server 集成
// =============================================================================

// TestServerConfig_ParsesEventField verifies that WithServeConfig correctly
// parses the "event" section from a server YAML config and populates the
// Event options slice.
// Validates: Requirements 10.1, 10.4
func TestServerConfig_ParsesEventField(t *testing.T) {
	yaml := []byte(
		"lambda: event\n" +
			"event:\n" +
			"  mode:\n" +
			"    debug: true\n",
	)

	opt := server.WithServeConfig(yaml)
	options := &server.Options{}
	opt.Apply(options)

	if options.Lambda != "event" {
		t.Fatalf("Lambda = %q, want 'event'", options.Lambda)
	}
	if len(options.Event) == 0 {
		t.Fatal("Event options slice is empty, expected at least one option")
	}

	eventOpts := event.NewOptions(options.Event...)
	if !eventOpts.DebugMode {
		t.Fatal("DebugMode = false, want true")
	}
}

// TestWithEventOptions verifies that WithEventOptions correctly adds event
// options to server.Options.
// Validates: Requirements 10.2, 10.3
func TestWithEventOptions(t *testing.T) {
	opt := server.WithEventOptions(event.WithDebugMode(true))
	options := &server.Options{}
	opt.Apply(options)

	if len(options.Event) != 1 {
		t.Fatalf("Event options len = %d, want 1", len(options.Event))
	}

	eventOpts := event.NewOptions(options.Event...)
	if !eventOpts.DebugMode {
		t.Fatal("DebugMode = false, want true after WithEventOptions")
	}
}

// TestServerConfig_LambdaEventModeSelection verifies that when the server
// YAML config has `lambda: event`, the parsed Lambda field is "event".
// Validates: Requirements 10.5
func TestServerConfig_LambdaEventModeSelection(t *testing.T) {
	yaml := []byte("lambda: event\n")

	opt := server.WithServeConfig(yaml)
	options := &server.Options{}
	opt.Apply(options)

	if options.Lambda != "event" {
		t.Fatalf("Lambda = %q, want 'event'", options.Lambda)
	}
}

// TestServerConfig_EventDebugMode verifies that event config with debug=true
// is correctly parsed and applied through the server config pipeline.
// Validates: Requirements 10.1, 10.4
func TestServerConfig_EventDebugMode(t *testing.T) {
	yaml := []byte(
		"lambda: event\n" +
			"event:\n" +
			"  mode:\n" +
			"    debug: true\n",
	)

	opt := server.WithServeConfig(yaml)
	options := &server.Options{}
	opt.Apply(options)

	eventOpts := event.NewOptions(options.Event...)
	if !eventOpts.DebugMode {
		t.Fatal("DebugMode = false, want true")
	}

	// Also verify debug=false
	yamlFalse := []byte(
		"lambda: event\n" +
			"event:\n" +
			"  mode:\n" +
			"    debug: false\n",
	)

	optFalse := server.WithServeConfig(yamlFalse)
	optionsFalse := &server.Options{}
	optFalse.Apply(optionsFalse)

	eventOptsFalse := event.NewOptions(optionsFalse.Event...)
	if eventOptsFalse.DebugMode {
		t.Fatal("DebugMode = true, want false")
	}
}
