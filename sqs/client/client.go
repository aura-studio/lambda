package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/aura-studio/lambda/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	*Options
	pendingRequests sync.Map // correlationId -> chan *sqs.Response
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		Options:  NewOptions(opts...),
		stopChan: make(chan struct{}),
	}

	if c.ResponseSqsId != "" {
		c.wg.Add(1)
		go c.listener()
	}

	return c
}

func (c *Client) Close() {
	close(c.stopChan)
	c.wg.Wait()
}

func (c *Client) listener() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stopChan:
			return
		default:
			output, err := c.SQSClient.ReceiveMessage(context.Background(), &awssqs.ReceiveMessageInput{
				QueueUrl:            &c.ResponseSqsId,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20,
			})
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			for _, msg := range output.Messages {
				c.handleIncomingMessage(msg)
				c.SQSClient.DeleteMessage(context.Background(), &awssqs.DeleteMessageInput{
					QueueUrl:      &c.ResponseSqsId,
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}


func (c *Client) handleIncomingMessage(msg types.Message) {
	if msg.Body == nil {
		return
	}
	b, err := base64.StdEncoding.DecodeString(*msg.Body)
	if err != nil {
		return
	}
	var resp sqs.Response
	if err := proto.Unmarshal(b, &resp); err != nil {
		return
	}

	if ch, ok := c.pendingRequests.Load(resp.CorrelationId); ok {
		ch.(chan *sqs.Response) <- &resp
	}
}

func (c *Client) Call(ctx context.Context, path string, payload []byte) (*sqs.Response, error) {
	correlationId := uuid.New().String()
	request := &sqs.Request{
		RequestSqsId:  c.RequestSqsId,
		ResponseSqsId: c.ResponseSqsId,
		CorrelationId: correlationId,
		Path:          path,
		Payload:       payload,
	}

	respChan := make(chan *sqs.Response, 1)
	c.pendingRequests.Store(correlationId, respChan)
	defer c.pendingRequests.Delete(correlationId)

	b, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	_, err = c.SQSClient.SendMessage(ctx, &awssqs.SendMessageInput{
		QueueUrl:    &c.RequestSqsId,
		MessageBody: aws.String(base64.StdEncoding.EncodeToString(b)),
	})
	if err != nil {
		return nil, err
	}

	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("request timeout")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) CallAsync(ctx context.Context, path string, payload []byte, callback func(*sqs.Response, error)) {
	go func() {
		resp, err := c.Call(ctx, path, payload)
		if callback != nil {
			callback(resp, err)
		}
	}()
}
