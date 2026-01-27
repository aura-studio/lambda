package tests

import (
	"testing"

	"github.com/aura-studio/lambda/reqresp"
	"google.golang.org/protobuf/proto"
)

// TestRequestProtobuf tests Request protobuf serialization
func TestRequestProtobuf(t *testing.T) {
	req := &reqresp.Request{
		CorrelationId: "test-123",
		Path:          "/api/pkg/v1/route",
		Payload:       []byte("test-payload"),
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
	if decoded.CorrelationId != req.CorrelationId {
		t.Errorf("CorrelationId = %q, want %q", decoded.CorrelationId, req.CorrelationId)
	}
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
		CorrelationId: "test-456",
		Payload:       []byte("response-payload"),
		Error:         "",
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
	if decoded.CorrelationId != resp.CorrelationId {
		t.Errorf("CorrelationId = %q, want %q", decoded.CorrelationId, resp.CorrelationId)
	}
	if string(decoded.Payload) != string(resp.Payload) {
		t.Errorf("Payload = %q, want %q", string(decoded.Payload), string(resp.Payload))
	}
	if decoded.Error != resp.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, resp.Error)
	}
}

// TestResponseProtobufWithError tests Response protobuf with error field
func TestResponseProtobufWithError(t *testing.T) {
	resp := &reqresp.Response{
		CorrelationId: "test-789",
		Payload:       nil,
		Error:         "something went wrong",
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
	if decoded.Error != resp.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, resp.Error)
	}
}

// TestRequestProtobufEmptyPayload tests Request with empty payload
func TestRequestProtobufEmptyPayload(t *testing.T) {
	req := &reqresp.Request{
		CorrelationId: "empty-payload",
		Path:          "/health-check",
		Payload:       nil,
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
	if decoded.CorrelationId != req.CorrelationId {
		t.Errorf("CorrelationId = %q, want %q", decoded.CorrelationId, req.CorrelationId)
	}
	if len(decoded.Payload) != 0 {
		t.Errorf("Payload should be empty, got %d bytes", len(decoded.Payload))
	}
}

// TestRequestProtobufLargePayload tests Request with large payload
func TestRequestProtobufLargePayload(t *testing.T) {
	// Create a large payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	req := &reqresp.Request{
		CorrelationId: "large-payload",
		Path:          "/api/pkg/v1/upload",
		Payload:       largePayload,
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
	if len(decoded.Payload) != len(largePayload) {
		t.Errorf("Payload length = %d, want %d", len(decoded.Payload), len(largePayload))
	}
}
