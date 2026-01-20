package http

import "github.com/aura-studio/lambda/dynamic"

type ServeOption interface {
	apply(*serveOptionBag)
}

type serveOptionBag struct {
	http    []Option
	dynamic []dynamic.Option
}

type httpServeOption struct{ opt Option }

func (o httpServeOption) apply(b *serveOptionBag) {
	if o.opt != nil {
		b.http = append(b.http, o.opt)
	}
}

type dynamicServeOption struct{ opt dynamic.Option }

func (o dynamicServeOption) apply(b *serveOptionBag) {
	if o.opt != nil {
		b.dynamic = append(b.dynamic, o.opt)
	}
}

func Http(opt Option) ServeOption { return httpServeOption{opt: opt} }

func Dyn(opt dynamic.Option) ServeOption { return dynamicServeOption{opt: opt} }
