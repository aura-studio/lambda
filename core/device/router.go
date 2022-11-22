package device

import (
	"context"

	"github.com/aura-studio/lambda/core/message"
)

type Router struct {
	*Base
	name string
	bus  bool
}

func NewRouter(name string) *Router {
	r := &Router{
		Base: NewBase(),
		name: name,
	}

	return r
}

func NewBus() *Router {
	return NewRouter("Bus").AsBus()
}

var bus *Router = NewBus()

func Bus() *Router {
	return bus
}

func (r *Router) String() string {
	return r.name
}

func (r *Router) Process(ctx context.Context, msg *message.Message) error {
	if r.bus {
		if !msg.Route.Dispatching() {
			msg.Route = msg.Route.Forward()
			return r.localProcess(ctx, msg)
		}

		return msg.Route.Error(ErrRouteDeadEnd)
	}

	if !msg.Route.Dispatching() {
		if r.gateway == nil {
			return ErrGatewayNotFound
		}
		return r.gateway.Process(ctx, msg)
	}
	msg.Route = msg.Route.Forward()
	return r.localProcess(ctx, msg)
}

func (r *Router) localProcess(ctx context.Context, msg *message.Message) error {
	device := r.Locate(msg.Route.Position())
	if device == nil {
		return msg.Route.Error(ErrRouteMissingDevice)
	}
	return device.Process(ctx, msg)
}

func (r *Router) Integrate(targetList ...interface{}) *Router {
	for _, target := range targetList {
		if device, ok := target.(Device); ok {
			r.Extend(device)
			device.Join(r)
		} else {
			for _, device := range extractHandlers(target) {
				r.Extend(device)
				device.Join(r)
			}
		}
	}
	return r
}

func (r *Router) AsBus() *Router {
	r.bus = true
	return r
}
