package tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aura-studio/lambda/dynamic"
)

func TestParseServiceInfo(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		wantBiz    string
		wantFw     string
		wantComp   string
		wantRt     string
		wantRes    string
		wantInst   string
	}{
		{
			name:       "完整6段格式",
			funcName:   "payment-orderservice-refund-prod-api-001",
			wantBiz:    "payment",
			wantFw:     "orderservice",
			wantComp:   "refund",
			wantRt:     "prod",
			wantRes:    "api",
			wantInst:   "001",
		},
		{
			name:       "5段格式（缺少instance）",
			funcName:   "payment-orderservice-refund-prod-api",
			wantBiz:    "payment",
			wantFw:     "orderservice",
			wantComp:   "refund",
			wantRt:     "prod",
			wantRes:    "api",
			wantInst:   "",
		},
		{
			name:       "3段格式",
			funcName:   "payment-orderservice-refund",
			wantBiz:    "payment",
			wantFw:     "orderservice",
			wantComp:   "refund",
			wantRt:     "",
			wantRes:    "",
			wantInst:   "",
		},
		{
			name:       "空字符串",
			funcName:   "",
			wantBiz:    "",
			wantFw:     "",
			wantComp:   "",
			wantRt:     "",
			wantRes:    "",
			wantInst:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			os.Setenv("AWS_LAMBDA_FUNCTION_NAME", tt.funcName)
			defer os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")

			gen := dynamic.NewMetaGenerator()
			result := gen.Generate("")

			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(result), &meta); err != nil {
				t.Fatalf("Failed to parse meta JSON: %v", err)
			}

			service, ok := meta["service"].(map[string]interface{})
			if !ok {
				t.Fatal("service field not found or not an object")
			}

			if got := service["business"].(string); got != tt.wantBiz {
				t.Errorf("business = %q, want %q", got, tt.wantBiz)
			}
			if got := service["framework"].(string); got != tt.wantFw {
				t.Errorf("framework = %q, want %q", got, tt.wantFw)
			}
			if got := service["component"].(string); got != tt.wantComp {
				t.Errorf("component = %q, want %q", got, tt.wantComp)
			}
			if got := service["runtime"].(string); got != tt.wantRt {
				t.Errorf("runtime = %q, want %q", got, tt.wantRt)
			}
			if got := service["resource"].(string); got != tt.wantRes {
				t.Errorf("resource = %q, want %q", got, tt.wantRes)
			}
			if got := service["instance"].(string); got != tt.wantInst {
				t.Errorf("instance = %q, want %q", got, tt.wantInst)
			}
		})
	}
}

func TestInitLambdaInfo(t *testing.T) {
	// 生成 meta（NewMetaGenerator 内部会调用 initLambdaInfo）
	gen := dynamic.NewMetaGenerator()
	result := gen.Generate("")

	t.Logf("Meta output:\n%s", result)

	// 解析 JSON
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(result), &meta); err != nil {
		t.Fatalf("Failed to parse meta JSON: %v", err)
	}

	// 检查 lambda 字段
	lambda, ok := meta["lambda"].(map[string]interface{})
	if !ok {
		t.Fatal("lambda field not found or not an object")
	}

	module := lambda["module"].(string)
	version := lambda["version"].(string)
	built := lambda["built"].(string)

	t.Logf("module: %q", module)
	t.Logf("version: %q", version)
	t.Logf("built: %q", built)

	// module 应该有值（测试时是 github.com/aura-studio/lambda）
	if module == "" {
		t.Error("module should not be empty")
	}
}
