package sqscli

import (
	"time"

	"github.com/aura-studio/lambda/sqs"
	"github.com/mohae/deepcopy"
)

type Options struct {
	SQSClient      sqs.SQSClient
	RequestSqsId   string
	ResponseSqsId  string
	DefaultTimeout time.Duration
}

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

var defaultOptions = &Options{
	DefaultTimeout: 30 * time.Second,
}

func NewOptions(opts ...Option) *Options {
	o := deepcopy.Copy(defaultOptions).(*Options)
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(o)
		}
	}
	return o
}

func WithSQSClient(client sqs.SQSClient) Option {
	return OptionFunc(func(o *Options) {
		o.SQSClient = client
	})
}

func WithRequestSqsId(id string) Option {
	return OptionFunc(func(o *Options) {
		o.RequestSqsId = id
	})
}

func WithResponseSqsId(id string) Option {
	return OptionFunc(func(o *Options) {
		o.ResponseSqsId = id
	})
}

func WithDefaultTimeout(timeout time.Duration) Option {
	return OptionFunc(func(o *Options) {
		o.DefaultTimeout = timeout
	})
}
