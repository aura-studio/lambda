package sqs

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/lambda"
)

var engine *Engine

// Serve runs the SQS Engine handler for AWS Lambda SQS events.
func Serve(sqsOpts []Option, dynamicOpts []dynamic.Option) {
	engine = NewEngine(sqsOpts, dynamicOpts)
	lambda.Start(engine.Invoke)
}

func Close() {
	if engine != nil {
		engine.Stop()
	}
}
