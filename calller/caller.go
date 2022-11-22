package calller

import (
	"context"
	"time"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/boost/safe"
	"github.com/aura-studio/lambda/core/device"
	"github.com/aura-studio/lambda/core/encoding"
	"github.com/aura-studio/lambda/core/message"
	"github.com/aura-studio/lambda/core/route"
)

type Caller struct {
	*device.Client
	timeout time.Duration
}

func NewCaller(d time.Duration) *Caller {
	client := device.NewClient("Client")
	device.Bus().Integrate(client)
	return &Caller{
		Client: client,
	}
}

func (c *Caller) Call(path string, req []byte) (resp []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	err = safe.DoWithContext(ctx, func() error {
		return c.Invoke(ctx, &message.Message{
			Route:    route.NewChainRoute(device.Addr(c), magic.GoogleChain(path)),
			Encoding: encoding.NewJSON(),
			Data:     req,
		}, device.NewFuncProcessor(func(ctx context.Context, msg *message.Message) error {
			resp = msg.Data
			return nil
		}))
	})

	return
}
