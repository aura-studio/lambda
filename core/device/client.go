package device

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/aura-studio/lambda/core/message"
)

var (
	ErrMissingProcessor = errors.New("processor cannot be found by message ID")
)

type Processor interface {
	Process(context.Context, *message.Message) error
}

type funcProcessor func(context.Context, *message.Message) error

func NewFuncProcessor(f funcProcessor) funcProcessor {
	return f
}

func (fp funcProcessor) Process(ctx context.Context, msg *message.Message) error {
	return fp(ctx, msg)
}

type Client struct {
	*Base
	name       string
	msgID      uint64
	processors sync.Map
}

func NewClient(name string) *Client {
	return &Client{
		Base: NewBase(),
		name: name,
	}
}

func (c *Client) String() string {
	return c.name
}

func (c *Client) Process(ctx context.Context, msg *message.Message) error {
	if !msg.Route.Dispatching() {
		if c.gateway == nil {
			return ErrGatewayNotFound
		}
		return c.gateway.Process(ctx, msg)
	}
	return c.localProcess(ctx, msg)
}

func (c *Client) localProcess(ctx context.Context, m *message.Message) error {
	v, ok := c.processors.Load(m.ID)
	if !ok {
		return ErrMissingProcessor
	}
	c.processors.Delete(m.ID)
	return v.(Processor).Process(ctx, m)
}

func (c *Client) Invoke(ctx context.Context, m *message.Message, p Processor) error {
	m.ID = atomic.AddUint64(&c.msgID, 1)
	c.processors.Store(m.ID, p)
	return c.Process(ctx, m)
}
