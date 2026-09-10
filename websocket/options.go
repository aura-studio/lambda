package websocket

import "github.com/mohae/deepcopy"

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type Options struct {
	// ManagementClient, when set, is used for all pushes regardless of the
	// request domain. Mainly useful for tests or single-domain deployments.
	ManagementClient ManagementClient
	// ManagementClientFactory builds a ManagementClient per request domain/stage.
	// Defaults to NewManagementClient when neither client nor factory is set.
	ManagementClientFactory ManagementClientFactory
	// ReplyMode pushes the handler response back to the requesting connection.
	ReplyMode bool
	DebugMode bool
	// ConnectHandler runs on $connect. Set ContextError on the context to reject
	// the connection (the gateway receives a non-200 status).
	ConnectHandler HandlerFunc
	// DisconnectHandler runs on $disconnect. Its result is ignored by the gateway.
	DisconnectHandler HandlerFunc
	// DefaultHandler, when set, replaces the built-in $default handler (the
	// envelope-to-tunnel Message dispatch). Use it when client messages should
	// not be constrained to the {"path","payload","meta"} envelope.
	DefaultHandler HandlerFunc
	// Routes holds custom route keys (e.g. "subscribe") mapped to handler chains,
	// installed after the built-in $connect/$disconnect/$default handlers.
	Routes map[string][]HandlerFunc
}

var defaultOptions = &Options{
	ManagementClient:        nil,
	ManagementClientFactory: nil,
	ReplyMode:               true,
	DebugMode:               false,
	ConnectHandler:          nil,
	DisconnectHandler:       nil,
	DefaultHandler:          nil,
	Routes:                  nil,
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

// -------------- WebSocket Options ----------------
func WithManagementClient(client ManagementClient) Option {
	return OptionFunc(func(o *Options) {
		o.ManagementClient = client
	})
}

func WithManagementClientFactory(factory ManagementClientFactory) Option {
	return OptionFunc(func(o *Options) {
		o.ManagementClientFactory = factory
	})
}

func WithReplyMode(reply bool) Option {
	return OptionFunc(func(o *Options) {
		o.ReplyMode = reply
	})
}

func WithDebugMode(debug bool) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = debug
	})
}

func WithConnectHandler(h HandlerFunc) Option {
	return OptionFunc(func(o *Options) {
		o.ConnectHandler = h
	})
}

// WithDefaultHandler replaces the built-in $default handler, freeing client
// messages from the {"path","payload","meta"} envelope. The handler receives
// the raw body via ContextRequest; ReplyMode does not apply to it.
func WithDefaultHandler(h HandlerFunc) Option {
	return OptionFunc(func(o *Options) {
		o.DefaultHandler = h
	})
}

func WithDisconnectHandler(h HandlerFunc) Option {
	return OptionFunc(func(o *Options) {
		o.DisconnectHandler = h
	})
}

func WithRoute(key string, handlers ...HandlerFunc) Option {
	return OptionFunc(func(o *Options) {
		if o.Routes == nil {
			o.Routes = map[string][]HandlerFunc{}
		}
		o.Routes[key] = append(o.Routes[key], handlers...)
	})
}
