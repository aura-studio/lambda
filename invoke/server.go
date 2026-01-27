package invoke

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/lambda"
)

// engine 是全局引擎变量
var engine *Engine

// Serve 启动 Lambda 处理器
// Requirements: 6.1 - 提供 Serve 函数用于启动 Lambda 处理器
// Requirements: 6.2 - 创建 Invoke_Engine 并注册到 Lambda 运行时
func Serve(invokeOpts []Option, dynamicOpts []dynamic.Option) {
	engine = NewEngine(invokeOpts, dynamicOpts)
	lambda.Start(engine.Invoke)
}

// Close 优雅关闭引擎
// Requirements: 6.3 - 提供 Close 函数用于优雅关闭引擎
func Close() {
	if engine != nil {
		engine.Stop()
	}
}
