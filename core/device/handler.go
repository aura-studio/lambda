package device

import (
	"context"
	"reflect"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/core/encoding"
	"github.com/aura-studio/lambda/core/message"
)

type Handler struct {
	*Base
	receiver reflect.Value
	method   reflect.Method
}

func (h *Handler) String() string {
	return h.method.Name
}

func (h *Handler) Process(ctx context.Context, msg *message.Message) error {
	if !msg.Route.Dispatching() {
		if h.gateway == nil {
			return ErrGatewayNotFound
		}
		return h.gateway.Process(ctx, msg)
	}
	return h.localProcess(ctx, msg)
}

func (h *Handler) localProcess(ctx context.Context, reqMsg *message.Message) error {
	if h.method.Type == magic.TypeNil {
		return nil
	}

	respMsg, err := h.do(ctx, reqMsg)
	if err != nil {
		return err
	}
	if respMsg != nil {
		return h.Process(ctx, respMsg)
	}

	return nil
}

func (h *Handler) do(ctx context.Context, reqMsg *message.Message) (*message.Message, error) {
	mt := h.method.Type
	var req interface{}
	if mt.In(2) == magic.TypeOfBytes {
		bytes := &encoding.Bytes{}
		err := reqMsg.Encoding.Unmarshal(reqMsg.Data, bytes)
		if err != nil {
			return nil, err
		}
		req = bytes.Data
	} else { // TODO: support array of struct
		req = reflect.New(mt.In(2).Elem()).Interface()
		err := reqMsg.Encoding.Unmarshal(reqMsg.Data, req)
		if err != nil {
			return nil, err
		}
	}

	in := []reflect.Value{h.receiver, reflect.ValueOf(ctx), reflect.ValueOf(req)}

	out := h.method.Func.Call(in)
	if e := out[1].Interface(); e != nil {
		return nil, e.(error)
	}

	resp := out[0].Interface()

	respMsg := &message.Message{
		ID:       reqMsg.ID,
		Route:    reqMsg.Route.Reverse(),
		Encoding: reqMsg.Encoding.Reverse(),
	}
	if resp != nil {
		respData, err := reqMsg.Encoding.Marshal(resp)
		if err != nil {
			return nil, err
		}
		respMsg.Data = respData
	}

	return respMsg, nil
}

func extractHandlers(c interface{}) []Device {
	t := reflect.TypeOf(c)
	if t.Kind() != reflect.Ptr {
		return nil
	}
	if t.Elem().Kind() != reflect.Struct {
		return nil
	}

	var devices []Device
	for index := 0; index < t.NumMethod(); index++ {
		method := t.Method(index)
		mt := method.Type

		switch {
		case mt.PkgPath() != "": // Check method is exported
			continue
		case mt.NumIn() != 3: // Check num in
			continue
		case mt.NumOut() != 2: // Check num in
			continue
		case !mt.In(1).Implements(magic.TypeOfContext): // Check context.Context
			continue
		case !mt.Out(1).Implements(magic.TypeOfError): // Check error
			continue
		case mt.In(2).Kind() != reflect.Ptr && mt.In(2) != magic.TypeOfBytes: // Check request:  pointer or bytes
			continue
		case mt.Out(0).Kind() != reflect.Ptr && mt.Out(0) != magic.TypeOfBytes: // Check response: pointer or bytes
			continue
		}

		receiver := reflect.ValueOf(c)
		handler := &Handler{
			Base:     NewBase(),
			receiver: receiver,
			method:   method,
		}
		devices = append(devices, handler)
	}
	return devices
}
