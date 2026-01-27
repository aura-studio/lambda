package event

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
	"google.golang.org/protobuf/proto"
)

// Engine is the core engine for handling Lambda Invoke Event mode requests.
// It processes protobuf-encoded requests containing batch items and dispatches
// them to registered handlers via the Router.
//
// Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7
type Engine struct {
	*Options
	*dynamic.Dynamic
	r       *Router
	running atomic.Int32
}

// NewEngine creates a new Engine instance with the given options.
// The engine starts in running state by default.
//
// Validates: Requirement 1.5
func NewEngine(eventOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(eventOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
	}
	e.running.Store(1)
	e.InstallHandlers()
	return e
}

// Start starts the engine, allowing it to accept new requests.
//
// Validates: Requirement 1.5
func (e *Engine) Start() {
	e.running.Store(1)
}

// Stop stops the engine, causing it to reject new requests.
//
// Validates: Requirement 1.5
func (e *Engine) Stop() {
	e.running.Store(0)
}

// IsRunning returns true if the engine is currently running.
//
// Validates: Requirement 1.5
func (e *Engine) IsRunning() bool {
	return e.running.Load() == 1
}

// Invoke processes a Lambda invocation payload containing batch items.
// It deserializes the protobuf payload and processes each item according
// to the configured RunMode.
//
// Returns:
//   - nil if all items processed successfully (or partial mode with some failures)
//   - error if processing failed according to RunMode rules
//
// Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.6, 1.7, 7.1, 7.2, 7.3, 7.4
func (e *Engine) Invoke(ctx context.Context, payload []byte) error {
	_ = ctx // context reserved for future use (timeouts, cancellation)

	// Check if engine is running (Requirement 1.4)
	if !e.IsRunning() {
		return fmt.Errorf("event: engine is stopped")
	}

	// Deserialize protobuf payload (Requirement 1.1, 1.2)
	var request Request
	if err := proto.Unmarshal(payload, &request); err != nil {
		if e.DebugMode {
			log.Printf("[Event] Unmarshal payload error: %v", err)
		}
		return fmt.Errorf("event: invalid protobuf payload: %w", err)
	}

	// Process batch items (Requirements 1.6, 1.7)
	return e.processItems(request.Items)
}

// processItems processes all items in the batch according to RunMode.
//
// Validates: Requirements 1.7, 7.1, 7.2, 7.3, 7.4
func (e *Engine) processItems(items []*Item) error {
	var lastErr error

	for i, item := range items {
		// Check if engine is still running
		if !e.IsRunning() {
			if e.DebugMode {
				log.Printf("[Event] Engine stopped during processing item %d", i)
			}
			return fmt.Errorf("event: engine stopped during processing")
		}

		err := e.processItem(item)
		if err != nil {
			if e.DebugMode {
				log.Printf("[Event] Error processing item %d (path=%s): %v", i, item.GetPath(), err)
			}

			// Handle error according to RunMode (Requirements 7.1, 7.2, 7.3, 7.4)
			switch e.RunMode {
			case RunModeStrict:
				// Stop processing remaining items and return error (Requirement 7.1)
				return err

			case RunModePartial:
				// Continue processing remaining items, record failure (Requirement 7.2)
				// In partial mode, we continue but don't return error at the end
				lastErr = err
				continue

			case RunModeBatch:
				// Return error immediately, fail entire batch (Requirement 7.3)
				return err

			case RunModeReentrant:
				// Record error but continue processing all items (Requirement 7.4)
				lastErr = err
				continue

			default:
				// Default to batch behavior
				return err
			}
		}
	}

	// Return behavior based on RunMode
	switch e.RunMode {
	case RunModePartial:
		// Partial mode: return nil even if there were failures (Requirement 7.2)
		return nil
	case RunModeReentrant:
		// Reentrant mode: return the last error if any (Requirement 7.4)
		return lastErr
	default:
		return nil
	}
}

// processItem processes a single item by dispatching it through the router.
//
// Validates: Requirements 1.3, 8.1
func (e *Engine) processItem(item *Item) (err error) {
	// Create context for this item
	c := &Context{
		Engine:    e,
		RawPath:   item.GetPath(),
		Path:      item.GetPath(),
		Request:   string(item.GetPayload()),
		DebugMode: e.DebugMode,
	}

	if e.DebugMode {
		log.Printf("[Event] Request: %s %s", c.Path, c.Request)
	}

	// Dispatch with panic recovery (Requirement 8.1)
	func() {
		defer func() {
			if r := recover(); r != nil {
				c.Err = fmt.Errorf("panic: %v", r)
			}
		}()
		e.r.Dispatch(c)
	}()

	if e.DebugMode && c.Err != nil {
		log.Printf("[Event] Error: %s %v", c.Path, c.Err)
	}

	// Return context error if any
	// Note: No response payload is returned (Requirement 1.3)
	return c.Err
}
