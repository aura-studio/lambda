package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aura-studio/lambda/event"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Client event 客户端
type Client struct {
	*Options
}

// NewClient 创建新的客户端实例
func NewClient(opts ...Option) *Client {
	return &Client{
		Options: NewOptions(opts...),
	}
}

// Send 同步发送 Event 调用到 Lambda 函数
func (c *Client) Send(ctx context.Context, path string, payload []byte) error {
	request := &event.Request{
		Path:    path,
		Payload: payload,
	}

	invokePayload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := c.LambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(c.FunctionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        invokePayload,
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("request timeout")
		}
		return fmt.Errorf("lambda invoke failed: %w", err)
	}

	if output.FunctionError != nil {
		return fmt.Errorf("%s: %s", *output.FunctionError, string(output.Payload))
	}

	return nil
}

// SendAsync 异步发送 Event 调用到 Lambda 函数
func (c *Client) SendAsync(ctx context.Context, path string, payload []byte, callback func(error)) {
	go func() {
		err := c.Send(ctx, path, payload)
		if callback != nil {
			callback(err)
		}
	}()
}
