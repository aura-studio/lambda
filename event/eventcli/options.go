package eventcli

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/mohae/deepcopy"
)

// LambdaClient Lambda 客户端接口
// Requirement 4.6: THE Event_Client SHALL support configurable Lambda client injection for testing
type LambdaClient interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput,
		optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// Options 客户端配置选项
type Options struct {
	LambdaClient LambdaClient
	FunctionName string
}

// Option 配置选项接口
type Option interface {
	Apply(o *Options)
}

// OptionFunc 配置选项函数类型
type OptionFunc func(*Options)

// Apply 实现 Option 接口
func (f OptionFunc) Apply(o *Options) { f(o) }

var defaultOptions = &Options{}

// NewOptions 创建新的配置选项
func NewOptions(opts ...Option) *Options {
	o := deepcopy.Copy(defaultOptions).(*Options)
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(o)
		}
	}
	return o
}

// WithLambdaClient 设置 Lambda 客户端
// Requirement 4.6: THE Event_Client SHALL support configurable Lambda client injection for testing
func WithLambdaClient(client LambdaClient) Option {
	return OptionFunc(func(o *Options) {
		o.LambdaClient = client
	})
}

// WithFunctionName 设置目标 Lambda 函数名称
// Requirement 4.5: THE Event_Client SHALL support configurable target function name
func WithFunctionName(name string) Option {
	return OptionFunc(func(o *Options) {
		o.FunctionName = name
	})
}
