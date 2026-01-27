package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/aura-studio/dynamic"
	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// Feature: event-lambda-handler
// Property 14: RunMode Strict Behavior
// *For any* batch of items where item K fails (K < N), in strict mode the Engine
// SHALL stop processing at item K and return an error, leaving items K+1 to N unprocessed.
//
// **Validates: Requirements 7.1**

// Property 15: RunMode Partial Behavior
// *For any* batch of items where item K fails, in partial mode the Engine SHALL
// continue processing all remaining items and record the failure.
//
// **Validates: Requirements 7.2**

// Property 16: RunMode Batch Behavior
// *For any* batch of items where any item fails, in batch mode the Engine SHALL
// return an error immediately (entire batch fails).
//
// **Validates: Requirements 7.3**

// Property 17: RunMode Reentrant Behavior
// *For any* batch of items where item K fails, in reentrant mode the Engine SHALL
// record the error but continue processing all items to completion.
//
// **Validates: Requirements 7.4**

// runModeTestTracker tracks which items were processed during a test
type runModeTestTracker struct {
	processedIndices []int
	failAtIndex      int
}

// batchWithFailure represents a batch configuration with a failure point
type batchWithFailure struct {
	batchSize int
	failIndex int
}

// genBatchWithFailure generates a batch size N (2-10) and a failure index K (0 to N-1)
// Uses a simpler approach that avoids FlatMap issues
func genBatchWithFailure() gopter.Gen {
	return gopter.CombineGens(
		gen.IntRange(2, 10),
		gen.IntRange(0, 9),
	).Map(func(values []interface{}) batchWithFailure {
		batchSize := values[0].(int)
		// Ensure failIndex is within valid range [0, batchSize-1]
		failIndex := values[1].(int) % batchSize
		return batchWithFailure{batchSize: batchSize, failIndex: failIndex}
	})
}

// createTestPackageForRunMode creates a test package that tracks processed items
// and fails at a specific index
func createTestPackageForRunMode(pkgName, version string, tracker *runModeTestTracker) {
	dynamic.RegisterPackage(pkgName, version, &mockTunnel{
		invoke: func(route, req string) string {
			// Extract index from route (e.g., "/item-0" -> 0)
			var idx int
			fmt.Sscanf(route, "/item-%d", &idx)
			tracker.processedIndices = append(tracker.processedIndices, idx)

			if idx == tracker.failAtIndex {
				panic(fmt.Sprintf("intentional failure at item %d", idx))
			}
			return "OK"
		},
	})
}

// createBatchRequest creates a batch request with N items
func createBatchRequest(pkgName, version string, n int) ([]byte, error) {
	items := make([]*event.Item, n)
	for i := 0; i < n; i++ {
		items[i] = &event.Item{
			Path:    fmt.Sprintf("/api/%s/%s/item-%d", pkgName, version, i),
			Payload: []byte(`{}`),
		}
	}
	req := &event.Request{Items: items}
	return proto.Marshal(req)
}

// TestRunModeStrictBehavior tests Property 14: RunMode Strict Behavior
// For any batch of items where item K fails (K < N), in strict mode the Engine
// SHALL stop processing at item K and return an error, leaving items K+1 to N unprocessed.
func TestRunModeStrictBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("strict mode stops at failure and leaves remaining items unprocessed", prop.ForAll(
		func(params batchWithFailure) bool {
			batchSize := params.batchSize
			failIndex := params.failIndex

			// Create unique package name for this test iteration
			pkgName := fmt.Sprintf("strict-test-%d-%d", batchSize, failIndex)
			version := "v1"

			tracker := &runModeTestTracker{
				processedIndices: make([]int, 0),
				failAtIndex:      failIndex,
			}
			createTestPackageForRunMode(pkgName, version, tracker)

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeStrict),
			}, nil)

			payload, err := createBatchRequest(pkgName, version, batchSize)
			if err != nil {
				t.Logf("Failed to create batch request: %v", err)
				return false
			}

			// Invoke should return an error
			err = e.Invoke(context.Background(), payload)
			if err == nil {
				t.Logf("Expected error in strict mode when item %d fails", failIndex)
				return false
			}

			// Verify items 0 to K were processed (K is the fail index)
			expectedProcessed := failIndex + 1
			if len(tracker.processedIndices) != expectedProcessed {
				t.Logf("Expected %d items processed, got %d (processed: %v)",
					expectedProcessed, len(tracker.processedIndices), tracker.processedIndices)
				return false
			}

			// Verify items were processed in order up to and including the failure
			for i := 0; i <= failIndex; i++ {
				if tracker.processedIndices[i] != i {
					t.Logf("Expected item %d at position %d, got %d",
						i, i, tracker.processedIndices[i])
					return false
				}
			}

			return true
		},
		genBatchWithFailure(),
	))

	properties.TestingRun(t)
}

// TestRunModePartialBehavior tests Property 15: RunMode Partial Behavior
// For any batch of items where item K fails, in partial mode the Engine SHALL
// continue processing all remaining items and record the failure.
func TestRunModePartialBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("partial mode continues processing all items after failure", prop.ForAll(
		func(params batchWithFailure) bool {
			batchSize := params.batchSize
			failIndex := params.failIndex

			// Create unique package name for this test iteration
			pkgName := fmt.Sprintf("partial-test-%d-%d", batchSize, failIndex)
			version := "v1"

			tracker := &runModeTestTracker{
				processedIndices: make([]int, 0),
				failAtIndex:      failIndex,
			}
			createTestPackageForRunMode(pkgName, version, tracker)

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModePartial),
			}, nil)

			payload, err := createBatchRequest(pkgName, version, batchSize)
			if err != nil {
				t.Logf("Failed to create batch request: %v", err)
				return false
			}

			// Invoke should return nil (partial mode doesn't return error)
			err = e.Invoke(context.Background(), payload)
			if err != nil {
				t.Logf("Partial mode should return nil, got error: %v", err)
				return false
			}

			// Verify ALL items were processed
			if len(tracker.processedIndices) != batchSize {
				t.Logf("Expected %d items processed, got %d (processed: %v)",
					batchSize, len(tracker.processedIndices), tracker.processedIndices)
				return false
			}

			// Verify items were processed in order
			for i := 0; i < batchSize; i++ {
				if tracker.processedIndices[i] != i {
					t.Logf("Expected item %d at position %d, got %d",
						i, i, tracker.processedIndices[i])
					return false
				}
			}

			return true
		},
		genBatchWithFailure(),
	))

	properties.TestingRun(t)
}

// TestRunModeBatchBehavior tests Property 16: RunMode Batch Behavior
// For any batch of items where any item fails, in batch mode the Engine SHALL
// return an error immediately (entire batch fails).
func TestRunModeBatchBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("batch mode returns error immediately on first failure", prop.ForAll(
		func(params batchWithFailure) bool {
			batchSize := params.batchSize
			failIndex := params.failIndex

			// Create unique package name for this test iteration
			pkgName := fmt.Sprintf("batch-test-%d-%d", batchSize, failIndex)
			version := "v1"

			tracker := &runModeTestTracker{
				processedIndices: make([]int, 0),
				failAtIndex:      failIndex,
			}
			createTestPackageForRunMode(pkgName, version, tracker)

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeBatch),
			}, nil)

			payload, err := createBatchRequest(pkgName, version, batchSize)
			if err != nil {
				t.Logf("Failed to create batch request: %v", err)
				return false
			}

			// Invoke should return an error immediately
			err = e.Invoke(context.Background(), payload)
			if err == nil {
				t.Logf("Expected error in batch mode when item %d fails", failIndex)
				return false
			}

			// Verify processing stopped at the failure point
			// Items 0 to K should be processed (K is the fail index)
			expectedProcessed := failIndex + 1
			if len(tracker.processedIndices) != expectedProcessed {
				t.Logf("Expected %d items processed, got %d (processed: %v)",
					expectedProcessed, len(tracker.processedIndices), tracker.processedIndices)
				return false
			}

			// Verify items were processed in order up to and including the failure
			for i := 0; i <= failIndex; i++ {
				if tracker.processedIndices[i] != i {
					t.Logf("Expected item %d at position %d, got %d",
						i, i, tracker.processedIndices[i])
					return false
				}
			}

			return true
		},
		genBatchWithFailure(),
	))

	properties.TestingRun(t)
}

// TestRunModeReentrantBehavior tests Property 17: RunMode Reentrant Behavior
// For any batch of items where item K fails, in reentrant mode the Engine SHALL
// record the error but continue processing all items to completion.
func TestRunModeReentrantBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("reentrant mode continues processing all items and returns error at end", prop.ForAll(
		func(params batchWithFailure) bool {
			batchSize := params.batchSize
			failIndex := params.failIndex

			// Create unique package name for this test iteration
			pkgName := fmt.Sprintf("reentrant-test-%d-%d", batchSize, failIndex)
			version := "v1"

			tracker := &runModeTestTracker{
				processedIndices: make([]int, 0),
				failAtIndex:      failIndex,
			}
			createTestPackageForRunMode(pkgName, version, tracker)

			e := event.NewEngine([]event.Option{
				event.WithRunMode(event.RunModeReentrant),
			}, nil)

			payload, err := createBatchRequest(pkgName, version, batchSize)
			if err != nil {
				t.Logf("Failed to create batch request: %v", err)
				return false
			}

			// Invoke should return an error (the recorded error)
			err = e.Invoke(context.Background(), payload)
			if err == nil {
				t.Logf("Expected error in reentrant mode when item %d fails", failIndex)
				return false
			}

			// Verify ALL items were processed
			if len(tracker.processedIndices) != batchSize {
				t.Logf("Expected %d items processed, got %d (processed: %v)",
					batchSize, len(tracker.processedIndices), tracker.processedIndices)
				return false
			}

			// Verify items were processed in order
			for i := 0; i < batchSize; i++ {
				if tracker.processedIndices[i] != i {
					t.Logf("Expected item %d at position %d, got %d",
						i, i, tracker.processedIndices[i])
					return false
				}
			}

			return true
		},
		genBatchWithFailure(),
	))

	properties.TestingRun(t)
}

// ============================================================================
// Additional Unit Tests for RunMode edge cases
// ============================================================================

// TestRunModeStrictWithMultipleFailures tests strict mode with multiple potential failures
func TestRunModeStrictWithMultipleFailures(t *testing.T) {
	pkgName := "strict-multi"
	version := "v1"
	processedIndices := make([]int, 0)

	// Register package that fails at indices 1 and 3
	dynamic.RegisterPackage(pkgName, version, &mockTunnel{
		invoke: func(route, req string) string {
			var idx int
			fmt.Sscanf(route, "/item-%d", &idx)
			processedIndices = append(processedIndices, idx)

			if idx == 1 || idx == 3 {
				panic(fmt.Sprintf("failure at item %d", idx))
			}
			return "OK"
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModeStrict),
	}, nil)

	payload, _ := createBatchRequest(pkgName, version, 5)
	err := e.Invoke(context.Background(), payload)

	// Should stop at first failure (index 1)
	if err == nil {
		t.Error("Expected error in strict mode")
	}

	// Should have processed items 0 and 1 only
	if len(processedIndices) != 2 {
		t.Errorf("Expected 2 items processed, got %d: %v", len(processedIndices), processedIndices)
	}
}

// TestRunModePartialWithMultipleFailures tests partial mode with multiple failures
func TestRunModePartialWithMultipleFailures(t *testing.T) {
	pkgName := "partial-multi"
	version := "v1"
	processedIndices := make([]int, 0)

	// Register package that fails at indices 1 and 3
	dynamic.RegisterPackage(pkgName, version, &mockTunnel{
		invoke: func(route, req string) string {
			var idx int
			fmt.Sscanf(route, "/item-%d", &idx)
			processedIndices = append(processedIndices, idx)

			if idx == 1 || idx == 3 {
				panic(fmt.Sprintf("failure at item %d", idx))
			}
			return "OK"
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModePartial),
	}, nil)

	payload, _ := createBatchRequest(pkgName, version, 5)
	err := e.Invoke(context.Background(), payload)

	// Partial mode should return nil
	if err != nil {
		t.Errorf("Partial mode should return nil, got: %v", err)
	}

	// Should have processed all 5 items
	if len(processedIndices) != 5 {
		t.Errorf("Expected 5 items processed, got %d: %v", len(processedIndices), processedIndices)
	}
}

// TestRunModeReentrantWithMultipleFailures tests reentrant mode with multiple failures
func TestRunModeReentrantWithMultipleFailures(t *testing.T) {
	pkgName := "reentrant-multi"
	version := "v1"
	processedIndices := make([]int, 0)

	// Register package that fails at indices 1 and 3
	dynamic.RegisterPackage(pkgName, version, &mockTunnel{
		invoke: func(route, req string) string {
			var idx int
			fmt.Sscanf(route, "/item-%d", &idx)
			processedIndices = append(processedIndices, idx)

			if idx == 1 || idx == 3 {
				panic(fmt.Sprintf("failure at item %d", idx))
			}
			return "OK"
		},
	})

	e := event.NewEngine([]event.Option{
		event.WithRunMode(event.RunModeReentrant),
	}, nil)

	payload, _ := createBatchRequest(pkgName, version, 5)
	err := e.Invoke(context.Background(), payload)

	// Reentrant mode should return the last error
	if err == nil {
		t.Error("Expected error in reentrant mode")
	}

	// Error should be from the last failure (index 3)
	if err.Error() != "panic: failure at item 3" {
		t.Errorf("Expected error from item 3, got: %v", err)
	}

	// Should have processed all 5 items
	if len(processedIndices) != 5 {
		t.Errorf("Expected 5 items processed, got %d: %v", len(processedIndices), processedIndices)
	}
}

// TestRunModeNoFailure tests all run modes with no failures
func TestRunModeNoFailure(t *testing.T) {
	modes := []event.RunMode{
		event.RunModeStrict,
		event.RunModePartial,
		event.RunModeBatch,
		event.RunModeReentrant,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			pkgName := fmt.Sprintf("nofail-%s", mode)
			version := "v1"
			processedCount := 0

			dynamic.RegisterPackage(pkgName, version, &mockTunnel{
				invoke: func(route, req string) string {
					processedCount++
					return "OK"
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(mode),
			}, nil)

			payload, _ := createBatchRequest(pkgName, version, 5)
			err := e.Invoke(context.Background(), payload)

			// No failures, so no error should be returned
			if err != nil {
				t.Errorf("Expected no error with no failures, got: %v", err)
			}

			// All items should be processed
			if processedCount != 5 {
				t.Errorf("Expected 5 items processed, got %d", processedCount)
			}
		})
	}
}

// TestRunModeFirstItemFailure tests all run modes when the first item fails
func TestRunModeFirstItemFailure(t *testing.T) {
	testCases := []struct {
		mode            event.RunMode
		expectError     bool
		expectedCount   int
		description     string
	}{
		{event.RunModeStrict, true, 1, "strict stops at first item"},
		{event.RunModePartial, false, 5, "partial continues all items"},
		{event.RunModeBatch, true, 1, "batch fails immediately"},
		{event.RunModeReentrant, true, 5, "reentrant continues all items"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.mode), func(t *testing.T) {
			pkgName := fmt.Sprintf("firstfail-%s", tc.mode)
			version := "v1"
			processedCount := 0

			dynamic.RegisterPackage(pkgName, version, &mockTunnel{
				invoke: func(route, req string) string {
					processedCount++
					var idx int
					fmt.Sscanf(route, "/item-%d", &idx)
					if idx == 0 {
						panic("first item failure")
					}
					return "OK"
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(tc.mode),
			}, nil)

			payload, _ := createBatchRequest(pkgName, version, 5)
			err := e.Invoke(context.Background(), payload)

			if tc.expectError && err == nil {
				t.Errorf("%s: expected error but got nil", tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tc.description, err)
			}

			if processedCount != tc.expectedCount {
				t.Errorf("%s: expected %d items processed, got %d",
					tc.description, tc.expectedCount, processedCount)
			}
		})
	}
}

// TestRunModeLastItemFailure tests all run modes when the last item fails
func TestRunModeLastItemFailure(t *testing.T) {
	testCases := []struct {
		mode            event.RunMode
		expectError     bool
		expectedCount   int
		description     string
	}{
		{event.RunModeStrict, true, 5, "strict processes all then fails"},
		{event.RunModePartial, false, 5, "partial continues all items"},
		{event.RunModeBatch, true, 5, "batch fails at last item"},
		{event.RunModeReentrant, true, 5, "reentrant continues all items"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.mode), func(t *testing.T) {
			pkgName := fmt.Sprintf("lastfail-%s", tc.mode)
			version := "v1"
			processedCount := 0

			dynamic.RegisterPackage(pkgName, version, &mockTunnel{
				invoke: func(route, req string) string {
					processedCount++
					var idx int
					fmt.Sscanf(route, "/item-%d", &idx)
					if idx == 4 { // Last item (index 4 of 5 items)
						panic("last item failure")
					}
					return "OK"
				},
			})

			e := event.NewEngine([]event.Option{
				event.WithRunMode(tc.mode),
			}, nil)

			payload, _ := createBatchRequest(pkgName, version, 5)
			err := e.Invoke(context.Background(), payload)

			if tc.expectError && err == nil {
				t.Errorf("%s: expected error but got nil", tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tc.description, err)
			}

			if processedCount != tc.expectedCount {
				t.Errorf("%s: expected %d items processed, got %d",
					tc.description, tc.expectedCount, processedCount)
			}
		})
	}
}
