package websocket

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/lambda"
)

var engine *Engine

// Serve runs the WebSocket Engine handler for API Gateway WebSocket proxy events.
func Serve(wsOpts []Option, dynamicOpts []dynamic.Option) {
	engine = NewEngine(wsOpts, dynamicOpts)
	lambda.Start(engine.Invoke)
}

func Close() {
}
