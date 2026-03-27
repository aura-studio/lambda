package reqresp

import (
	"context"
	"log"

	"github.com/aura-studio/lambda/dynamic"
)

type Engine struct {
	*Options
	*Router
	*dynamic.Dynamic
}

func NewEngine(reqrespOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(reqrespOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
		Router:  NewRouter(),
	}
	e.InstallHandlers()
	return e
}

func (e *Engine) Invoke(ctx context.Context, req *Request) (*Response, error) {
	_ = ctx

	c := &Context{}
	c.Set(ContextPath, req.Path)
	c.Set(ContextRequest, string(req.Payload))

	if e.DebugMode {
		log.Printf("[ReqResp] Request: %s %s", c.GetString(ContextPath), c.GetString(ContextRequest))
	}

	e.Router.Dispatch(c)

	if e.DebugMode {
		log.Printf("[ReqResp] Response: %s %s", c.GetString(ContextPath), c.GetString(ContextResponse))
	}

	resp := &Response{
		Payload: []byte(c.GetString(ContextResponse)),
	}
	if v, ok := c.Get(ContextPanic); ok && v != nil {
		resp.Error = v.(error).Error()
	} else if err := c.GetError(); err != nil {
		resp.Error = err.Error()
	}
	if resp.Error != "" && e.DebugMode {
		log.Printf("[ReqResp] Error: %s", resp.Error)
	}

	return resp, nil
}
