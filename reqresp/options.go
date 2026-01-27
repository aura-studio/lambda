package reqresp

import "github.com/mohae/deepcopy"

// Option is the interface for configuring Options
type Option interface {
	Apply(o *Options)
}

// OptionFunc is a function that implements the Option interface
type OptionFunc func(*Options)

// Apply implements the Option interface
func (f OptionFunc) Apply(o *Options) { f(o) }

// Options holds the configuration for the reqresp engine
type Options struct {
	StaticLinkMap map[string]string // 静态路径映射
	PrefixLinkMap map[string]string // 前缀路径映射
	DebugMode     bool              // 调试模式
}

var defaultOptions = &Options{
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
	DebugMode:     false,
}

// NewOptions creates a new Options instance with the given options applied
func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(o)
		}
	}
}

// -------------- ReqResp Options ----------------

// WithDebugMode sets the debug mode for the reqresp engine
func WithDebugMode(debug bool) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = debug
	})
}

// WithStaticLinkMap sets the static link map for path mapping
func WithStaticLinkMap(linkMap map[string]string) Option {
	return OptionFunc(func(o *Options) {
		for srcPath, dstPath := range linkMap {
			o.StaticLinkMap[srcPath] = dstPath
		}
	})
}

// WithPrefixLinkMap sets the prefix link map for path mapping
func WithPrefixLinkMap(linkMap map[string]string) Option {
	return OptionFunc(func(o *Options) {
		for srcPrefix, dstPrefix := range linkMap {
			o.PrefixLinkMap[srcPrefix] = dstPrefix
		}
	})
}

// WithStaticLink adds a single static link mapping
func WithStaticLink(srcPath, dstPath string) Option {
	return OptionFunc(func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	})
}

// WithPrefixLink adds a single prefix link mapping
func WithPrefixLink(srcPrefix, dstPrefix string) Option {
	return OptionFunc(func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	})
}
