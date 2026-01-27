package reqresp

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/lambda"
)

// engine 是全局引擎变量
var engine *Engine

// Serve 启动 Lambda 处理器
func Serve(reqrespOpts []Option, dynamicOpts []dynamic.Option) {
	engine = NewEngine(reqrespOpts, dynamicOpts)
	lambda.Start(engine.Invoke)
}

// Close 优雅关闭引擎
func Close() {
	if engine != nil {
		engine.Stop()
	}
}
