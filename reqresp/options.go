package reqresp

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type Options struct {
	DebugMode bool
}

var defaultOptions = &Options{
	DebugMode: false,
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

