package tests

import (
	"testing"

	lambdaws "github.com/aura-studio/lambda/websocket"
)

// TestWebSocketConfigDefaults tests that a fresh Options keeps built-in defaults
func TestWebSocketConfigDefaults(t *testing.T) {
	o := lambdaws.NewOptions()
	if o.DebugMode {
		t.Error("DebugMode default should be false")
	}
	if !o.ReplyMode {
		t.Error("ReplyMode default should be true")
	}
}

// TestWebSocketWithConfig tests parsing a valid websocket YAML config
func TestWebSocketWithConfig(t *testing.T) {
	opt := lambdaws.WithConfig([]byte("mode:\n  debug: true\n  reply: false\n"))
	o := lambdaws.NewOptions(opt)

	if !o.DebugMode {
		t.Error("DebugMode should be true")
	}
	if o.ReplyMode {
		t.Error("ReplyMode should be false")
	}
}

// TestWebSocketWithConfigKeepsReplyDefault tests that omitting reply keeps the default true
func TestWebSocketWithConfigKeepsReplyDefault(t *testing.T) {
	opt := lambdaws.WithConfig([]byte("mode:\n  debug: true\n"))
	o := lambdaws.NewOptions(opt)

	if !o.ReplyMode {
		t.Error("ReplyMode should stay true when reply is omitted")
	}
}

// TestWebSocketWithConfigInvalid tests that invalid YAML panics on apply
func TestWebSocketWithConfigInvalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("WithConfig with invalid YAML should panic")
		}
	}()

	opt := lambdaws.WithConfig([]byte("mode: [not-valid"))
	lambdaws.NewOptions(opt)
}

// TestWebSocketConfigCandidates tests the default config candidates list
func TestWebSocketConfigCandidates(t *testing.T) {
	candidates := lambdaws.DefaultConfigCandidates()
	if len(candidates) == 0 {
		t.Error("DefaultConfigCandidates should not be empty")
	}
	found := false
	for _, c := range candidates {
		if c == "websocket.yaml" {
			found = true
		}
	}
	if !found {
		t.Error("candidates should include websocket.yaml")
	}
}
