package invokecli

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/mohae/deepcopy"
)

// LambdaClient Lambda 客户端接口
type LambdaClient interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput,
		optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// Options 客户端配置选项
type Options struct {
	LambdaClient   LambdaClient
	FunctionName   string
	DefaultTimeout time.Duration
}

// Option 配置选项接口
type Option interface {
	Apply(o *Options)
}

// OptionFunc 配置选项函数类型
type OptionFunc func(*Options)

// Apply 实现 Option 接口
func (f OptionFunc) Apply(o *Options) { f(o) }

var defaultOptions = &Options{
	DefaultTimeout: 30 * time.Second,
}

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
func WithLambdaClient(client LambdaClient) Option {
	return OptionFunc(func(o *Options) {
		o.LambdaClient = client
	})
}

// WithFunctionName 设置目标 Lambda 函数名称
func WithFunctionName(name string) Option {
	return OptionFunc(func(o *Options) {
		o.FunctionName = name
	})
}

// WithDefaultTimeout 设置默认调用超时时间
func WithDefaultTimeout(timeout time.Duration) Option {
	return OptionFunc(func(o *Options) {
		o.DefaultTimeout = timeout
	})
}
