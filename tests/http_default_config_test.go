package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	lambdahttp "github.com/aura-studio/lambda/http"
	"github.com/aws/aws-lambda-go/events"
)

func TestHTTPWithDefaultConfigFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "http.yml")
	if err := os.WriteFile(p, []byte(`debug: true
cors: true
staticLink: []
prefixLink: []
headerLinkKey: []
`), 0o644); err != nil {
		t.Fatalf("write http.yaml: %v", err)
	}

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	o := lambdahttp.NewOptions(lambdahttp.WithDefaultConfigFile())
	if !o.DebugMode {
		t.Fatalf("DebugMode = false")
	}
	if !o.CorsMode {
		t.Fatalf("CorsMode = false")
	}
}

func TestSQSLocalHTTPServerMode(t *testing.T) {
	// t.Skip("Skipping integration test - requires running server")

	os.Setenv("_LAMBDA_SERVER_PORT", "8081")
	defer os.Unsetenv("_LAMBDA_SERVER_PORT")

	mainContent := `package main
import (
	"github.com/aws/aws-lambda-go/lambda"
	lambdasqs "github.com/aura-studio/lambda/sqs"
)
func main() {
	engine := lambdasqs.NewEngine()
	lambda.Start(engine.HandleSQSMessagesWithResponse)
}`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "-")
	cmd.Stdin = bytes.NewBufferString(mainContent)
	cmd.Env = append(os.Environ(), "_LAMBDA_SERVER_PORT=8081")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	t.Log("Waiting for server to start...")
	if !waitForServer("8081", 10*time.Second) {
		t.Fatal("Server failed to start within timeout")
	}
	t.Log("Server is ready")

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:   "test-msg-1",
				Body:        "test-body",
				EventSource: "aws:sqs",
			},
		},
	}

	eventData, err := json.Marshal(sqsEvent)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// 尝试多个可能的端点
	endpoints := []string{
		"http://localhost:8081/2015-03-31/functions/function/invocations",
		"http://localhost:8081/invoke",
		"http://localhost:8081/",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		t.Logf("Trying endpoint: %s", endpoint)
		resp, err := http.Post(
			endpoint,
			"application/json",
			bytes.NewBuffer(eventData),
		)
		if err != nil {
			lastErr = err
			t.Logf("Request error: %v", err)
			continue
		}
		defer resp.Body.Close()

		t.Logf("Response status: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusOK {
			var sqsResponse events.SQSEventResponse
			if err := json.NewDecoder(resp.Body).Decode(&sqsResponse); err != nil {
				t.Logf("Decode error: %v", err)
				continue
			}
			t.Logf("Response: %+v", sqsResponse)
			t.Log("Local HTTP server mode test completed successfully")
			return
		}
	}

	if lastErr != nil {
		t.Fatalf("All endpoints failed, last error: %v", lastErr)
	} else {
		t.Fatal("All endpoints returned non-200 status")
	}
}

func waitForServer(port string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:" + port + "/")
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}
