package sqs

import "github.com/aura-studio/lambda/dynamic"

type ServeOption interface {
	apply(*serveOptionBag)
}

type serveOptionBag struct {
	sqs     []Option
	dynamic []dynamic.Option
}

type sqsServeOption struct{ opt Option }

func (o sqsServeOption) apply(b *serveOptionBag) {
	if o.opt != nil {
		b.sqs = append(b.sqs, o.opt)
	}
}

type dynamicServeOption struct{ opt dynamic.Option }

func (o dynamicServeOption) apply(b *serveOptionBag) {
	if o.opt != nil {
		b.dynamic = append(b.dynamic, o.opt)
	}
}

func SQS(opt Option) ServeOption { return sqsServeOption{opt: opt} }

func Dyn(opt dynamic.Option) ServeOption { return dynamicServeOption{opt: opt} }
