package event

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/lambda"
)

// engine is the global engine variable for the Event module.
var engine *Engine

// Serve creates an Event_Engine and starts the Lambda handler.
// It registers the engine's Invoke method as the Lambda handler.
//
// Validates: Requirements 9.1, 9.3
func Serve(eventOpts []Option, dynamicOpts []dynamic.Option) {
	engine = NewEngine(eventOpts, dynamicOpts)
	lambda.Start(engine.Invoke)
}

// Close stops the Event_Engine gracefully.
//
// Validates: Requirement 9.2
func Close() {
	if engine != nil {
		engine.Stop()
	}
}
