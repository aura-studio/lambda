package tests

import (
	"testing"

	lambdasqs "github.com/aura-studio/lambda/sqs"
	"google.golang.org/protobuf/proto"
)

// TestSQSRequestProtobuf tests Request protobuf serialization
func TestSQSRequestProtobuf(t *testing.T) {
	req := &lambdasqs.Request{
		RequestSqsId:  "req-sqs-123",
		ResponseSqsId: "resp-sqs-456",
		CorrelationId: "corr-789",
		Path:          "/api/pkg/v1/route",
		Payload:       []byte("test-payload"),
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded lambdasqs.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if decoded.RequestSqsId != req.RequestSqsId {
		t.Errorf("RequestSqsId = %q, want %q", decoded.RequestSqsId, req.RequestSqsId)
	}
	if decoded.ResponseSqsId != req.ResponseSqsId {
		t.Errorf("ResponseSqsId = %q, want %q", decoded.ResponseSqsId, req.ResponseSqsId)
	}
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

// TestSQSResponseProtobuf tests Response protobuf serialization
func TestSQSResponseProtobuf(t *testing.T) {
	resp := &lambdasqs.Response{
		RequestSqsId:  "req-sqs-123",
		ResponseSqsId: "resp-sqs-456",
		CorrelationId: "corr-789",
		Payload:       []byte("response-payload"),
		Error:         "",
	}

	// Marshal
	data, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal Response: %v", err)
	}

	// Unmarshal
	var decoded lambdasqs.Response
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Response: %v", err)
	}

	// Verify
	if decoded.RequestSqsId != resp.RequestSqsId {
		t.Errorf("RequestSqsId = %q, want %q", decoded.RequestSqsId, resp.RequestSqsId)
	}
	if decoded.ResponseSqsId != resp.ResponseSqsId {
		t.Errorf("ResponseSqsId = %q, want %q", decoded.ResponseSqsId, resp.ResponseSqsId)
	}
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

// TestSQSResponseProtobufWithError tests Response protobuf with error field
func TestSQSResponseProtobufWithError(t *testing.T) {
	resp := &lambdasqs.Response{
		RequestSqsId:  "req-sqs-err",
		ResponseSqsId: "resp-sqs-err",
		CorrelationId: "corr-err",
		Payload:       nil,
		Error:         "something went wrong",
	}

	// Marshal
	data, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal Response: %v", err)
	}

	// Unmarshal
	var decoded lambdasqs.Response
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Response: %v", err)
	}

	// Verify
	if decoded.Error != resp.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, resp.Error)
	}
}

// TestSQSRequestProtobufEmptyPayload tests Request with empty payload
func TestSQSRequestProtobufEmptyPayload(t *testing.T) {
	req := &lambdasqs.Request{
		RequestSqsId:  "",
		ResponseSqsId: "",
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
	var decoded lambdasqs.Request
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

// TestSQSRequestProtobufLargePayload tests Request with large payload
func TestSQSRequestProtobufLargePayload(t *testing.T) {
	// Create a large payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	req := &lambdasqs.Request{
		RequestSqsId:  "large-req",
		ResponseSqsId: "large-resp",
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
	var decoded lambdasqs.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if len(decoded.Payload) != len(largePayload) {
		t.Errorf("Payload length = %d, want %d", len(decoded.Payload), len(largePayload))
	}
}

// TestSQSRequestProtobufAllFields tests Request with all fields populated
func TestSQSRequestProtobufAllFields(t *testing.T) {
	req := &lambdasqs.Request{
		RequestSqsId:  "full-req-sqs",
		ResponseSqsId: "full-resp-sqs",
		CorrelationId: "full-corr",
		Path:          "/api/full/v1/test",
		Payload:       []byte(`{"full":"payload"}`),
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded lambdasqs.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify all fields
	if decoded.RequestSqsId != req.RequestSqsId {
		t.Errorf("RequestSqsId = %q, want %q", decoded.RequestSqsId, req.RequestSqsId)
	}
	if decoded.ResponseSqsId != req.ResponseSqsId {
		t.Errorf("ResponseSqsId = %q, want %q", decoded.ResponseSqsId, req.ResponseSqsId)
	}
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
