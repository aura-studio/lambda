package device

import (
	"context"
	"errors"

	"github.com/aura-studio/lambda/core/message"

	"github.com/disiqueira/gotree"
)

var (
	ErrRouteDeadEnd       = errors.New("route has gone to a dead end")
	ErrRouteMissingDevice = errors.New("route has gone to a missing device")
	ErrGatewayNotFound    = errors.New("gateway is not found")
)

type Device interface {
	String() string
	Extend(Device)
	Join(Device)
	Locate(name string) Device
	Gateway() Device
	Devices() map[string][]Device
	Process(context.Context, *message.Message) error
}

func Tree(device Device) string {
	var add = func(d Device, t gotree.Tree) {}
	add = func(d Device, t gotree.Tree) {
		for _, deviceGroup := range d.Devices() {
			for _, device := range deviceGroup {
				add(device, t.Add(device.String()))
			}
		}
	}

	tree := gotree.New(device.String())
	add(device, tree)
	return tree.Print()
}

func Addr(device Device) []string {
	reverseAddr := []string{}
	for device != nil {
		reverseAddr = append(reverseAddr, device.String())
		device = device.Gateway()
	}
	addr := make([]string, 0, len(reverseAddr))
	for index := len(reverseAddr) - 1; index >= 0; index-- {
		addr = append(addr, reverseAddr[index])
	}
	return addr
}
