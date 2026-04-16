package tests

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/aura-studio/lambda/event"
	"google.golang.org/protobuf/proto"
)

// TestProperty1_ProtobufRequestJSONRoundTrip verifies that Request JSON
// serialization followed by deserialization preserves Path and Payload.
//
// **Validates: Requirements 1.1**
func TestProperty1_ProtobufRequestJSONRoundTrip(t *testing.T) {
	f := func(path string, payload []byte) bool {
		req := &event.Request{
			Path:    path,
			Payload: payload,
		}

		// JSON marshal
		data, err := json.Marshal(req)
		if err != nil {
			t.Logf("Failed to marshal: %v", err)
			return false
		}

		// JSON unmarshal
		var decoded event.Request
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Logf("Failed to unmarshal: %v", err)
			return false
		}

		// Verify Path
		if decoded.Path != req.Path {
			t.Logf("Path mismatch: got %q, want %q", decoded.Path, req.Path)
			return false
		}

		// Verify Payload using bytes.Equal
		// Note: JSON omitempty causes nil and empty []byte to both unmarshal as nil,
		// so we treat nil and empty as equivalent.
		origPayload := req.Payload
		decodedPayload := decoded.Payload
		if len(origPayload) == 0 && len(decodedPayload) == 0 {
			return true
		}
		if !bytes.Equal(decodedPayload, origPayload) {
			t.Logf("Payload mismatch: got %v, want %v", decodedPayload, origPayload)
			return false
		}

		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property 1 failed: %v", err)
	}
}

// TestEventRequestStructFields tests that Request has Path and Payload fields
// with correct values after construction.
//
// Requirements: 1.1
func TestEventRequestStructFields(t *testing.T) {
	req := &event.Request{
		Path:    "/api/pkg/v1/route",
		Payload: []byte("test-payload"),
	}

	// Verify Path field
	if req.Path != "/api/pkg/v1/route" {
		t.Errorf("Path = %q, want %q", req.Path, "/api/pkg/v1/route")
	}

	// Verify Payload field
	if !bytes.Equal(req.Payload, []byte("test-payload")) {
		t.Errorf("Payload = %q, want %q", string(req.Payload), "test-payload")
	}

	// Verify protobuf round-trip preserves fields
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	var decoded event.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	if decoded.Path != req.Path {
		t.Errorf("Decoded Path = %q, want %q", decoded.Path, req.Path)
	}
	if !bytes.Equal(decoded.Payload, req.Payload) {
		t.Errorf("Decoded Payload = %q, want %q", string(decoded.Payload), string(req.Payload))
	}
}

// TestEventNoResponseType verifies that the event package does not export a
// Response type, reflecting the fire-and-forget semantics of Event mode.
//
// Requirements: 1.4
func TestEventNoResponseType(t *testing.T) {
	// Use reflect to check that the event package does not define a Response type.
	// We look up "Response" in the event package by attempting to get a type
	// from a nil pointer of the hypothetical *event.Response.
	// Since Go doesn't allow direct package-level type lookup via reflect,
	// we verify by checking that the proto file descriptor only registers
	// one message (Request) and no message named "Response".
	reqType := reflect.TypeOf((*event.Request)(nil)).Elem()

	// Verify Request type exists and has expected fields
	if reqType.Kind() != reflect.Struct {
		t.Fatalf("event.Request should be a struct, got %v", reqType.Kind())
	}

	// Verify Request has Path field (string)
	pathField, ok := reqType.FieldByName("Path")
	if !ok {
		t.Fatal("event.Request should have a Path field")
	}
	if pathField.Type.Kind() != reflect.String {
		t.Errorf("Path field type = %v, want string", pathField.Type.Kind())
	}

	// Verify Request has Payload field ([]byte)
	payloadField, ok := reqType.FieldByName("Payload")
	if !ok {
		t.Fatal("event.Request should have a Payload field")
	}
	if payloadField.Type.Kind() != reflect.Slice || payloadField.Type.Elem().Kind() != reflect.Uint8 {
		t.Errorf("Payload field type = %v, want []byte", payloadField.Type)
	}

	// Verify the proto file descriptor only defines Request, not Response.
	// The event.proto file registers exactly 1 message type.
	fd := event.File_event_event_proto
	messages := fd.Messages()
	for i := 0; i < messages.Len(); i++ {
		msgName := string(messages.Get(i).Name())
		if msgName == "Response" {
			t.Errorf("event.proto should not define a Response message, but found one")
		}
	}

	// Also verify there is exactly 1 message (Request)
	if messages.Len() != 1 {
		t.Errorf("event.proto should define exactly 1 message, got %d", messages.Len())
	}
	if string(messages.Get(0).Name()) != "Request" {
		t.Errorf("event.proto message name = %q, want %q", messages.Get(0).Name(), "Request")
	}
}
