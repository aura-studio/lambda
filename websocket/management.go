package websocket

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
)

// ManagementClient is the subset of the API Gateway management API used to push
// messages to and manage established WebSocket connections.
type ManagementClient interface {
	PostToConnection(ctx context.Context, params *apigatewaymanagementapi.PostToConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.PostToConnectionOutput, error)
	DeleteConnection(ctx context.Context, params *apigatewaymanagementapi.DeleteConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.DeleteConnectionOutput, error)
	GetConnection(ctx context.Context, params *apigatewaymanagementapi.GetConnectionInput, optFns ...func(*apigatewaymanagementapi.Options)) (*apigatewaymanagementapi.GetConnectionOutput, error)
}

// ManagementClientFactory builds a ManagementClient bound to the API Gateway
// domain/stage that received the event (https://{domainName}/{stage}).
type ManagementClientFactory func(ctx context.Context, domainName, stage string) (ManagementClient, error)

// NewManagementClient is the default ManagementClientFactory.
func NewManagementClient(ctx context.Context, domainName, stage string) (ManagementClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s/%s", domainName, stage))
	}), nil
}

// IsGone reports whether err means the connection no longer exists (HTTP 410),
// in which case the caller should drop its stored connection record.
func IsGone(err error) bool {
	var gone *types.GoneException
	return errors.As(err, &gone)
}
