package device

import (
	"context"

	"github.com/aura-studio/lambda/core/message"
)

type VirtualRouter struct {
	*Base
	name string
	bus  bool
}

func NewVirtualRouter(name string) *VirtualRouter {
	r := &VirtualRouter{
		Base: NewBase(),
		name: name,
	}

	return r
}

func (r *VirtualRouter) String() string {
	return r.name
}

func (r *VirtualRouter) Process(ctx context.Context, msg *message.Message) error {
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

func (r *VirtualRouter) localProcess(ctx context.Context, msg *message.Message) error {
	device := r.Locate(msg.Route.Position())
	if device == nil {
		return msg.Route.Error(ErrRouteMissingDevice)
	}
	return device.Process(ctx, msg)
}

func (r *VirtualRouter) Integrate(targetList ...interface{}) *VirtualRouter {
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

func (r *VirtualRouter) AsBus() *VirtualRouter {
	r.bus = true
	return r
}
