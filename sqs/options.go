package sqs

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type RunMode string

const (
	RunModeStrict    RunMode = "strict"
	RunModePartial   RunMode = "partial"
	RunModeBatch     RunMode = "batch"
	RunModeReentrant RunMode = "reentrant"
)

type Options struct {
	// reserved for future sqs-specific options
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
	SQSClient     SQSClient
	RunMode       RunMode
	ReplyMode     bool
	DebugMode     bool
}

var defaultOptions = &Options{
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
	SQSClient:     nil,
	RunMode:       RunModeBatch,
	ReplyMode:     false,
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

// -------------- Sqs Options ----------------
func WithSQSClient(client SQSClient) Option {
	return OptionFunc(func(o *Options) {
		o.SQSClient = client
	})
}

func WithRunMode(mode RunMode) Option {
	return OptionFunc(func(o *Options) {
		switch mode {
		case RunModeStrict, RunModePartial, RunModeBatch, RunModeReentrant:
			o.RunMode = mode
		default:
			panic("sqs: unrecognized run mode: " + string(mode))
		}
	})
}

func WithReplyMode(reply bool) Option {
	return OptionFunc(func(o *Options) {
		o.ReplyMode = reply
	})
}

func WithDebugMode(debug bool) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = debug
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
