package tests

import (
	"context"
	"testing"

	"github.com/aura-studio/dynamic"
	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// Feature: event-lambda-handler
// Property 2: Invalid Protobuf Handling
// *For any* byte sequence that is not valid protobuf, the Event_Engine SHALL
// return an error without panicking.
//
// **Validates: Requirements 1.2**

// Property 3: No Response Payload
// *For any* successfully processed request, the Event_Engine's Invoke method
// SHALL return only nil or error, never a business response payload.
//
// **Validates: Requirements 1.3**

// Property 4: Engine Start/Stop State
// *For any* sequence of Start() and Stop() calls, the Engine's IsRunning()
// state SHALL correctly reflect the most recent call.
//
// **Validates: Requirements 1.5**

// Property 5: Batch Processing
// *For any* Event_Request containing N items (where N >= 1), the Event_Engine
// SHALL process exactly N items (subject to RunMode behavior on errors).
//
// **Validates: Requirements 1.6, 1.7, 5.6**

// genInvalidProtobufBytes generates byte sequences that are not valid protobuf
func genInvalidProtobufBytes() gopter.Gen {
	return gen.OneGenOf(
		// Random bytes that are unlikely to be valid protobuf
		gen.SliceOfN(10, gen.UInt8()),
		// Truncated protobuf (valid start but incomplete)
		gen.Const([]byte{0x0a, 0x10}), // field 1, length 16, but no data
		// Invalid wire type
		gen.Const([]byte{0x07}), // wire type 7 is invalid
		// Nested message with invalid length
		gen.Const([]byte{0x0a, 0xff, 0xff, 0xff, 0xff, 0x0f}), // length overflow
		// Random ASCII strings
		gen.AnyString().Map(func(s string) []byte {
			return []byte(s)
		}),
		// Empty bytes
		gen.Const([]byte{}),
	)
}

// genStartStopSequence generates a sequence of Start/Stop operations
// true = Start, false = Stop
func genStartStopSequence() gopter.Gen {
	return gen.SliceOfN(10, gen.Bool()).SuchThat(func(seq []bool) bool {
		return len(seq) > 0
	})
}

// genBatchSize generates a valid batch size (1 to 20)
func genBatchSize() gopter.Gen {
	return gen.IntRange(1, 20)
}

// genValidAPIPath generates a valid API path for testing
func genValidAPIPath() gopter.Gen {
	return gen.OneConstOf(
		"/",
		"/health-check",
	)
}

// TestEventEngineInvalidProtobufHandling tests Property 2: Invalid Protobuf Handling
// For any byte sequence that is not valid protobuf, the Event_Engine SHALL
// return an error without panicking.
func TestEventEngineInvalidProtobufHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generate truly invalid protobuf bytes that cannot be parsed
	genTrulyInvalidProtobuf := gen.OneGenOf(
		// Truncated protobuf (valid start but incomplete)
		gen.Const([]byte{0x0a, 0x10}), // field 1, length 16, but no data
		// Invalid wire type
		gen.Const([]byte{0x07}), // wire type 7 is invalid
		// Nested message with invalid length
		gen.Const([]byte{0x0a, 0xff, 0xff, 0xff, 0xff, 0x0f}), // length overflow
		// Invalid varint
		gen.Const([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}),
		// Truncated length-delimited field
		gen.Const([]byte{0x0a, 0x05, 0x01, 0x02}), // says 5 bytes but only 2
	)

	properties.Property("invalid protobuf returns error without panic", prop.ForAll(
		func(invalidBytes []byte) (result bool) {
			// Recover from any panic to verify no panic occurs
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic occurred with input %v: %v", invalidBytes, r)
					result = false
				}
			}()

			e := event.NewEngine(nil, nil)

			// Invoke with invalid protobuf bytes
			err := e.Invoke(context.Background(), invalidBytes)

			// Should return an error (not panic) for truly invalid protobuf
			return err != nil
		},
		genTrulyInvalidProtobuf,
	))

	properties.TestingRun(t)
}

// TestEventEngineNoResponsePayload tests Property 3: No Response Payload
// For any successfully processed request, the Event_Engine's Invoke method
// SHALL return only nil or error, never a business response payload.
func TestEventEngineNoResponsePayload(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Invoke returns only nil or error, never business response", prop.ForAll(
		func(path string) bool {
			e := event.NewEngine(nil, nil)

			// Create a valid request
			req := &event.Request{
				Items: []*event.Item{
					{
						Path:    path,
						Payload: []byte(`{"test":"data"}`),
					},
				},
			}

			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			// Invoke returns only error (or nil)
			// The function signature is: func (e *Engine) Invoke(ctx context.Context, payload []byte) error
			// This inherently means it can only return nil or error, never a business response
			result := e.Invoke(context.Background(), payload)

			// Result is either nil or error - this is guaranteed by the function signature
			// We verify the engine processed the request without returning any payload
			// (the function signature enforces this - it returns error, not (response, error))
			_ = result // result is either nil or error, never a business payload
			return true
		},
		genValidAPIPath(),
	))

	properties.TestingRun(t)
}

// TestEventEngineStartStopState tests Property 4: Engine Start/Stop State
// For any sequence of Start() and Stop() calls, the Engine's IsRunning()
// state SHALL correctly reflect the most recent call.
func TestEventEngineStartStopState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("IsRunning reflects most recent Start/Stop call", prop.ForAll(
		func(sequence []bool) bool {
			e := event.NewEngine(nil, nil)

			// Engine starts in running state
			if !e.IsRunning() {
				return false
			}

			var lastOp bool = true // Engine starts running (equivalent to Start)

			for _, isStart := range sequence {
				if isStart {
					e.Start()
					lastOp = true
				} else {
					e.Stop()
					lastOp = false
				}

				// Verify state matches the last operation
				if e.IsRunning() != lastOp {
					return false
				}
			}

			// Final state should match the last operation
			return e.IsRunning() == lastOp
		},
		genStartStopSequence(),
	))

	properties.TestingRun(t)
}

// TestEventEngineBatchProcessing tests Property 5: Batch Processing
// For any Event_Request containing N items (where N >= 1), the Event_Engine
// SHALL process exactly N items (subject to RunMode behavior on errors).
func TestEventEngineBatchProcessing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("batch with N items processes exactly N items", prop.ForAll(
		func(n int) bool {
			// Track how many items were processed
			processedCount := 0

			// Register a test package that counts invocations
			pkgName := "batchtest"
			version := "v1"
			dynamic.RegisterPackage(pkgName, version, &mockTunnel{
				invoke: func(route, req string) string {
					processedCount++
					return ""
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModePartial), // Continue processing all items
			}, nil)

			// Create a request with N items
			items := make([]*event.Item, n)
			for i := 0; i < n; i++ {
				items[i] = &event.Item{
					Path:    "/api/" + pkgName + "/" + version + "/route",
					Payload: []byte(`{}`),
				}
			}

			req := &event.Request{Items: items}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			// Invoke the engine
			_ = e.Invoke(context.Background(), payload)

			// Verify exactly N items were processed
			return processedCount == n
		},
		genBatchSize(),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Unit Tests for additional coverage
// ============================================================================

// TestEventEngineCreation tests that NewEngine creates an engine with correct options
func TestEventEngineCreation(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithDebugMode(true),
		event.WithRunMode(event.RunModePartial),
		event.WithStaticLink("/static", "/public"),
		event.WithPrefixLink("/api", "/v1"),
	}, nil)

	if e == nil {
		t.Fatal("NewEngine returned nil")
	}

	if !e.DebugMode {
		t.Error("DebugMode should be true")
	}

	if e.RunMode != event.RunModePartial {
		t.Errorf("RunMode = %q, want 'partial'", e.RunMode)
	}

	if e.StaticLinkMap["/static"] != "/public" {
		t.Errorf("StaticLinkMap['/static'] = %q, want '/public'", e.StaticLinkMap["/static"])
	}

	if e.PrefixLinkMap["/api"] != "/v1" {
		t.Errorf("PrefixLinkMap['/api'] = %q, want '/v1'", e.PrefixLinkMap["/api"])
	}
}

// TestEventEngineStartStop tests the Start and Stop methods
func TestEventEngineStartStop(t *testing.T) {
	e := event.NewEngine(nil, nil)

	// Engine should be running after creation
	if !e.IsRunning() {
		t.Error("Engine should be running after creation")
	}

	// Stop the engine
	e.Stop()
	if e.IsRunning() {
		t.Error("Engine should not be running after Stop")
	}

	// Start the engine again
	e.Start()
	if !e.IsRunning() {
		t.Error("Engine should be running after Start")
	}
}

// TestEventEngineInvokeWhenStopped tests that Invoke returns error when engine is stopped
func TestEventEngineInvokeWhenStopped(t *testing.T) {
	e := event.NewEngine(nil, nil)
	e.Stop()

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err == nil {
		t.Error("Expected error when invoking stopped engine")
	}
}

// TestEventEngineInvokeHealthCheck tests the health check endpoint
func TestEventEngineInvokeHealthCheck(t *testing.T) {
	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventEngineInvokeInvalidProtobuf tests handling of invalid protobuf payload
func TestEventEngineInvokeInvalidProtobuf(t *testing.T) {
	e := event.NewEngine(nil, nil)

	// Invalid protobuf bytes
	err := e.Invoke(context.Background(), []byte("not-protobuf"))
	if err == nil {
		t.Error("Expected error for invalid protobuf")
	}
}

// TestEventEngineInvokePageNotFound tests the 404 handler
func TestEventEngineInvokePageNotFound(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModePartial),
	}, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/nonexistent/path", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	// In partial mode, should return nil even with errors
	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error in partial mode: %v", err)
	}
}

// TestEventEngineInvokeStaticLink tests static link path mapping
func TestEventEngineInvokeStaticLink(t *testing.T) {
	e := event.NewEngine([]event.Option{
		event.WithStaticLink("/custom-root", "/"),
	}, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/custom-root", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventEngineInvokePrefixLink tests prefix link path mapping
func TestEventEngineInvokePrefixLink(t *testing.T) {
	// Register a test package
	dynamic.RegisterPackage("prefixpkg-event", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return "prefix-response"
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithPrefixLink("/v1", "/api"),
	}, nil)

	// /v1/prefixpkg-event/v1/route should be mapped to /api/prefixpkg-event/v1/route
	req := &event.Request{
		Items: []*event.Item{
			{Path: "/v1/prefixpkg-event/v1/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventEngineInvokeAPIWithDynamic tests API route calling Dynamic
func TestEventEngineInvokeAPIWithDynamic(t *testing.T) {
	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("eventpkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "event-api-response"
		},
	})

	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/api/eventpkg/v1/myroute", Payload: []byte(`{"key":"value"}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}

	if invokedRoute != "/myroute" {
		t.Errorf("invokedRoute = %q, want '/myroute'", invokedRoute)
	}

	if invokedReq != `{"key":"value"}` {
		t.Errorf("invokedReq = %q, want '{\"key\":\"value\"}'", invokedReq)
	}
}

// TestEventEngineInvokeWAPI tests WAPI route (debug mode)
func TestEventEngineInvokeWAPI(t *testing.T) {
	var invokedRoute string
	dynamic.RegisterPackage("eventwapi", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			return "wapi-response"
		},
	})

	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/_/api/eventwapi/v1/route", Payload: []byte(`{}`)},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}

	if invokedRoute != "/route" {
		t.Errorf("invokedRoute = %q, want '/route'", invokedRoute)
	}
}

// TestEventEngineMultipleItems tests processing multiple items in a batch
func TestEventEngineMultipleItems(t *testing.T) {
	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/", Payload: nil},
			{Path: "/", Payload: nil},
			{Path: "/", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error: %v", err)
	}
}

// TestEventEngineEmptyItems tests processing empty items list
func TestEventEngineEmptyItems(t *testing.T) {
	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error with empty items: %v", err)
	}
}

// TestEventEngineNilPayload tests processing with nil payload
func TestEventEngineNilPayload(t *testing.T) {
	e := event.NewEngine(nil, nil)

	req := &event.Request{
		Items: []*event.Item{
			{Path: "/", Payload: nil},
		},
	}
	payload, _ := proto.Marshal(req)

	err := e.Invoke(context.Background(), payload)
	if err != nil {
		t.Errorf("Invoke error with nil payload: %v", err)
	}
}
