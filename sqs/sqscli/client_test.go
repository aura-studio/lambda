package sqscli

import (
	"context"
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/aura-studio/lambda/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"google.golang.org/protobuf/proto"
)

type mockSQSClient struct {
	sqs.SQSClient
	sentMessages     []*awssqs.SendMessageInput
	receivedMessages []*awssqs.ReceiveMessageInput
	deletedMessages  []*awssqs.DeleteMessageInput
	mu               sync.Mutex
	responseChan     chan types.Message
}

func (m *mockSQSClient) SendMessage(ctx context.Context, params *awssqs.SendMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, params)
	return &awssqs.SendMessageOutput{}, nil
}

func (m *mockSQSClient) ReceiveMessage(ctx context.Context, params *awssqs.ReceiveMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	m.mu.Lock()
	m.receivedMessages = append(m.receivedMessages, params)
	m.mu.Unlock()

	select {
	case msg := <-m.responseChan:
		return &awssqs.ReceiveMessageOutput{
			Messages: []types.Message{msg},
		}, nil
	case <-time.After(100 * time.Millisecond):
		return &awssqs.ReceiveMessageOutput{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *mockSQSClient) DeleteMessage(ctx context.Context, params *awssqs.DeleteMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletedMessages = append(m.deletedMessages, params)
	return &awssqs.DeleteMessageOutput{}, nil
}

func TestClient_Call(t *testing.T) {
	mock := &mockSQSClient{
		responseChan: make(chan types.Message, 1),
	}
	client := NewClient(
		WithSQSClient(mock),
		WithRequestSqsId("req-queue"),
		WithResponseSqsId("resp-queue"),
		WithDefaultTimeout(2*time.Second),
	)
	defer client.Close()

	// Simulate server response in a goroutine
	go func() {
		for {
			mock.mu.Lock()
			if len(mock.sentMessages) > 0 {
				sent := mock.sentMessages[0]
				mock.sentMessages = mock.sentMessages[1:]
				mock.mu.Unlock()

				// Decode request to get correlation ID
				b, _ := base64.StdEncoding.DecodeString(*sent.MessageBody)
				var req sqs.Request
				proto.Unmarshal(b, &req)

				// Create response
				resp := &sqs.Response{
					CorrelationId: req.CorrelationId,
					Payload:       []byte("OK"),
				}
				rb, _ := proto.Marshal(resp)
				mock.responseChan <- types.Message{
					Body:          aws.String(base64.StdEncoding.EncodeToString(rb)),
					ReceiptHandle: aws.String("handle"),
				}
				return
			}
			mock.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	resp, err := client.Call(context.Background(), "/test", []byte("hello"))
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if string(resp.Payload) != "OK" {
		t.Fatalf("Expected OK, got %s", string(resp.Payload))
	}
}

func TestClient_CallTimeout(t *testing.T) {
	mock := &mockSQSClient{
		responseChan: make(chan types.Message, 1),
	}
	client := NewClient(
		WithSQSClient(mock),
		WithRequestSqsId("req-queue"),
		WithResponseSqsId("resp-queue"),
		WithDefaultTimeout(100*time.Millisecond),
	)
	defer client.Close()

	_, err := client.Call(context.Background(), "/test", []byte("hello"))
	if err == nil || err.Error() != "request timeout" {
		t.Fatalf("Expected timeout error, got %v", err)
	}
}
