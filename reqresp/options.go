package reqresp

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type Options struct {
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
	DebugMode     bool
}

var defaultOptions = &Options{
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
	DebugMode:     false,
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

func WithDebugMode(debug bool) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = debug
	})
}

func WithStaticLinkMap(linkMap map[string]string) Option {
	return OptionFunc(func(o *Options) {
		for srcPath, dstPath := range linkMap {
			o.StaticLinkMap[srcPath] = dstPath
		}
	})
}

func WithPrefixLinkMap(linkMap map[string]string) Option {
	return OptionFunc(func(o *Options) {
		for srcPrefix, dstPrefix := range linkMap {
			o.PrefixLinkMap[srcPrefix] = dstPrefix
		}
	})
}

func WithStaticLink(srcPath, dstPath string) Option {
	return OptionFunc(func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	})
}

func WithPrefixLink(srcPrefix, dstPrefix string) Option {
	return OptionFunc(func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	})
}
