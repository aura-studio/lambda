package reqrespcli

import (
	"context"
	"fmt"
	"time"

	"github.com/aura-studio/lambda/reqresp"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"google.golang.org/protobuf/proto"
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
	// 创建 Request protobuf
	request := &reqresp.Request{
		Path:    path,
		Payload: payload,
	}

	// 序列化请求
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 设置超时上下文
	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 调用 Lambda
	output, err := c.LambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(c.FunctionName),
		Payload:      requestBytes,
	})
	if err != nil {
		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, fmt.Errorf("lambda invoke failed: %w", err)
	}

	// 检查 Lambda 函数错误
	if output.FunctionError != nil {
		return &reqresp.Response{
			Error: *output.FunctionError,
		}, nil
	}

	// 解析 Response protobuf
	response := &reqresp.Response{}
	if err := proto.Unmarshal(output.Payload, response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
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
