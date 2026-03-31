package http

import (
	"strings"

	"github.com/mohae/deepcopy"
)

type Option interface {
	Apply(o *Options)
}

type HttpOption func(*Options)

func (f HttpOption) Apply(o *Options) { f(o) }

type LinkRule struct {
	Dst     string
	Methods []string // empty or ["ALL"] means all methods
}

type Options struct {
	// Http Options
	Address          string
	DebugMode        bool
	CorsMode         bool
	StaticLinkMap    map[string]LinkRule
	PrefixLinkMap    map[string]LinkRule
	PageNotFoundPath string
}

var defaultOptions = &Options{
	Address:          ":8080",
	DebugMode:        false,
	CorsMode:         false,
	StaticLinkMap:    map[string]LinkRule{},
	PrefixLinkMap:    map[string]LinkRule{},
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

func WithCorsMode() Option {
	return HttpOption(func(o *Options) {
		o.CorsMode = true
	})
}


func (r LinkRule) MatchMethod(method string) bool {
	if len(r.Methods) == 0 {
		return true
	}
	for _, m := range r.Methods {
		if strings.EqualFold(m, "ALL") {
			return true
		}
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func WithStaticLink(srcPath, dstPath string, methods ...string) Option {
	return HttpOption(func(o *Options) {
		o.StaticLinkMap[normalizePath(srcPath)] = LinkRule{Dst: normalizePath(dstPath), Methods: methods}
	})
}

func WithPrefixLink(srcPrefix string, dstPrefix string, methods ...string) Option {
	return HttpOption(func(o *Options) {
		o.PrefixLinkMap[normalizePath(srcPrefix)] = LinkRule{Dst: normalizePath(dstPrefix), Methods: methods}
	})
}

func WithPageNotFoundPath(path string) Option {
	return HttpOption(func(o *Options) {
		o.PageNotFoundPath = path
	})
}

