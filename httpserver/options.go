package httpserver

type Options struct {
	namespace string
	links     [][]string
}

type Option func(*Options)

var options = Options{}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.namespace = namespace
	}
}

func WithLink(oldPath, newPath string) Option {
	return func(o *Options) {
		o.links = append(o.links, []string{oldPath, newPath})
	}
}
