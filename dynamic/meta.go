package dynamic

import (
	"encoding/json"
	"os"
	"runtime/debug"
	"strings"
)

// ServiceInfo 服务信息，从 AWS_LAMBDA_FUNCTION_NAME 解析
type ServiceInfo struct {
	Business  string `json:"business"`
	Framework string `json:"framework"`
	Runtime   string `json:"runtime"`
	Resource  string `json:"resource"`
	Instance  string `json:"instance"`
}

// LambdaInfo Lambda 构建信息
type LambdaInfo struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Built   string `json:"built"`
}

// WarehouseInfo 仓库配置信息
type WarehouseInfo struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

// Meta 完整的 meta 信息结构
type Meta struct {
	Service   ServiceInfo   `json:"service"`
	Lambda    LambdaInfo    `json:"lambda"`
	Warehouse WarehouseInfo `json:"warehouse"`
}

// MetaGenerator meta 信息生成器
type MetaGenerator struct {
	localWarehouse  string
	remoteWarehouse string
}

// NewMetaGenerator 创建 meta 生成器
func NewMetaGenerator(localWarehouse, remoteWarehouse string) *MetaGenerator {
	return &MetaGenerator{
		localWarehouse:  localWarehouse,
		remoteWarehouse: remoteWarehouse,
	}
}

// parseServiceInfo 从 AWS_LAMBDA_FUNCTION_NAME 解析服务信息
// 格式: business-framework-runtime-resource-instance
func parseServiceInfo() ServiceInfo {
	funcName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	parts := strings.SplitN(funcName, "-", 5)

	info := ServiceInfo{}
	if len(parts) > 0 {
		info.Business = parts[0]
	}
	if len(parts) > 1 {
		info.Framework = parts[1]
	}
	if len(parts) > 2 {
		info.Runtime = parts[2]
	}
	if len(parts) > 3 {
		info.Resource = parts[3]
	}
	if len(parts) > 4 {
		info.Instance = parts[4]
	}

	return info
}

// parseLambdaInfo 从 debug.ReadBuildInfo 获取构建信息
func parseLambdaInfo() LambdaInfo {
	info := LambdaInfo{}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	info.Module = buildInfo.Main.Path
	info.Version = buildInfo.Main.Version

	// 查找 BuildTime 设置
	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.time" {
			info.Built = setting.Value
			break
		}
	}

	return info
}

// Generate 生成 meta 信息，合并 tunnel 的 Meta
func (g *MetaGenerator) Generate(tunnelMeta string) string {
	meta := Meta{
		Service: parseServiceInfo(),
		Lambda:  parseLambdaInfo(),
		Warehouse: WarehouseInfo{
			Local:  g.localWarehouse,
			Remote: g.remoteWarehouse,
		},
	}

	// 先序列化基础 meta
	result, err := json.Marshal(meta)
	if err != nil {
		return "{}"
	}

	// 如果 tunnelMeta 为空或无效，直接返回基础 meta
	if tunnelMeta == "" {
		return string(result)
	}

	// 尝试合并 tunnel 的 meta 信息
	var baseMap map[string]interface{}
	if err := json.Unmarshal(result, &baseMap); err != nil {
		return string(result)
	}

	var tunnelMap map[string]interface{}
	if err := json.Unmarshal([]byte(tunnelMeta), &tunnelMap); err != nil {
		// tunnelMeta 不是有效 JSON，直接返回基础 meta
		return string(result)
	}

	// 同层级合并，tunnel 的字段不覆盖已有字段
	for k, v := range tunnelMap {
		if _, exists := baseMap[k]; !exists {
			baseMap[k] = v
		}
	}

	merged, err := json.Marshal(baseMap)
	if err != nil {
		return string(result)
	}

	return string(merged)
}
