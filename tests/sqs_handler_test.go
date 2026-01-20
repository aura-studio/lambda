package tests

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdasqs "github.com/aura-studio/lambda/sqs"
	events "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"google.golang.org/protobuf/proto"
)

type mockSQSClient struct {
	messages []*sqs.SendMessageInput
}

func (m *mockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.messages = append(m.messages, params)
	return &sqs.SendMessageOutput{}, nil
}

func mustPBRequest(t *testing.T, r *lambdasqs.Request) string {
	t.Helper()
	b, err := proto.Marshal(r)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	return string(b)
}

type mockTunnel struct {
	invoke func(string, string) string
}

func (m *mockTunnel) Invoke(route string, req string) string {
	return m.invoke(route, req)
}

func (m *mockTunnel) Init() {
}

func (m *mockTunnel) Close() {
}

func TestSQSHandler_PartialFailures(t *testing.T) {
	e := lambdasqs.NewEngine()

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: "not-proto"}, // invalid protobuf -> fail
		{MessageId: "2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`), RequestSqsId: "c2"})}, // GetPackage likely fails -> fail
	}}

	err := e.HandleSQSMessagesWithoutResponse(context.Background(), ev)
	if err == nil {
		t.Fatal("HandleSQSMessagesWithoutResponse expected error, got nil")
	}
	// HandleSQSMessagesWithoutResponse returns error on failure, but doesn't return partial failures response directly.
	// The test logic for partial failures might need adjustment or use HandleSQSMessagesWithResponse.
	// Assuming we want to test partial failures, we should use HandleSQSMessagesWithResponse.

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 2 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
}

func TestSQSHandler_ResponseRouting(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine(lambdasqs.SQS(lambdasqs.WithSQSClient(mock)), lambdasqs.SQS(lambdasqs.WithResponseSwitch(true)))

	dynamic.RegisterPackage("pkg", "version", &mockTunnel{
		invoke: func(route, req string) string {
			return "OK"
		},
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "ignored-1", Body: mustPBRequest(t, &lambdasqs.Request{RequestSqsId: "client-1", ResponseSqsId: "server-9", CorrelationId: "corr-1", Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
		{MessageId: "ignored-2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`), RequestSqsId: "client-2"})},
	}}

	_, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(mock.messages) != 1 {
		t.Fatalf("out len = %d", len(mock.messages))
	}
	if *mock.messages[0].QueueUrl != "server-9" {
		t.Fatalf("QueueUrl = %q", *mock.messages[0].QueueUrl)
	}

	b, err := base64.StdEncoding.DecodeString(*mock.messages[0].MessageBody)
	if err != nil {
		t.Fatalf("base64.Decode: %v", err)
	}

	var rsp lambdasqs.Response
	if err := proto.Unmarshal(b, &rsp); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}

	if rsp.RequestSqsId != "client-1" {
		t.Fatalf("RequestSqsId = %q", rsp.RequestSqsId)
	}
	if rsp.ResponseSqsId != "server-9" {
		t.Fatalf("ResponseSqsId = %q", rsp.ResponseSqsId)
	}
	if strings.TrimSpace(string(rsp.Payload)) != "OK" {
		t.Fatalf("Payload = %q", string(rsp.Payload))
	}
}

func TestSQSHandler_NoResponse_AllowsEmptyClientSqsId(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine(lambdasqs.SQS(lambdasqs.WithSQSClient(mock)))

	dynamic.RegisterPackage("pkg", "version", &mockTunnel{
		invoke: func(route, req string) string {
			return "OK"
		},
	})

	// No ServerSqsId and no `?rsp=` in path => no response needed.
	// ClientSqsId is allowed to be empty in this case.
	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "m1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(mock.messages) != 0 {
		t.Fatalf("out len = %d", len(mock.messages))
	}
}

func TestSQSHandler_HealthCheck_OK(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine(lambdasqs.SQS(lambdasqs.WithSQSClient(mock)))

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "h1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(mock.messages) != 0 {
		t.Fatalf("out len = %d", len(mock.messages))
	}
}

func TestSQSHandler_APIPrefix_StripsToWildcardPath(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine(lambdasqs.SQS(lambdasqs.WithSQSClient(mock)))

	var gotRoute string
	dynamic.RegisterPackage("pkg", "version", &mockTunnel{
		invoke: func(route, req string) string {
			gotRoute = route
			return "OK"
		},
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "a1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(mock.messages) != 0 {
		t.Fatalf("out len = %d", len(mock.messages))
	}
	if gotRoute != "/route" {
		t.Fatalf("gotRoute = %q", gotRoute)
	}
}

func TestSQSHandler_APIPath_RequiresPrefix(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine(lambdasqs.SQS(lambdasqs.WithSQSClient(mock)))

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "p1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(mock.messages) != 0 {
		t.Fatalf("out len = %d", len(mock.messages))
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
}
