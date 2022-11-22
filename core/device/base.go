package device

import (
	"context"
	"math/rand"
	"sync"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/core/message"
)

type Base struct {
	devices map[string][]Device
	gateway Device
	rwMutex sync.RWMutex
}

var _ Device = (*Base)(nil)
var empty Device = NewBase()

func NewBase() *Base {
	return &Base{
		devices: make(map[string][]Device),
	}
}

func (b *Base) String() string {
	return magic.Anonymous
}

func (b *Base) Join(device Device) {
	b.gateway = device
}

func (b *Base) Gateway() Device {
	return b.gateway
}

func (b *Base) Devices() map[string][]Device {
	return b.devices
}

func (b *Base) Extend(device Device) {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()

	name := device.String()

	for _, d := range b.devices[name] {
		if d == device {
			return
		}
	}
	b.devices[name] = append(b.devices[name], device)
}

func (b *Base) Locate(name string) Device {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()

	devices, ok := b.devices[name]
	if !ok {
		return nil
	}
	return devices[rand.Intn(len(devices))]
}

func (b *Base) Process(_ context.Context, msg *message.Message) error {
	return msg.Route.Error(ErrRouteDeadEnd)
}
