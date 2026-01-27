package client

import (
	"net/http"
	"time"

	"github.com/mohae/deepcopy"
)

// HTTPClient HTTP 客户端接口
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Options 客户端配置选项
type Options struct {
	HTTPClient     HTTPClient
	BaseURL        string
	DefaultTimeout time.Duration
	Headers        map[string]string
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
	HTTPClient:     http.DefaultClient,
	DefaultTimeout: 30 * time.Second,
	Headers:        make(map[string]string),
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

// WithHTTPClient 设置 HTTP 客户端
func WithHTTPClient(client HTTPClient) Option {
	return OptionFunc(func(o *Options) {
		o.HTTPClient = client
	})
}

// WithBaseURL 设置基础 URL
func WithBaseURL(url string) Option {
	return OptionFunc(func(o *Options) {
		o.BaseURL = url
	})
}

// WithDefaultTimeout 设置默认超时时间
func WithDefaultTimeout(timeout time.Duration) Option {
	return OptionFunc(func(o *Options) {
		o.DefaultTimeout = timeout
	})
}

// WithHeaders 设置默认请求头
func WithHeaders(headers map[string]string) Option {
	return OptionFunc(func(o *Options) {
		o.Headers = headers
	})
}

// WithHeader 添加单个请求头
func WithHeader(key, value string) Option {
	return OptionFunc(func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	})
}
