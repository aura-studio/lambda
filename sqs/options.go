package sqs

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

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

var defaultOptions = &Options{
	StaticLinkMap:  map[string]string{},
	PrefixLinkMap:  map[string]string{},
	SQSClient:      nil,
	ErrorSuspend:   false,
	PartialRetry:   false,
	ResponseSwitch: false,
	Release:        false,
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

// -------------- Sqs Options ----------------
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

func WithStaticLink(srcPath, dstPath string) Option {
	return OptionFunc(func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	})
}

func WithPrefixLink(srcPrefix string, dstPrefix string) Option {
	return OptionFunc(func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	})
}
