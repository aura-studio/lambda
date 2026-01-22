package server

import (
	"github.com/aura-studio/lambda/http"
	"github.com/aura-studio/lambda/sqs"
)

func Serve(addr string, opts ...ServeOption) error {
	options := &Options{}
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(options)
		}
	}

	switch options.ServerType {
	case "sqs":
		sqs.Serve(options.Sqs, options.Dynamic)
		return nil
	case "http":
		fallthrough
	default:
		return http.Serve(addr, options.Http, options.Dynamic)
	}
}

func Close() error {
	if err := http.Close(); err != nil {
		return err
	}
	sqs.Close()
	return nil
}
