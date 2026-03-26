package reqresp

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
)

type Engine struct {
	*Options
	*dynamic.Dynamic
	r       *Router
	running atomic.Int32
}

func NewEngine(reqrespOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(reqrespOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
	}
	e.running.Store(1)
	e.InstallHandlers()
	return e
}

func (e *Engine) Start() {
	e.running.Store(1)
}

func (e *Engine) Stop() {
	e.running.Store(0)
}

func (e *Engine) IsRunning() bool {
	return e.running.Load() == 1
}

func (e *Engine) Invoke(ctx context.Context, req *Request) (*Response, error) {
	_ = ctx

	if e.running.Load() == 0 {
		return &Response{Error: "engine is stopped"}, nil
	}

	c := &Context{
		Engine:  e,
		RawPath: req.Path,
		Path:    req.Path,
		Request: string(req.Payload),
	}

	if e.DebugMode {
		log.Printf("[ReqResp] Request: %s %s", c.Path, c.Request)
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				c.Err = fmt.Errorf("panic: %v", r)
			}
		}()
		e.r.Dispatch(c)
	}()

	if e.DebugMode {
		log.Printf("[ReqResp] Response: %s %s", c.Path, c.Response)
	}

	resp := &Response{
		Payload: []byte(c.Response),
	}
	if c.Err != nil {
		resp.Error = c.Err.Error()
		if e.DebugMode {
			log.Printf("[ReqResp] Error: %v", c.Err)
		}
	}

	return resp, nil
}
