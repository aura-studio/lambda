package event

import (
	"context"
	"log"

	"github.com/aura-studio/lambda/dynamic"
)

// Engine is the core event processing engine.
type Engine struct {
	*Options
	*Router
	*dynamic.Dynamic
}

// NewEngine creates a new Engine with the given event and dynamic options,
// and automatically installs the default route handlers.
func NewEngine(eventOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(eventOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
		Router:  NewRouter(),
	}
	e.InstallHandlers()
	return e
}

// Invoke processes an incoming event request. Unlike reqresp.Engine.Invoke,
// this method only returns an error (no Response), reflecting the
// fire-and-forget semantics of AWS Lambda Event invocation.
func (e *Engine) Invoke(ctx context.Context, req *Request) error {
	_ = ctx

	c := &Context{}
	c.Set(ContextPath, req.Path)
	c.Set(ContextRequest, string(req.Payload))

	if e.DebugMode {
		log.Printf("[Event] Request: %s %s", c.GetString(ContextPath), c.GetString(ContextRequest))
	}

	e.Router.Dispatch(c)

	if e.DebugMode {
		log.Printf("[Event] Response: %s %s", c.GetString(ContextPath), c.GetString(ContextResponse))
	}

	if v, ok := c.Get(ContextPanic); ok && v != nil {
		return v.(error)
	}
	if err := c.GetError(); err != nil {
		return err
	}

	return nil
}
