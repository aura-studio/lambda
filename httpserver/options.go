package httpserver

type Options struct {
	Namespace     string
	StaticLinkMap map[string]string
	PrefixLinkMap map[string]string
}

type Option func(*Options)

var options = Options{
	Namespace:     "",
	StaticLinkMap: map[string]string{},
	PrefixLinkMap: map[string]string{},
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

func WithStaticLink(oldPath, newPath string) Option {
	return func(o *Options) {
		o.StaticLinkMap[oldPath] = newPath
	}
}

func WithPrefixLink(oldPrefix string, newPrefix string) Option {
	return func(o *Options) {
		o.PrefixLinkMap[oldPrefix] = newPrefix
	}
}
