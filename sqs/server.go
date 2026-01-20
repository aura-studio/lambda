package sqs

import "github.com/aws/aws-lambda-go/lambda"

var engine *Engine

// Serve runs the SQS Engine handler for AWS Lambda SQS events.
func Serve(opts ...ServeOption) {
	engine = NewEngine(opts...)
	engine.SetInvokeFunc(engine.invokeFunc)
	lambda.Start(engine)
}

func Close() {
	if engine != nil {
		engine.Stop()
	}
}
