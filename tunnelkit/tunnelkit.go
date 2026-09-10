// Package tunnelkit is an optional scaffold for building dynamic.Tunnel
// implementations. It absorbs the three pieces of boilerplate every tunnel
// author otherwise repeats: route dispatch, the {"meta","data"} base64
// envelope, and the meta.Error business-error convention.
//
// The underlying dynamic.Tunnel interface is untouched, so existing tunnels
// keep working; tunnelkit only serves new code:
//
//	echo := tunnelkit.New("")
//	echo.Handle("/echo", func(c *tunnelkit.Context) error {
//		return c.JSON(map[string]any{"echo": json.RawMessage(c.Payload)})
//	})
//	dynamic.RegisterPackage("echo", "v1", echo)
//
// It works with every engine in this repository (http / sqs / websocket),
// since all of them wrap tunnel requests with the same envelope.
package tunnelkit

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// RspMetaError is the response meta key engines translate into a client-visible
// error. Kept identical to the engines' convention.
const RspMetaError = "Error"

// HandlerFunc handles one registered route. Return an error to produce a
// meta.Error response; otherwise whatever was written via JSON/Data/String is
// delivered as the payload.
type HandlerFunc func(c *Context) error

// Context carries one tunnel invocation: the decoded request and the response
// being built.
type Context struct {
	// Route is the path inside the package (e.g. "/echo").
	Route string
	// Meta is the request envelope's meta map (never nil).
	Meta map[string]any
	// Payload is the raw decoded request data.
	Payload []byte

	resp     []byte
	respMeta map[string]any
}

// UnmarshalPayload decodes the payload as JSON into v.
func (c *Context) UnmarshalPayload(v any) error {
	return json.Unmarshal(c.Payload, v)
}

// JSON marshals v and sets it as the response payload.
func (c *Context) JSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.resp = b
	return nil
}

// Data sets raw bytes as the response payload.
func (c *Context) Data(b []byte) { c.resp = b }

// String sets a plain string as the response payload.
func (c *Context) String(s string) { c.resp = []byte(s) }

// SetRespMeta adds a key to the response envelope's meta map.
func (c *Context) SetRespMeta(key string, value any) {
	if c.respMeta == nil {
		c.respMeta = map[string]any{}
	}
	c.respMeta[key] = value
}

// Tunnel implements dynamic.Tunnel with route registration.
type Tunnel struct {
	meta   string
	routes map[string]HandlerFunc
	order  []string // registration order, for deterministic Meta()
}

// New creates a Tunnel. Pass an empty meta string to have Meta() generated
// from the registered routes.
func New(meta string) *Tunnel {
	return &Tunnel{
		meta:   meta,
		routes: map[string]HandlerFunc{},
	}
}

// Handle registers h for route (e.g. "/echo"). Registering the same route
// twice replaces the previous handler.
func (t *Tunnel) Handle(route string, h HandlerFunc) {
	if _, exists := t.routes[route]; !exists {
		t.order = append(t.order, route)
	}
	t.routes[route] = h
}

// Meta returns the meta string passed to New, or a generated route listing
// when none was given.
func (t *Tunnel) Meta() string {
	if t.meta != "" {
		return t.meta
	}
	b, _ := json.Marshal(map[string]any{"routes": t.order})
	return string(b)
}

func (t *Tunnel) Init()  {}
func (t *Tunnel) Close() {}

// Invoke unwraps the envelope, dispatches to the registered handler, and wraps
// the result. Unknown routes, malformed envelopes, and handler errors all map
// to the meta.Error convention.
func (t *Tunnel) Invoke(route, req string) string {
	var envelope struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}
	if err := json.Unmarshal([]byte(req), &envelope); err != nil {
		return errResponse(fmt.Sprintf("invalid request envelope: %v", err))
	}

	payload, err := base64.StdEncoding.DecodeString(envelope.Data)
	if err != nil {
		return errResponse(fmt.Sprintf("invalid request data: %v", err))
	}

	h, ok := t.routes[route]
	if !ok {
		return errResponse("unknown route: " + route)
	}

	c := &Context{Route: route, Meta: envelope.Meta, Payload: payload}
	if c.Meta == nil {
		c.Meta = map[string]any{}
	}

	if err := h(c); err != nil {
		return errResponse(err.Error())
	}

	return okResponse(c.resp, c.respMeta)
}

func okResponse(payload []byte, meta map[string]any) string {
	if meta == nil {
		meta = map[string]any{}
	}
	resp, _ := json.Marshal(map[string]any{
		"meta": meta,
		"data": base64.StdEncoding.EncodeToString(payload),
	})
	return string(resp)
}

func errResponse(msg string) string {
	resp, _ := json.Marshal(map[string]any{
		"meta": map[string]any{RspMetaError: msg},
		"data": "",
	})
	return string(resp)
}
