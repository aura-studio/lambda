package event

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty7_DoSafeCapturesPanic verifies that for any string panic value,
// doSafe captures the panic and returns a non-nil error containing the panic value.
//
// **Validates: Requirements 6.6**
func TestProperty7_DoSafeCapturesPanic(t *testing.T) {
	engine := &Engine{
		Options: NewOptions(),
		Router:  NewRouter(),
	}

	f := func(panicValue string) bool {
		err := engine.doSafe(func() {
			panic(panicValue)
		})
		if err == nil {
			return false
		}
		return strings.Contains(err.Error(), panicValue)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 7 (doSafe captures panic) failed: %v", err)
	}
}

// TestProperty7_DoSafeNoPanic verifies that doSafe returns nil when no panic occurs.
//
// **Validates: Requirements 6.6**
func TestProperty7_DoSafeNoPanic(t *testing.T) {
	engine := &Engine{
		Options: NewOptions(),
		Router:  NewRouter(),
	}

	err := engine.doSafe(func() {})
	if err != nil {
		t.Errorf("Expected nil error when no panic, got: %v", err)
	}
}

// TestProperty8_DoDebugCapturesStdout verifies that for any string s,
// doDebug captures stdout output written via fmt.Fprint(os.Stdout, s).
//
// **Validates: Requirements 6.7**
func TestProperty8_DoDebugCapturesStdout(t *testing.T) {
	engine := &Engine{
		Options: NewOptions(),
		Router:  NewRouter(),
	}

	f := func(s string) bool {
		stdout, _, err := engine.doDebug(func() {
			fmt.Fprint(os.Stdout, s)
		})
		if err != nil {
			return false
		}
		return strings.Contains(stdout, s)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 8 (doDebug captures stdout) failed: %v", err)
	}
}
