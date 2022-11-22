package device_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/boost/ref"
	"github.com/aura-studio/lambda/core/device"
	"github.com/aura-studio/lambda/core/encoding"
	"github.com/aura-studio/lambda/core/message"
	"github.com/aura-studio/lambda/core/route"
)

var e1 = encoding.NewJSON()
var e2 = encoding.NewBase64()

type Try struct {
	logChan chan<- string
}

type Ping struct {
	Text string
}
type Pong struct {
	Text string
}

func (t *Try) Echo(_ context.Context, req *Ping) (resp *Pong, err error) {
	t.logChan <- fmt.Sprintf("%v", req)
	resp = &Pong{Text: req.Text}
	return resp, err
}

func (t *Try) EchoBytes(_ context.Context, req []byte) (resp []byte, err error) {
	t.logChan <- fmt.Sprintf("%v", string(req))
	resp = req
	return resp, err
}

type Client1 struct {
	*device.Base
	logChan chan<- string
}

func (s *Client1) Process(ctx context.Context, msg *message.Message) error {
	if !msg.Route.Dispatching() {
		return s.Gateway().Process(ctx, msg)
	}
	resp := &Pong{}
	if err := e1.Unmarshal(msg.Data, resp); err != nil {
		return err
	}
	s.logChan <- fmt.Sprintf("%v", resp)
	return nil
}

type Client2 struct {
	*device.Base
	logChan chan<- string
}

func (s *Client2) Process(ctx context.Context, msg *message.Message) error {
	if !msg.Route.Dispatching() {
		return s.Gateway().Process(ctx, msg)
	}
	bytes := &encoding.Bytes{}
	if err := e2.Unmarshal(msg.Data, bytes); err != nil {
		return err
	}
	s.logChan <- fmt.Sprintf("%v", string(bytes.Data))
	return nil
}

func TestRouter1(t *testing.T) {
	const (
		timeout     = 10
		version     = "1.0.0"
		logChanSize = 2
		msgID       = 0
	)

	logChan := make(chan string, logChanSize)
	client := &Client1{device.NewBase(), logChan}
	try := &Try{logChan}
	service := device.NewRouter(ref.TypeName(try)).Integrate(try)
	router := device.NewRouter(version).Integrate(service)
	bus := device.NewBus().Integrate(client, router)

	t.Logf("\n%s", device.Tree(bus))

	ctx := context.Background()
	r := route.NewChainRoute(magic.GoogleChain("/anonymous"), magic.GoogleChain("/1.0.0/try/echo"))
	reqData, err := encoding.Marshal(e1, &Ping{
		Text: "libra: Hello, world!",
	})
	if err != nil {
		t.Fatalf("unexpected error getting from encoding: %v", err)
	}
	msg := &message.Message{
		ID:       msgID,
		Route:    r,
		Encoding: e1,
		Data:     reqData,
	}

	if err = client.Process(ctx, msg); err != nil {
		t.Fatalf("unexpected error getting from device: %v", err)
	}
	var timeoutChan = time.After(time.Duration(timeout) * time.Second)
	var in string
	var out string
	for {
		select {
		case <-timeoutChan:
			t.Fatal("timeout when getting report from task")
		case msg := <-logChan:
			if in == "" {
				in = msg
				break
			}
			out = msg
			if in != out {
				t.Fatal("expecting out msg equals to in msg")
			}
			return
		}
	}
}

func TestRouter2(t *testing.T) {
	const (
		timeout     = 10
		version     = "1.0.0"
		logChanSize = 2
		msgID       = 0
	)
	logChan := make(chan string, logChanSize)
	client := &Client2{device.NewBase(), logChan}
	try := &Try{logChan}
	service := device.NewRouter(ref.TypeName(try)).Integrate(try)
	router := device.NewRouter(version).Integrate(service)
	bus := device.NewBus().Integrate(client, router)

	t.Logf("\n%s", device.Tree(bus))

	ctx := context.Background()
	r := route.NewChainRoute(magic.GoogleChain("/anonymous"), magic.GoogleChain("/1.0.0/try/echo-bytes"))

	reqData, err := encoding.Marshal(e2, []byte("libra: Hello, world!"))
	if err != nil {
		t.Fatalf("unexpected error getting from encoding: %v", err)
	}
	msg := &message.Message{
		ID:       msgID,
		Route:    r,
		Encoding: e2,
		Data:     reqData,
	}

	if err = client.Process(ctx, msg); err != nil {
		t.Fatalf("unexpected error getting from device: %v", err)
	}
	var timeoutChan = time.After(time.Duration(timeout) * time.Second)
	var in string
	var out string
	for {
		select {
		case <-timeoutChan:
			t.Fatal("timeout when getting report from task")
		case msg := <-logChan:
			if in == "" {
				in = msg
				break
			}
			out = msg
			if in != out {
				t.Fatal("expecting out msg equals to in msg")
			}
			return
		}
	}
}

func TestRouter3(t *testing.T) {
	const (
		timeout     = 10
		version     = "1.0.0"
		logChanSize = 2
	)
	logChan := make(chan string, logChanSize)
	client := device.NewClient("Anonymous")
	try := &Try{logChan}
	service := device.NewRouter(ref.TypeName(try)).Integrate(try)
	router := device.NewRouter(version).Integrate(service)
	bus := device.NewBus().Integrate(client, router)

	t.Logf("\n%s", device.Tree(bus))

	ctx := context.Background()
	r := route.NewChainRoute(magic.GoogleChain("/anonymous"), magic.GoogleChain("/1.0.0/try/echo-bytes"))

	reqData, err := encoding.Marshal(e2, []byte("libra: Hello, world!"))
	if err != nil {
		t.Fatalf("unexpected error getting from encoding: %v", err)
	}
	msg := &message.Message{
		ID:       0,
		Route:    r,
		Encoding: e2,
		Data:     reqData,
	}

	processor := device.NewFuncProcessor(func(context.Context, *message.Message) error {
		bytes := &encoding.Bytes{}
		if err := e2.Unmarshal(msg.Data, bytes); err != nil {
			return err
		}
		logChan <- fmt.Sprintf("%v", string(bytes.Data))
		return nil
	})

	if err = client.Invoke(ctx, msg, processor); err != nil {
		t.Fatalf("unexpected error getting from device: %v", err)
	}
	var timeoutChan = time.After(time.Duration(timeout) * time.Second)
	var in string
	var out string
	for {
		select {
		case <-timeoutChan:
			t.Fatal("timeout when getting report from task")
		case msg := <-logChan:
			if in == "" {
				in = msg
				break
			}
			out = msg
			if in != out {
				t.Fatal("expecting out msg equals to in msg")
			}
			return
		}
	}
}
