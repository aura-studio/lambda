package tests

import (
	"bytes"
	"testing"

	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// Feature: event-lambda-handler
// Property 1: Protobuf Round-Trip
// *For any* valid Event_Request object, serializing it to protobuf bytes and then
// deserializing those bytes SHALL produce an equivalent Event_Request object.
//
// **Validates: Requirements 1.1, 5.4, 5.5**

// genPath generates random valid path strings
func genPath() gopter.Gen {
	return gen.AnyString().Map(func(s string) string {
		// Ensure path starts with /
		if len(s) == 0 || s[0] != '/' {
			return "/" + s
		}
		return s
	})
}

// genPayload generates random byte payloads
func genPayload() gopter.Gen {
	return gen.SliceOf(gen.UInt8())
}

// genItem generates a random event.Item
func genItem() gopter.Gen {
	return gopter.CombineGens(
		genPath(),
		genPayload(),
	).Map(func(values []interface{}) *event.Item {
		return &event.Item{
			Path:    values[0].(string),
			Payload: values[1].([]byte),
		}
	})
}

// genRequest generates a random event.Request with 0 to N items
func genRequest() gopter.Gen {
	return gen.SliceOf(genItem()).Map(func(items []*event.Item) *event.Request {
		return &event.Request{
			Items: items,
		}
	})
}

// itemsEqual compares two Item slices for equality
func itemsEqual(a, b []*event.Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Path != b[i].Path {
			return false
		}
		if !bytes.Equal(a[i].Payload, b[i].Payload) {
			return false
		}
	}
	return true
}

// TestEventProtobufRoundTrip tests Property 1: Protobuf Round-Trip
// For any valid Event_Request object, serializing it to protobuf bytes and then
// deserializing those bytes SHALL produce an equivalent Event_Request object.
func TestEventProtobufRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100 // Run at least 100 iterations
	properties := gopter.NewProperties(parameters)

	properties.Property("round-trip: marshal then unmarshal produces equivalent Request", prop.ForAll(
		func(req *event.Request) bool {
			// Serialize to protobuf bytes
			data, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Marshal failed: %v", err)
				return false
			}

			// Deserialize back to Request
			var decoded event.Request
			if err := proto.Unmarshal(data, &decoded); err != nil {
				t.Logf("Unmarshal failed: %v", err)
				return false
			}

			// Verify equivalence
			return itemsEqual(req.Items, decoded.Items)
		},
		genRequest(),
	))

	properties.TestingRun(t)
}

// TestEventItemProtobufRoundTrip tests round-trip for individual Item messages
func TestEventItemProtobufRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("round-trip: marshal then unmarshal produces equivalent Item", prop.ForAll(
		func(item *event.Item) bool {
			// Serialize to protobuf bytes
			data, err := proto.Marshal(item)
			if err != nil {
				t.Logf("Marshal failed: %v", err)
				return false
			}

			// Deserialize back to Item
			var decoded event.Item
			if err := proto.Unmarshal(data, &decoded); err != nil {
				t.Logf("Unmarshal failed: %v", err)
				return false
			}

			// Verify equivalence
			return item.Path == decoded.Path && bytes.Equal(item.Payload, decoded.Payload)
		},
		genItem(),
	))

	properties.TestingRun(t)
}

// TestEventProtobufRoundTripWithEmptyItems tests round-trip with empty items slice
func TestEventProtobufRoundTripWithEmptyItems(t *testing.T) {
	req := &event.Request{
		Items: []*event.Item{},
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded event.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify - empty slice should round-trip correctly
	if len(decoded.Items) != 0 {
		t.Errorf("Expected empty items, got %d items", len(decoded.Items))
	}
}

// TestEventProtobufRoundTripWithNilPayload tests round-trip with nil payload
func TestEventProtobufRoundTripWithNilPayload(t *testing.T) {
	req := &event.Request{
		Items: []*event.Item{
			{
				Path:    "/api/test/v1/route",
				Payload: nil,
			},
		},
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded event.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if len(decoded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(decoded.Items))
	}
	if decoded.Items[0].Path != req.Items[0].Path {
		t.Errorf("Path = %q, want %q", decoded.Items[0].Path, req.Items[0].Path)
	}
	if len(decoded.Items[0].Payload) != 0 {
		t.Errorf("Payload should be empty, got %d bytes", len(decoded.Items[0].Payload))
	}
}

// TestEventProtobufRoundTripWithMultipleItems tests round-trip with multiple items
func TestEventProtobufRoundTripWithMultipleItems(t *testing.T) {
	req := &event.Request{
		Items: []*event.Item{
			{
				Path:    "/api/pkg1/v1/route1",
				Payload: []byte("payload1"),
			},
			{
				Path:    "/api/pkg2/v2/route2",
				Payload: []byte("payload2"),
			},
			{
				Path:    "/api/pkg3/v3/route3",
				Payload: []byte(`{"json":"payload"}`),
			},
		},
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded event.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if !itemsEqual(req.Items, decoded.Items) {
		t.Errorf("Items do not match after round-trip")
	}
}

// TestEventProtobufRoundTripWithLargePayload tests round-trip with large payload
func TestEventProtobufRoundTripWithLargePayload(t *testing.T) {
	// Create a large payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	req := &event.Request{
		Items: []*event.Item{
			{
				Path:    "/api/pkg/v1/upload",
				Payload: largePayload,
			},
		},
	}

	// Marshal
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Request: %v", err)
	}

	// Unmarshal
	var decoded event.Request
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Request: %v", err)
	}

	// Verify
	if len(decoded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(decoded.Items))
	}
	if !bytes.Equal(decoded.Items[0].Payload, largePayload) {
		t.Errorf("Large payload does not match after round-trip")
	}
}
