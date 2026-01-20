package http

import "github.com/aura-studio/lambda/dynamic"

type ServeOption any

type serveOptionBag struct {
	http    []Option
	dynamic []dynamic.Option
}

func (b *serveOptionBag) apply(opts ...ServeOption) {
	for _, opt := range opts {
		switch o := opt.(type) {
		case Option:
			b.http = append(b.http, o)
		case dynamic.Option:
			b.dynamic = append(b.dynamic, o)
		}
	}
}
