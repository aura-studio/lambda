package tests

import (
	"bytes"
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/aura-studio/lambda/event"
	"github.com/aura-studio/lambda/event/eventcli"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// Feature: event-lambda-handler
// Property 11: Client Protobuf Serialization
// *For any* items sent via the Event_Client, the payload sent to Lambda SHALL be
// valid protobuf bytes that can be deserialized to an Event_Request.
//
// **Validates: Requirements 4.4**

// Feature: event-lambda-handler
// Property 12: Client Batch Sending
// *For any* list of items sent via SendBatch(), all items SHALL be included in
// a single Event_Request.
//
// **Validates: Requirements 4.8**

// mockLambdaClient captures the payload sent to Lambda for verification
type mockLambdaClient struct {
	mu             sync.Mutex
	capturedInputs []*lambda.InvokeInput
}

func (m *mockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput,
	optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedInputs = append(m.capturedInputs, params)
	return &lambda.InvokeOutput{
		StatusCode: 202, // Event invocation returns 202 Accepted
	}, nil
}

func (m *mockLambdaClient) getLastInput() *lambda.InvokeInput {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.capturedInputs) == 0 {
		return nil
	}
	return m.capturedInputs[len(m.capturedInputs)-1]
}

func (m *mockLambdaClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedInputs = nil
}

// genClientPath generates random valid path strings for client tests
func genClientPath() gopter.Gen {
	return gen.AnyString().Map(func(s string) string {
		// Ensure path starts with /
		if len(s) == 0 || s[0] != '/' {
			return "/" + s
		}
		return s
	})
}

// genClientPayload generates random byte payloads for client tests
func genClientPayload() gopter.Gen {
	return gen.SliceOf(gen.UInt8())
}

// genClientItem generates a random eventcli.Item
func genClientItem() gopter.Gen {
	return gopter.CombineGens(
		genClientPath(),
		genClientPayload(),
	).Map(func(values []interface{}) eventcli.Item {
		return eventcli.Item{
			Path:    values[0].(string),
			Payload: values[1].([]byte),
		}
	})
}

// genClientItems generates a non-empty slice of eventcli.Item (1 to 10 items)
func genClientItems() gopter.Gen {
	// Generate a size between 1 and 10, then generate that many items
	return gen.IntRange(1, 10).FlatMap(func(size interface{}) gopter.Gen {
		n := size.(int)
		gens := make([]gopter.Gen, n)
		for i := 0; i < n; i++ {
			gens[i] = genClientItem()
		}
		return gopter.CombineGens(gens...).Map(func(values []interface{}) []eventcli.Item {
			items := make([]eventcli.Item, len(values))
			for i, v := range values {
				items[i] = v.(eventcli.Item)
			}
			return items
		})
	}, reflect.TypeOf([]eventcli.Item{}))
}

// TestEventClientProtobufSerialization tests Property 11: Client Protobuf Serialization
// For any items sent via the Event_Client, the payload sent to Lambda SHALL be
// valid protobuf bytes that can be deserialized to an Event_Request.
func TestEventClientProtobufSerialization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100 // Run at least 100 iterations
	properties := gopter.NewProperties(parameters)

	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	// Test Send() method - single item
	properties.Property("Send: payload is valid protobuf deserializable to Event_Request", prop.ForAll(
		func(path string, payload []byte) bool {
			mock.reset()

			err := client.Send(context.Background(), path, payload)
			if err != nil {
				t.Logf("Send failed: %v", err)
				return false
			}

			input := mock.getLastInput()
			if input == nil {
				t.Log("No input captured")
				return false
			}

			// Verify the payload is valid protobuf
			var req event.Request
			if err := proto.Unmarshal(input.Payload, &req); err != nil {
				t.Logf("Failed to unmarshal payload as Event_Request: %v", err)
				return false
			}

			// Verify the request contains the expected item
			if len(req.Items) != 1 {
				t.Logf("Expected 1 item, got %d", len(req.Items))
				return false
			}

			if req.Items[0].Path != path {
				t.Logf("Path mismatch: expected %q, got %q", path, req.Items[0].Path)
				return false
			}

			if !bytes.Equal(req.Items[0].Payload, payload) {
				t.Log("Payload mismatch")
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// TestEventClientBatchSending tests Property 12: Client Batch Sending
// For any list of items sent via SendBatch(), all items SHALL be included in
// a single Event_Request.
func TestEventClientBatchSending(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100 // Run at least 100 iterations
	properties := gopter.NewProperties(parameters)

	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	properties.Property("SendBatch: all items included in single Event_Request", prop.ForAll(
		func(items []eventcli.Item) bool {
			mock.reset()

			err := client.SendBatch(context.Background(), items)
			if err != nil {
				t.Logf("SendBatch failed: %v", err)
				return false
			}

			input := mock.getLastInput()
			if input == nil {
				t.Log("No input captured")
				return false
			}

			// Verify the payload is valid protobuf
			var req event.Request
			if err := proto.Unmarshal(input.Payload, &req); err != nil {
				t.Logf("Failed to unmarshal payload as Event_Request: %v", err)
				return false
			}

			// Verify all items are included
			if len(req.Items) != len(items) {
				t.Logf("Item count mismatch: expected %d, got %d", len(items), len(req.Items))
				return false
			}

			// Verify each item matches
			for i, item := range items {
				if req.Items[i].Path != item.Path {
					t.Logf("Item %d path mismatch: expected %q, got %q", i, item.Path, req.Items[i].Path)
					return false
				}
				if !bytes.Equal(req.Items[i].Payload, item.Payload) {
					t.Logf("Item %d payload mismatch", i)
					return false
				}
			}

			return true
		},
		genClientItems(),
	))

	properties.TestingRun(t)
}

// TestEventClientSendUsesEventInvocationType verifies that Send uses Event invocation type
func TestEventClientSendUsesEventInvocationType(t *testing.T) {
	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	err := client.Send(context.Background(), "/api/test/v1/route", []byte("payload"))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify InvocationType is Event
	if input.InvocationType != types.InvocationTypeEvent {
		t.Errorf("Expected InvocationType %v, got %v", types.InvocationTypeEvent, input.InvocationType)
	}
}

// TestEventClientSendBatchUsesEventInvocationType verifies that SendBatch uses Event invocation type
func TestEventClientSendBatchUsesEventInvocationType(t *testing.T) {
	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	items := []eventcli.Item{
		{Path: "/api/pkg1/v1/route1", Payload: []byte("payload1")},
		{Path: "/api/pkg2/v2/route2", Payload: []byte("payload2")},
	}

	err := client.SendBatch(context.Background(), items)
	if err != nil {
		t.Fatalf("SendBatch failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify InvocationType is Event
	if input.InvocationType != types.InvocationTypeEvent {
		t.Errorf("Expected InvocationType %v, got %v", types.InvocationTypeEvent, input.InvocationType)
	}
}

// TestEventClientFunctionName verifies that the function name is correctly set
func TestEventClientFunctionName(t *testing.T) {
	mock := &mockLambdaClient{}
	functionName := "my-lambda-function"
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName(functionName),
	)

	err := client.Send(context.Background(), "/api/test/v1/route", []byte("payload"))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify FunctionName is correctly set
	if input.FunctionName == nil || *input.FunctionName != functionName {
		t.Errorf("Expected FunctionName %q, got %v", functionName, input.FunctionName)
	}
}

// TestEventClientSendBatchWithEmptyItems tests SendBatch with empty items slice
func TestEventClientSendBatchWithEmptyItems(t *testing.T) {
	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	err := client.SendBatch(context.Background(), []eventcli.Item{})
	if err != nil {
		t.Fatalf("SendBatch with empty items failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify the payload is valid protobuf with empty items
	var req event.Request
	if err := proto.Unmarshal(input.Payload, &req); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if len(req.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(req.Items))
	}
}

// TestEventClientSendWithNilPayload tests Send with nil payload
func TestEventClientSendWithNilPayload(t *testing.T) {
	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	err := client.Send(context.Background(), "/api/test/v1/route", nil)
	if err != nil {
		t.Fatalf("Send with nil payload failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify the payload is valid protobuf
	var req event.Request
	if err := proto.Unmarshal(input.Payload, &req); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if len(req.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(req.Items))
	}

	if req.Items[0].Path != "/api/test/v1/route" {
		t.Errorf("Path mismatch: expected %q, got %q", "/api/test/v1/route", req.Items[0].Path)
	}

	if len(req.Items[0].Payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(req.Items[0].Payload))
	}
}

// TestEventClientSendBatchWithLargePayload tests SendBatch with large payloads
func TestEventClientSendBatchWithLargePayload(t *testing.T) {
	mock := &mockLambdaClient{}
	client := eventcli.NewClient(
		eventcli.WithLambdaClient(mock),
		eventcli.WithFunctionName("test-function"),
	)

	// Create a large payload (100KB)
	largePayload := make([]byte, 100*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	items := []eventcli.Item{
		{Path: "/api/pkg1/v1/upload", Payload: largePayload},
		{Path: "/api/pkg2/v2/upload", Payload: largePayload},
	}

	err := client.SendBatch(context.Background(), items)
	if err != nil {
		t.Fatalf("SendBatch with large payload failed: %v", err)
	}

	input := mock.getLastInput()
	if input == nil {
		t.Fatal("No input captured")
	}

	// Verify the payload is valid protobuf
	var req event.Request
	if err := proto.Unmarshal(input.Payload, &req); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if len(req.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(req.Items))
	}

	// Verify large payloads are preserved
	for i, item := range req.Items {
		if !bytes.Equal(item.Payload, largePayload) {
			t.Errorf("Item %d: large payload not preserved correctly", i)
		}
	}
}
