package sqs

import "github.com/mohae/deepcopy"

type Options struct {
	// reserved for future sqs-specific options
	StaticLinkMap  map[string]string
	PrefixLinkMap  map[string]string
	SQSClient      SQSClient
	ErrorSuspend   bool
	PartialRetry   bool
	ResponseSwitch bool
	Release        bool
}

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

var defaultOptions = &Options{}

func NewOptions(opts ...Option) *Options {
	o := deepcopy.Copy(defaultOptions).(*Options)
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(o)
		}
	}
	return o
}

func WithSQSClient(client SQSClient) Option {
	return OptionFunc(func(o *Options) {
		o.SQSClient = client
	})
}

func WithErrorSuspend(suspend bool) Option {
	return OptionFunc(func(o *Options) {
		o.ErrorSuspend = suspend
	})
}

func WithPartialRetry(partial bool) Option {
	return OptionFunc(func(o *Options) {
		o.PartialRetry = partial
	})
}

func WithResponseSwitch(sw bool) Option {
	return OptionFunc(func(o *Options) {
		o.ResponseSwitch = sw
	})
}

func WithRelease(release bool) Option {
	return OptionFunc(func(o *Options) {
		o.Release = release
	})
}
