package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aura-studio/lambda/reqresp"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// Client reqresp 客户端
type Client struct {
	*Options
}

// NewClient 创建新的客户端实例
func NewClient(opts ...Option) *Client {
	return &Client{
		Options: NewOptions(opts...),
	}
}

// Call 同步调用 Lambda 函数
func (c *Client) Call(ctx context.Context, path string, payload []byte) (*reqresp.Response, error) {
	request := &reqresp.Request{
		Path:    path,
		Payload: payload,
	}

	invokePayload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := c.LambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(c.FunctionName),
		Payload:      invokePayload,
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, fmt.Errorf("lambda invoke failed: %w", err)
	}

	if output.FunctionError != nil {
		return &reqresp.Response{
			Error: fmt.Sprintf("%s: %s", *output.FunctionError, string(output.Payload)),
		}, nil
	}

	response := &reqresp.Response{}
	if err := json.Unmarshal(output.Payload, response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// CallAsync 异步调用 Lambda 函数
func (c *Client) CallAsync(ctx context.Context, path string, payload []byte, callback func(*reqresp.Response, error)) {
	go func() {
		resp, err := c.Call(ctx, path, payload)
		if callback != nil {
			callback(resp, err)
		}
	}()
}
