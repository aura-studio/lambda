package eventcli

import (
	"context"
	"fmt"

	"github.com/aura-studio/lambda/event"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"google.golang.org/protobuf/proto"
)

// Item 表示要发送的单个请求项
type Item struct {
	Path    string
	Payload []byte
}

// Client Event 客户端
type Client struct {
	*Options
}

// NewClient 创建新的客户端实例
func NewClient(opts ...Option) *Client {
	return &Client{
		Options: NewOptions(opts...),
	}
}

// Send 发送单个 item
// Requirement 4.7: THE Event_Client SHALL support sending single item via Send() method
// Requirement 4.1: THE Event_Client SHALL use Lambda Invoke API with InvocationType set to "Event"
// Requirement 4.4: THE Event_Client SHALL serialize the request using protobuf format
func (c *Client) Send(ctx context.Context, path string, payload []byte) error {
	// 创建包含单个 item 的 Request
	request := &event.Request{
		Items: []*event.Item{
			{
				Path:    path,
				Payload: payload,
			},
		},
	}

	return c.invoke(ctx, request)
}

// SendBatch 批量发送多个 items
// Requirement 4.8: THE Event_Client SHALL support sending multiple items in batch via SendBatch() method
// Requirement 4.1: THE Event_Client SHALL use Lambda Invoke API with InvocationType set to "Event"
// Requirement 4.4: THE Event_Client SHALL serialize the request using protobuf format
func (c *Client) SendBatch(ctx context.Context, items []Item) error {
	// 转换为 protobuf Item 列表
	protoItems := make([]*event.Item, len(items))
	for i, item := range items {
		protoItems[i] = &event.Item{
			Path:    item.Path,
			Payload: item.Payload,
		}
	}

	request := &event.Request{
		Items: protoItems,
	}

	return c.invoke(ctx, request)
}

// invoke 内部方法，执行 Lambda 调用
func (c *Client) invoke(ctx context.Context, request *event.Request) error {
	// Requirement 4.4: 使用 protobuf 序列化请求
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Requirement 4.1: 使用 InvocationType: Event 调用 Lambda
	_, err = c.LambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(c.FunctionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        requestBytes,
	})

	// Requirement 4.3: 如果 Lambda Invoke API 返回错误，返回错误给调用方
	if err != nil {
		return fmt.Errorf("lambda invoke failed: %w", err)
	}

	// Requirement 4.2: 如果 Lambda Invoke API 成功返回，立即返回 nil
	return nil
}
