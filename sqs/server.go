package sqs

import "github.com/aws/aws-lambda-go/lambda"

var engine *Engine

// Serve runs the SQS Engine handler for AWS Lambda SQS events.
func Serve(opts ...ServeOption) {
	engine = NewEngine(opts...)
	lambda.Start(engine.Invoke)
}

func Close() {
	if engine != nil {
		engine.Stop()
	}
}
