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
	ReleaseMode           bool
	CorsMode              bool
	LocalLibrary          string
	RemoteLibrary         string
	LibraryNamespace      string
	LibraryDefaultVersion string
	StaticLinkMap         map[string]string
	PrefixLinkMap         map[string]string
	StaticPackages        []*Package
	PreloadPackages       []*Package
	HeaderLinkMap         map[string]string
}

func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

type Option func(*Options)

var defaultOptions = &Options{
	LibraryNamespace:      "",
	LibraryDefaultVersion: "",
	StaticLinkMap:         map[string]string{},
	PrefixLinkMap:         map[string]string{},
	StaticPackages:        []*Package{},
	PreloadPackages:       []*Package{},
	HeaderLinkMap:         map[string]string{},
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithReleaseMode() Option {
	return func(o *Options) {
		o.ReleaseMode = true
	}
}

func WithCors() Option {
	return func(o *Options) {
		o.CorsMode = true
	}
}

func WithLocalLibrary() Option {
	return func(o *Options) {
		o.LocalLibrary = "local-library"
	}
}

func WithRemoteLibrary(remoteLibrary string) Option {
	return func(o *Options) {
		o.RemoteLibrary = remoteLibrary
	}
}

func WithLibraryNamespace(libraryNamespace string) Option {
	return func(o *Options) {
		o.LibraryNamespace = libraryNamespace
	}
}

func WithLibraryDefaultVersion(libraryDefaultVersion string) Option {
	return func(o *Options) {
		o.LibraryDefaultVersion = libraryDefaultVersion
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
