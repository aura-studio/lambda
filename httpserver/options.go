package httpserver

import (
	"github.com/mohae/deepcopy"
)

type Options struct {
	// Http Options
	ReleaseMode   bool
	CorsMode      bool
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
	HeaderLinkMap map[string]string
}

func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

type Option func(*Options)

var defaultOptions = &Options{
	ReleaseMode:   false,
	CorsMode:      false,
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
	HeaderLinkMap: map[string]string{},
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// -------------- HttpServer Options ----------------
func WithReleaseMode() Option {
	return func(o *Options) {
		o.ReleaseMode = true
	}
}

func WithCors() Option {
	return func(o *Options) {
		o.CorsMode = true
	}
}

func WithStaticLink(srcPath, dstPath string) Option {
	return func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	}
}

func WithPrefixLink(srcPrefix string, dstPrefix string) Option {
	return func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	}
}

func WithHeaderLinkKey(key string, prefix string) Option {
	return func(o *Options) {
		o.HeaderLinkMap[key] = prefix
	}
}
