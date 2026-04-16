package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"

	"github.com/aura-studio/lambda/event"
)

// TestProperty2_NewOptionsAppliesDebugMode verifies that NewOptions(WithDebugMode(v))
// returns Options with DebugMode equal to v for any boolean v.
//
// **Validates: Requirements 2.3**
func TestProperty2_NewOptionsAppliesDebugMode(t *testing.T) {
	f := func(debug bool) bool {
		opts := event.NewOptions(event.WithDebugMode(debug))
		return opts.DebugMode == debug
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 2 (WithDebugMode) failed: %v", err)
	}
}

// TestProperty2_NewOptionsDefaultDebugMode verifies that NewOptions() without
// any options returns Options with DebugMode == false.
//
// **Validates: Requirements 2.3**
func TestProperty2_NewOptionsDefaultDebugMode(t *testing.T) {
	opts := event.NewOptions()
	if opts.DebugMode {
		t.Errorf("Expected default DebugMode to be false, got true")
	}
}

// TestProperty3_YAMLConfigParsesDebugMode verifies that for any boolean debug value,
// constructing a YAML string with mode.debug set to that value and parsing it via
// WithConfig produces Options with the correct DebugMode.
//
// **Validates: Requirements 3.1, 3.3**
func TestProperty3_YAMLConfigParsesDebugMode(t *testing.T) {
	f := func(debug bool) bool {
		yaml := fmt.Sprintf("mode:\n  debug: %v\n", debug)
		opts := event.NewOptions(event.WithConfig([]byte(yaml)))
		return opts.DebugMode == debug
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 3 (YAML config parsing) failed: %v", err)
	}
}

// --- Unit Tests for Task 2.6 ---

// TestEventNewOptionsDefaults tests that NewOptions() without arguments returns
// Options with default values (DebugMode == false).
//
// Requirements: 2.2, 2.4
func TestEventNewOptionsDefaults(t *testing.T) {
	opts := event.NewOptions()
	if opts.DebugMode {
		t.Fatalf("DebugMode = true, expected false for default options")
	}
}

// TestEventWithConfigInvalidYAMLPanics tests that WithConfig with invalid YAML
// panics with a descriptive error message containing "event.WithConfig".
//
// Requirements: 3.4
func TestEventWithConfigInvalidYAMLPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("Expected panic for invalid YAML, but no panic occurred")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "event.WithConfig") {
			t.Fatalf("Panic message %q does not contain 'event.WithConfig'", msg)
		}
	}()

	invalidYAML := []byte(`{{{invalid yaml`)
	_ = event.NewOptions(event.WithConfig(invalidYAML))
}

// TestEventWithConfigFileNotFoundPanics tests that WithConfigFile with a non-existent
// path panics with a descriptive error message containing "event.WithConfigFile".
//
// Requirements: 3.5
func TestEventWithConfigFileNotFoundPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("Expected panic for non-existent config file, but no panic occurred")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "event.WithConfigFile") {
			t.Fatalf("Panic message %q does not contain 'event.WithConfigFile'", msg)
		}
	}()

	_ = event.NewOptions(event.WithConfigFile("/non/existent/path/event.yml"))
}

// TestEventDefaultConfigCandidates tests that DefaultConfigCandidates returns
// the expected list of candidate paths.
//
// Requirements: 4.1
func TestEventDefaultConfigCandidates(t *testing.T) {
	candidates := event.DefaultConfigCandidates()

	if len(candidates) != 4 {
		t.Fatalf("Expected 4 candidates, got %d", len(candidates))
	}

	expected := []string{
		"event.yaml",
		"event.yml",
		filepath.FromSlash("event/event.yaml"),
		filepath.FromSlash("event/event.yml"),
	}

	for i, exp := range expected {
		if candidates[i] != exp {
			t.Fatalf("Candidate[%d] = %q, expected %q", i, candidates[i], exp)
		}
	}
}
