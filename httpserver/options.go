package httpserver

import (
	"github.com/aura-studio/dynamic"
	"github.com/mohae/deepcopy"
)

type Package struct {
	Name   string
	Commit string
	Tunnel dynamic.Tunnel
}

type Options struct {
	Namespace       string
	StaticLinkMap   map[string]string
	PrefixLinkMap   map[string]string
	StaticPackages  []*Package
	PreloadPackages []*Package
	HeaderLinkMap   map[string]string
}

func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

type Option func(*Options)

var defaultOptions = &Options{
	Namespace:       "",
	StaticLinkMap:   map[string]string{},
	PrefixLinkMap:   map[string]string{},
	StaticPackages:  []*Package{},
	PreloadPackages: []*Package{},
	HeaderLinkMap:   map[string]string{},
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
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

func WithStaticPackage(packageName, commit string, tunnel dynamic.Tunnel) Option {
	return func(o *Options) {
		o.StaticPackages = append(o.StaticPackages, &Package{
			Name:   packageName,
			Commit: commit,
			Tunnel: tunnel,
		})
	}
}

func WithPreloadPackage(packageName, commit string) Option {
	return func(o *Options) {
		o.PreloadPackages = append(o.PreloadPackages, &Package{
			Name:   packageName,
			Commit: commit,
		})
	}
}

func WithHeaderLinkKey(key string, prefix string) Option {
	return func(o *Options) {
		o.HeaderLinkMap[key] = prefix
	}
}
