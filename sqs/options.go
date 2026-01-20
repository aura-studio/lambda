package sqs

import "github.com/mohae/deepcopy"

type Options struct {
	// reserved for future sqs-specific options
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
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
