package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	reqrespclient "github.com/aura-studio/lambda/reqresp/client"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
)

const (
	onlineGeoIPRegion       = "us-west-1"
	onlineGeoIPFunctionName = "scp-lambda-geoip-test-function-default"
)

func newOnlineLambdaClient(t *testing.T) *awslambda.Client {
	t.Helper()

	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(onlineGeoIPRegion),
	)
	if err != nil {
		t.Fatalf("LoadDefaultConfig failed: %v", err)
	}

	return awslambda.NewFromConfig(cfg)
}

func requireOnlineSyncTests(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_ONLINE_AWS_SYNC_TESTS") != "1" {
		t.Skip("skip synchronous online AWS tests: set RUN_ONLINE_AWS_SYNC_TESTS=1 to enable")
	}
}

func TestReqRespGeoIPOnline_QueryRequestResponse(t *testing.T) {
	requireOnlineSyncTests(t)

	lambdaClient := newOnlineLambdaClient(t)
	reqBody, err := json.Marshal(map[string]any{
		"RemoteAddr": "8.8.8.8",
	})
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}

	client := reqrespclient.NewClient(
		reqrespclient.WithLambdaClient(lambdaClient),
		reqrespclient.WithFunctionName(onlineGeoIPFunctionName),
		reqrespclient.WithDefaultTimeout(60*time.Second),
	)

	response, err := client.Call(context.Background(), "/api/geoip/v1/query", reqBody)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if response.Error != "" {
		t.Fatalf("Response error = %q", response.Error)
	}
	if len(response.Payload) == 0 {
		t.Fatal("Response payload should not be empty")
	}
}
