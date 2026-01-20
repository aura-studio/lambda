package dynamic

import (
	"github.com/mohae/deepcopy"
)

type Option func(*Options)

type Options struct {
	// Dynamic Options
	Os                    string
	Arch                  string
	Compiler              string
	Variant               string
	LocalWarehouse        string
	RemoteWarehouse       string
	PackageNamespace      string
	PackageDefaultVersion string
	StaticPackages        []*Package
	PreloadPackages       []*Package
}

var defaultOptions = &Options{
	Os:                    "",
	Arch:                  "",
	Compiler:              "",
	Variant:               "",
	LocalWarehouse:        "",
	RemoteWarehouse:       "",
	PackageNamespace:      "",
	PackageDefaultVersion: "",
	StaticPackages:        []*Package{},
	PreloadPackages:       []*Package{},
}

func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}
