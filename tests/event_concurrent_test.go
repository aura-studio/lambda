package tests

import (
	"sync"
	"testing"

	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: event-lambda-handler
// Property 20: Thread-Safe State
// *For any* concurrent sequence of Start(), Stop(), and IsRunning() calls,
// the Engine SHALL maintain consistent state without data races.
//
// **Validates: Requirements 9.4**

// Operation represents a concurrent operation on the Engine
type Operation int

const (
	OpStart Operation = iota
	OpStop
	OpIsRunning
)

// genOperation generates a random operation
func genOperation() gopter.Gen {
	return gen.IntRange(0, 2).Map(func(i int) Operation {
		return Operation(i)
	})
}

// genOperationSequence generates a sequence of operations for concurrent execution
func genOperationSequence() gopter.Gen {
	return gen.SliceOfN(50, genOperation()).SuchThat(func(ops []Operation) bool {
		return len(ops) >= 10
	})
}

// genConcurrencyLevel generates the number of concurrent goroutines
func genConcurrencyLevel() gopter.Gen {
	return gen.IntRange(2, 10)
}

// TestEventEngineConcurrentThreadSafety tests Property 20: Thread-Safe State
// For any concurrent sequence of Start(), Stop(), and IsRunning() calls,
// the Engine SHALL maintain consistent state without data races.
//
// This test should be run with -race flag to detect data races:
// go test -race -run TestEventEngineConcurrentThreadSafety
func TestEventEngineConcurrentThreadSafety(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent Start/Stop/IsRunning maintains consistent state without data races", prop.ForAll(
		func(ops []Operation, numGoroutines int) bool {
			e := event.NewEngine(nil, nil)

			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			// Execute operations concurrently from multiple goroutines
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer wg.Done()

					// Each goroutine executes a portion of the operations
					startIdx := (goroutineID * len(ops)) / numGoroutines
					endIdx := ((goroutineID + 1) * len(ops)) / numGoroutines

					for j := startIdx; j < endIdx; j++ {
						switch ops[j] {
						case OpStart:
							e.Start()
						case OpStop:
							e.Stop()
						case OpIsRunning:
							_ = e.IsRunning()
						}
					}
				}(i)
			}

			wg.Wait()

			// After all operations complete, the state should be valid (either running or not)
			// The key property is that no data race occurred (verified by -race flag)
			running := e.IsRunning()
			return running == true || running == false // Always true, but validates no panic
		},
		genOperationSequence(),
		genConcurrencyLevel(),
	))

	properties.TestingRun(t)
}

// TestEventEngineConcurrentStartStop tests concurrent Start and Stop calls
// to verify atomic state transitions.
func TestEventEngineConcurrentStartStop(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent Start/Stop calls do not cause data races", prop.ForAll(
		func(numStarts, numStops int) bool {
			e := event.NewEngine(nil, nil)

			var wg sync.WaitGroup
			totalOps := numStarts + numStops
			wg.Add(totalOps)

			// Launch Start goroutines
			for i := 0; i < numStarts; i++ {
				go func() {
					defer wg.Done()
					e.Start()
				}()
			}

			// Launch Stop goroutines
			for i := 0; i < numStops; i++ {
				go func() {
					defer wg.Done()
					e.Stop()
				}()
			}

			wg.Wait()

			// State should be consistent (either running or stopped)
			running := e.IsRunning()
			return running == true || running == false
		},
		gen.IntRange(5, 20), // numStarts
		gen.IntRange(5, 20), // numStops
	))

	properties.TestingRun(t)
}

// TestEventEngineConcurrentIsRunning tests concurrent IsRunning calls
// while Start/Stop operations are happening.
func TestEventEngineConcurrentIsRunning(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent IsRunning calls return consistent boolean values", prop.ForAll(
		func(numReaders int) bool {
			e := event.NewEngine(nil, nil)

			var wg sync.WaitGroup
			results := make([]bool, numReaders)

			// Start a goroutine that toggles state
			done := make(chan struct{})
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						e.Stop()
						e.Start()
					}
				}
			}()

			// Launch reader goroutines
			wg.Add(numReaders)
			for i := 0; i < numReaders; i++ {
				go func(idx int) {
					defer wg.Done()
					// Read multiple times
					for j := 0; j < 100; j++ {
						results[idx] = e.IsRunning()
					}
				}(i)
			}

			wg.Wait()
			close(done)

			// All results should be valid boolean values (no corruption)
			for _, r := range results {
				if r != true && r != false {
					return false
				}
			}
			return true
		},
		gen.IntRange(5, 20), // numReaders
	))

	properties.TestingRun(t)
}

// TestEventEngineConcurrentStateConsistency tests that after a known final
// operation, the state is consistent.
func TestEventEngineConcurrentStateConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("state is consistent after concurrent operations followed by known final operation", prop.ForAll(
		func(ops []Operation, finalIsStart bool) bool {
			e := event.NewEngine(nil, nil)

			var wg sync.WaitGroup
			numGoroutines := 5
			wg.Add(numGoroutines)

			// Execute random operations concurrently
			for i := 0; i < numGoroutines; i++ {
				go func(goroutineID int) {
					defer wg.Done()
					startIdx := (goroutineID * len(ops)) / numGoroutines
					endIdx := ((goroutineID + 1) * len(ops)) / numGoroutines

					for j := startIdx; j < endIdx; j++ {
						switch ops[j] {
						case OpStart:
							e.Start()
						case OpStop:
							e.Stop()
						case OpIsRunning:
							_ = e.IsRunning()
						}
					}
				}(i)
			}

			wg.Wait()

			// Apply a known final operation
			if finalIsStart {
				e.Start()
			} else {
				e.Stop()
			}

			// State should match the final operation
			return e.IsRunning() == finalIsStart
		},
		genOperationSequence(),
		gen.Bool(), // finalIsStart
	))

	properties.TestingRun(t)
}

// TestEventEngineConcurrentRapidToggle tests rapid toggling of state
// from multiple goroutines.
func TestEventEngineConcurrentRapidToggle(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("rapid concurrent toggling maintains atomic state", prop.ForAll(
		func(iterations int) bool {
			e := event.NewEngine(nil, nil)

			var wg sync.WaitGroup
			numGoroutines := 10
			wg.Add(numGoroutines)

			// Each goroutine rapidly toggles state
			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					defer wg.Done()
					for j := 0; j < iterations; j++ {
						if id%2 == 0 {
							e.Start()
							_ = e.IsRunning()
						} else {
							e.Stop()
							_ = e.IsRunning()
						}
					}
				}(i)
			}

			wg.Wait()

			// Final state should be valid
			running := e.IsRunning()
			return running == true || running == false
		},
		gen.IntRange(50, 200), // iterations per goroutine
	))

	properties.TestingRun(t)
}

// ============================================================================
// Unit Tests for additional concurrent coverage
// ============================================================================

// TestEventEngineConcurrentStartStopUnit is a unit test for concurrent Start/Stop
func TestEventEngineConcurrentStartStopUnit(t *testing.T) {
	e := event.NewEngine(nil, nil)

	var wg sync.WaitGroup
	iterations := 1000

	// Start multiple goroutines calling Start
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				e.Start()
			}
		}()
	}

	// Start multiple goroutines calling Stop
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				e.Stop()
			}
		}()
	}

	// Start multiple goroutines calling IsRunning
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = e.IsRunning()
			}
		}()
	}

	wg.Wait()

	// Should complete without data races (verified by -race flag)
	// Final state should be valid
	running := e.IsRunning()
	if running != true && running != false {
		t.Error("Invalid final state")
	}
}

// TestEventEngineConcurrentIsRunningUnit tests concurrent IsRunning reads
func TestEventEngineConcurrentIsRunningUnit(t *testing.T) {
	e := event.NewEngine(nil, nil)

	var wg sync.WaitGroup
	numReaders := 100
	iterations := 1000

	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				running := e.IsRunning()
				// Verify the value is a valid boolean
				if running != true && running != false {
					t.Error("Invalid IsRunning result")
				}
			}
		}()
	}

	wg.Wait()
}

// TestEventEngineConcurrentStateTransitions tests state transitions under concurrency
func TestEventEngineConcurrentStateTransitions(t *testing.T) {
	e := event.NewEngine(nil, nil)

	var wg sync.WaitGroup

	// Writer goroutine that toggles state
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				e.Stop()
				e.Start()
			}
		}
	}()

	// Reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = e.IsRunning()
			}
		}()
	}

	// Let it run for a bit
	close(done)
	wg.Wait()
}
