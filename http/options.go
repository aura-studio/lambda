package http

import (
	"github.com/mohae/deepcopy"
)

type Option interface {
	Apply(o *Options)
}

type HttpOption func(*Options)

func (f HttpOption) Apply(o *Options) { f(o) }

type Options struct {
	// Http Options
	Address          string
	DebugMode        bool
	CorsMode         bool
	StaticLinkMap    map[string]string
	PrefixLinkMap    map[string]string
	HeaderLinkMap    map[string]string
	PageNotFoundPath string
}

var defaultOptions = &Options{
	Address:          ":8080",
	DebugMode:        false,
	CorsMode:         false,
	StaticLinkMap:    map[string]string{},
	PrefixLinkMap:    map[string]string{},
	HeaderLinkMap:    map[string]string{},
	PageNotFoundPath: "",
}

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

// -------------- Http Options ----------------
func WithAddress(addr string) Option {
	return HttpOption(func(o *Options) {
		o.Address = addr
	})
}

func WithDebugMode() Option {
	return HttpOption(func(o *Options) {
		o.DebugMode = true
	})
}

func WithCors() Option {
	return HttpOption(func(o *Options) {
		o.CorsMode = true
	})
}

func WithStaticLink(srcPath, dstPath string) Option {
	return HttpOption(func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	})
}

func WithPrefixLink(srcPrefix string, dstPrefix string) Option {
	return HttpOption(func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	})
}

func WithHeaderLinkKey(key string, prefix string) Option {
	return HttpOption(func(o *Options) {
		o.HeaderLinkMap[key] = prefix
	})
}

func WithPageNotFoundPath(path string) Option {
	return HttpOption(func(o *Options) {
		o.PageNotFoundPath = path
	})
}
