package tests

import (
	"encoding/json"
	"testing"

	"github.com/aura-studio/lambda/dynamic"
)

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
