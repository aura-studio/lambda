package tests

import (
	"testing"

	"github.com/aura-studio/lambda/reqresp"
	"google.golang.org/protobuf/proto"
)

// TestRequestProtobuf tests Request protobuf serialization
func TestRequestProtobuf(t *testing.T) {
	req := &reqresp.Request{
		Path:    "/api/pkg/v1/route",
		Payload: []byte("test-payload"),
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded reqresp.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if decoded.Path != req.Path {
		t.Errorf("Path = %q, want %q", decoded.Path, req.Path)
	}
	if string(decoded.Payload) != string(req.Payload) {
		t.Errorf("Payload = %q, want %q", string(decoded.Payload), string(req.Payload))
	}
}

// TestResponseProtobuf tests Response protobuf serialization
func TestResponseProtobuf(t *testing.T) {
	resp := &reqresp.Response{
		Payload: []byte("response-payload"),
		Error:   "",
	}

	// Marshal
	data, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal Response: %v", err)
	}

	// Unmarshal
	var decoded reqresp.Response
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Response: %v", err)
	}

	// Verify
	if string(decoded.Payload) != string(resp.Payload) {
		t.Errorf("Payload = %q, want %q", string(decoded.Payload), string(resp.Payload))
	}
	if decoded.Error != resp.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, resp.Error)
	}
}
