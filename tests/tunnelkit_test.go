package tests

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aura-studio/lambda/tunnelkit"
)

func buildEnvelope(t *testing.T, meta map[string]any, payload string) string {
	t.Helper()
	if meta == nil {
		meta = map[string]any{}
	}
	b, err := json.Marshal(map[string]any{
		"meta": meta,
		"data": base64.StdEncoding.EncodeToString([]byte(payload)),
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return string(b)
}

func parseEnvelope(t *testing.T, rsp string) (map[string]any, string) {
	t.Helper()
	var env struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}
	if err := json.Unmarshal([]byte(rsp), &env); err != nil {
		t.Fatalf("response is not an envelope: %v (raw: %s)", err, rsp)
	}
	data, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	return env.Meta, string(data)
}

// TestTunnelkitRoundTrip tests a registered handler through the full envelope cycle
func TestTunnelkitRoundTrip(t *testing.T) {
	tn := tunnelkit.New("")
	tn.Handle("/echo", func(c *tunnelkit.Context) error {
		return c.JSON(map[string]any{"route": c.Route, "echo": json.RawMessage(c.Payload)})
	})

	rsp := tn.Invoke("/echo", buildEnvelope(t, nil, `{"a":1}`))
	meta, data := parseEnvelope(t, rsp)

	if meta["Error"] != nil {
		t.Fatalf("unexpected error meta: %v", meta["Error"])
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if got["route"] != "/echo" {
		t.Errorf("route = %v, want '/echo'", got["route"])
	}
}

// TestTunnelkitUnmarshalPayload tests the payload decoding helper
func TestTunnelkitUnmarshalPayload(t *testing.T) {
	tn := tunnelkit.New("")
	var in struct {
		Amount int `json:"amount"`
	}
	tn.Handle("/bet", func(c *tunnelkit.Context) error {
		if err := c.UnmarshalPayload(&in); err != nil {
			return err
		}
		return nil
	})

	rsp := tn.Invoke("/bet", buildEnvelope(t, nil, `{"amount":5}`))
	meta, _ := parseEnvelope(t, rsp)
	if meta["Error"] != nil {
		t.Fatalf("unexpected error meta: %v", meta["Error"])
	}
	if in.Amount != 5 {
		t.Errorf("amount = %d, want 5", in.Amount)
	}
}

// TestTunnelkitUnknownRoute tests the unknown route error convention
func TestTunnelkitUnknownRoute(t *testing.T) {
	tn := tunnelkit.New("")
	rsp := tn.Invoke("/nosuch", buildEnvelope(t, nil, ""))
	meta, _ := parseEnvelope(t, rsp)

	errMsg, _ := meta["Error"].(string)
	if errMsg == "" {
		t.Fatal("unknown route should produce meta.Error")
	}
	if errMsg != "unknown route: /nosuch" {
		t.Errorf("error = %q, want 'unknown route: /nosuch'", errMsg)
	}
}

// TestTunnelkitHandlerError tests that a handler error maps to meta.Error
func TestTunnelkitHandlerError(t *testing.T) {
	tn := tunnelkit.New("")
	tn.Handle("/fail", func(c *tunnelkit.Context) error {
		return errors.New("boom")
	})

	rsp := tn.Invoke("/fail", buildEnvelope(t, nil, ""))
	meta, data := parseEnvelope(t, rsp)
	if meta["Error"] != "boom" {
		t.Errorf("meta.Error = %v, want 'boom'", meta["Error"])
	}
	if data != "" {
		t.Errorf("data should be empty on error, got %q", data)
	}
}

// TestTunnelkitInvalidEnvelope tests malformed request handling
func TestTunnelkitInvalidEnvelope(t *testing.T) {
	tn := tunnelkit.New("")
	rsp := tn.Invoke("/echo", "not-json-at-all")
	meta, _ := parseEnvelope(t, rsp)
	if meta["Error"] == nil {
		t.Error("invalid envelope should produce meta.Error")
	}
}

// TestTunnelkitMetaGeneration tests Meta() generation and explicit meta passthrough
func TestTunnelkitMetaGeneration(t *testing.T) {
	tn := tunnelkit.New("")
	tn.Handle("/a", func(c *tunnelkit.Context) error { return nil })
	tn.Handle("/b", func(c *tunnelkit.Context) error { return nil })

	var generated struct {
		Routes []string `json:"routes"`
	}
	if err := json.Unmarshal([]byte(tn.Meta()), &generated); err != nil {
		t.Fatalf("generated Meta is not JSON: %v", err)
	}
	if len(generated.Routes) != 2 || generated.Routes[0] != "/a" || generated.Routes[1] != "/b" {
		t.Errorf("generated routes = %v, want [/a /b] in registration order", generated.Routes)
	}

	explicit := tunnelkit.New(`{"name":"custom"}`)
	if explicit.Meta() != `{"name":"custom"}` {
		t.Errorf("explicit Meta passthrough failed: %s", explicit.Meta())
	}
}

// TestTunnelkitStringAndData tests raw response writers
func TestTunnelkitStringAndData(t *testing.T) {
	tn := tunnelkit.New("")
	tn.Handle("/text", func(c *tunnelkit.Context) error {
		c.String("plain")
		return nil
	})
	tn.Handle("/bytes", func(c *tunnelkit.Context) error {
		c.Data([]byte{0x01, 0x02})
		return nil
	})

	_, data := parseEnvelope(t, tn.Invoke("/text", buildEnvelope(t, nil, "")))
	if data != "plain" {
		t.Errorf("data = %q, want 'plain'", data)
	}

	_, raw := parseEnvelope(t, tn.Invoke("/bytes", buildEnvelope(t, nil, "")))
	if raw != string([]byte{0x01, 0x02}) {
		t.Errorf("raw data mismatch: %v", []byte(raw))
	}
}

// TestTunnelkitRequestMeta tests that request meta reaches the handler
func TestTunnelkitRequestMeta(t *testing.T) {
	tn := tunnelkit.New("")
	var gotMeta map[string]any
	tn.Handle("/meta", func(c *tunnelkit.Context) error {
		gotMeta = c.Meta
		return nil
	})

	tn.Invoke("/meta", buildEnvelope(t, map[string]any{"traceId": "t-1"}, ""))
	if gotMeta["traceId"] != "t-1" {
		t.Errorf("request meta not propagated: %v", gotMeta)
	}
}
