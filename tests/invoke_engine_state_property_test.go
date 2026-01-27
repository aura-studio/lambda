package tests

import (
	"context"
	"testing"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// **Feature: invoke-lambda-handler, Property 6: 引擎状态控制**
// **Validates: Requirements 1.6**
//
// Property 6: 引擎状态控制
// For any Invoke_Engine 实例，调用 Stop 后引擎 SHALL 停止处理新请求；
// 调用 Start 后引擎 SHALL 恢复处理请求。

// stateTransition represents a state transition operation
type stateTransition int

const (
	transitionStart stateTransition = iota
	transitionStop
)

// genStateTransition generates random state transitions
func genStateTransition() gopter.Gen {
	return gen.IntRange(0, 1).Map(func(i int) stateTransition {
		return stateTransition(i)
	})
}

// genStateTransitionSequence generates a sequence of state transitions
func genStateTransitionSequence() gopter.Gen {
	return gen.SliceOfN(10, genStateTransition())
}

// genValidPath generates valid request paths for testing
func genValidRequestPath() gopter.Gen {
	return gen.OneConstOf("/health-check", "/", "/health-check")
}

// genCorrelationId generates random correlation IDs
func genCorrelationId() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) > 20 {
			return s[:20]
		}
		return s
	})
}

// TestEngineStopPreventsNewRequests tests that Stop() prevents new requests from being processed
// **Validates: Requirements 1.6**
func TestEngineStopPreventsNewRequests(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine Stop: calling Stop() prevents new requests from being processed", prop.ForAll(
		func(correlationId string, path string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Verify engine is running initially
			if !engine.IsRunning() {
				t.Logf("Engine should be running after creation")
				return false
			}

			// Stop the engine
			engine.Stop()

			// Verify engine is stopped
			if engine.IsRunning() {
				t.Logf("Engine should not be running after Stop()")
				return false
			}

			// Create a request
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          path,
				Payload:       []byte("test"),
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			// Invoke the engine
			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			// Unmarshal response
			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Response should contain error indicating engine is stopped
			if resp.Error == "" {
				t.Logf("Expected error in response when engine is stopped, got empty error")
				return false
			}

			if resp.Error != "engine is stopped" {
				t.Logf("Expected error 'engine is stopped', got %q", resp.Error)
				return false
			}

			return true
		},
		genCorrelationId(),
		genValidRequestPath(),
	))

	properties.TestingRun(t)
}

// TestEngineStartResumesRequests tests that Start() resumes request processing
// **Validates: Requirements 1.6**
func TestEngineStartResumesRequests(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine Start: calling Start() after Stop() resumes request processing", prop.ForAll(
		func(correlationId string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Stop the engine
			engine.Stop()

			// Verify engine is stopped
			if engine.IsRunning() {
				t.Logf("Engine should not be running after Stop()")
				return false
			}

			// Start the engine
			engine.Start()

			// Verify engine is running
			if !engine.IsRunning() {
				t.Logf("Engine should be running after Start()")
				return false
			}

			// Create a health-check request
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          "/health-check",
				Payload:       nil,
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			// Invoke the engine
			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			// Unmarshal response
			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Response should be successful (no error)
			if resp.Error != "" {
				t.Logf("Expected no error after Start(), got %q", resp.Error)
				return false
			}

			// Response should contain "OK" for health-check
			if string(resp.Payload) != "OK" {
				t.Logf("Expected payload 'OK', got %q", string(resp.Payload))
				return false
			}

			return true
		},
		genCorrelationId(),
	))

	properties.TestingRun(t)
}

// TestEngineStateTransitionSequence tests that engine state follows transition sequence correctly
// **Validates: Requirements 1.6**
func TestEngineStateTransitionSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine state transitions: state follows transition sequence correctly", prop.ForAll(
		func(transitions []stateTransition) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Engine starts in running state
			expectedRunning := true

			// Apply each transition and verify state
			for i, transition := range transitions {
				switch transition {
				case transitionStart:
					engine.Start()
					expectedRunning = true
				case transitionStop:
					engine.Stop()
					expectedRunning = false
				}

				// Verify state matches expected
				if engine.IsRunning() != expectedRunning {
					t.Logf("State mismatch at transition %d: expected running=%t, got running=%t",
						i, expectedRunning, engine.IsRunning())
					return false
				}
			}

			return true
		},
		genStateTransitionSequence(),
	))

	properties.TestingRun(t)
}

// TestEngineStateAffectsRequestProcessing tests that engine state correctly affects request processing
// **Validates: Requirements 1.6**
func TestEngineStateAffectsRequestProcessing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine state affects requests: running engine processes requests, stopped engine rejects them", prop.ForAll(
		func(transitions []stateTransition, correlationId string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Apply transitions
			for _, transition := range transitions {
				switch transition {
				case transitionStart:
					engine.Start()
				case transitionStop:
					engine.Stop()
				}
			}

			// Create a health-check request
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          "/health-check",
				Payload:       nil,
			}
			payload, err := proto.Marshal(req)
			if err != nil {
				t.Logf("Failed to marshal request: %v", err)
				return false
			}

			// Invoke the engine
			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Invoke returned error: %v", err)
				return false
			}

			// Unmarshal response
			var resp invoke.Response
			if err := proto.Unmarshal(respBytes, &resp); err != nil {
				t.Logf("Failed to unmarshal response: %v", err)
				return false
			}

			// Check response based on engine state
			if engine.IsRunning() {
				// Running engine should process request successfully
				if resp.Error != "" {
					t.Logf("Running engine should process request without error, got %q", resp.Error)
					return false
				}
				if string(resp.Payload) != "OK" {
					t.Logf("Running engine should return 'OK' for health-check, got %q", string(resp.Payload))
					return false
				}
			} else {
				// Stopped engine should reject request
				if resp.Error != "engine is stopped" {
					t.Logf("Stopped engine should return 'engine is stopped' error, got %q", resp.Error)
					return false
				}
			}

			return true
		},
		genStateTransitionSequence(),
		genCorrelationId(),
	))

	properties.TestingRun(t)
}

// TestEngineMultipleStartStopCycles tests that engine handles multiple start/stop cycles correctly
// **Validates: Requirements 1.6**
func TestEngineMultipleStartStopCycles(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine multiple cycles: engine handles multiple start/stop cycles correctly", prop.ForAll(
		func(numCycles int, correlationId string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Perform multiple start/stop cycles
			for i := 0; i < numCycles; i++ {
				// Stop the engine
				engine.Stop()
				if engine.IsRunning() {
					t.Logf("Engine should be stopped after Stop() in cycle %d", i)
					return false
				}

				// Verify request is rejected when stopped
				req := &invoke.Request{
					CorrelationId: correlationId,
					Path:          "/health-check",
					Payload:       nil,
				}
				payload, _ := proto.Marshal(req)
				respBytes, _ := engine.Invoke(context.Background(), payload)
				var resp invoke.Response
				proto.Unmarshal(respBytes, &resp)
				if resp.Error != "engine is stopped" {
					t.Logf("Stopped engine should reject request in cycle %d, got error %q", i, resp.Error)
					return false
				}

				// Start the engine
				engine.Start()
				if !engine.IsRunning() {
					t.Logf("Engine should be running after Start() in cycle %d", i)
					return false
				}

				// Verify request is processed when running
				respBytes, _ = engine.Invoke(context.Background(), payload)
				proto.Unmarshal(respBytes, &resp)
				if resp.Error != "" {
					t.Logf("Running engine should process request in cycle %d, got error %q", i, resp.Error)
					return false
				}
				if string(resp.Payload) != "OK" {
					t.Logf("Running engine should return 'OK' in cycle %d, got %q", i, string(resp.Payload))
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10), // 1-10 cycles
		genCorrelationId(),
	))

	properties.TestingRun(t)
}

// TestEngineIdempotentStart tests that calling Start() multiple times is idempotent
// **Validates: Requirements 1.6**
func TestEngineIdempotentStart(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine idempotent Start: calling Start() multiple times keeps engine running", prop.ForAll(
		func(numStarts int, correlationId string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Call Start() multiple times
			for i := 0; i < numStarts; i++ {
				engine.Start()
				if !engine.IsRunning() {
					t.Logf("Engine should be running after Start() call %d", i)
					return false
				}
			}

			// Verify engine still processes requests
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          "/health-check",
				Payload:       nil,
			}
			payload, _ := proto.Marshal(req)
			respBytes, _ := engine.Invoke(context.Background(), payload)
			var resp invoke.Response
			proto.Unmarshal(respBytes, &resp)

			if resp.Error != "" {
				t.Logf("Engine should process request after multiple Start() calls, got error %q", resp.Error)
				return false
			}

			return true
		},
		gen.IntRange(1, 20), // 1-20 Start() calls
		genCorrelationId(),
	))

	properties.TestingRun(t)
}

// TestEngineIdempotentStop tests that calling Stop() multiple times is idempotent
// **Validates: Requirements 1.6**
func TestEngineIdempotentStop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Engine idempotent Stop: calling Stop() multiple times keeps engine stopped", prop.ForAll(
		func(numStops int, correlationId string) bool {
			// Create a new engine
			engine := invoke.NewEngine(nil, nil)

			// Call Stop() multiple times
			for i := 0; i < numStops; i++ {
				engine.Stop()
				if engine.IsRunning() {
					t.Logf("Engine should be stopped after Stop() call %d", i)
					return false
				}
			}

			// Verify engine rejects requests
			req := &invoke.Request{
				CorrelationId: correlationId,
				Path:          "/health-check",
				Payload:       nil,
			}
			payload, _ := proto.Marshal(req)
			respBytes, _ := engine.Invoke(context.Background(), payload)
			var resp invoke.Response
			proto.Unmarshal(respBytes, &resp)

			if resp.Error != "engine is stopped" {
				t.Logf("Engine should reject request after multiple Stop() calls, got error %q", resp.Error)
				return false
			}

			return true
		},
		gen.IntRange(1, 20), // 1-20 Stop() calls
		genCorrelationId(),
	))

	properties.TestingRun(t)
}

// Ensure dynamic package is imported for NewEngine
var _ = dynamic.Option(nil)
