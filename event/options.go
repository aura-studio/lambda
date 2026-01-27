package event

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type RunMode string

const (
	RunModeStrict    RunMode = "strict"    // 遇错停止
	RunModePartial   RunMode = "partial"   // 继续处理，记录失败
	RunModeBatch     RunMode = "batch"     // 整批失败
	RunModeReentrant RunMode = "reentrant" // 记录错误，继续处理
)

type Options struct {
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
	RunMode       RunMode
	DebugMode     bool
}

var defaultOptions = &Options{
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
	RunMode:       RunModeBatch, // Default to "batch" RunMode (Requirement 7.5)
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

// WithRunMode sets the run mode for error handling (Requirement 7.6)
func WithRunMode(mode RunMode) Option {
	return OptionFunc(func(o *Options) {
		switch mode {
		case RunModeStrict, RunModePartial, RunModeBatch, RunModeReentrant:
			o.RunMode = mode
		default:
			panic("event: unrecognized run mode: " + string(mode))
		}
	})
}

// WithDebugMode enables or disables debug mode
func WithDebugMode(debug bool) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = debug
	})
}

// WithStaticLink adds a static path mapping
func WithStaticLink(srcPath, dstPath string) Option {
	return OptionFunc(func(o *Options) {
		o.StaticLinkMap[srcPath] = dstPath
	})
}

// WithPrefixLink adds a prefix path mapping
func WithPrefixLink(srcPrefix, dstPrefix string) Option {
	return OptionFunc(func(o *Options) {
		o.PrefixLinkMap[srcPrefix] = dstPrefix
	})
}
