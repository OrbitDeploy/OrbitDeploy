package main

import (
	"fmt"
	"os"
	"strings"
)

// serverBase 返回不带 /api 的服务基础 URL。
// 要求用户输入的 ORBIT_API_BASE 默认不带 /api，API 路径由系统自动拼接。
func serverBase() string {
	b := os.Getenv("ORBIT_API_BASE")
	if b == "" {
		// 默认值不带 /api，API 路径由系统自动拼接。
		b = "http://localhost:8285"
	}
	b = strings.TrimRight(b, "/")
	b = strings.TrimSuffix(b, "/api")
	return b
}

// apiRoot 返回标准化后的 API 根路径（基础路径 + "/api"）。
func apiRoot() string {
	return serverBase() + "/api"
}

// 接口注册表：在此定义所有 API 路径，使用 fmt 格式化字符串。
var endpoints = map[string]string{
	"cli.configure.initiate":     "cli/configure/initiate",
	"cli.configure.status":       "cli/configure/status/%s",
	"auth.refresh_token":         "auth/refresh_token",
	"cli.device_auth.sessions":   "cli/device-auth/sessions",
	"cli.device_auth.token":      "cli/device-auth/token/%s",
	"auth.logout":                "auth/logout",
	"projects.variables":         "projects/%s/variables",
	"projects.images":            "projects/%s/images",
	"projects.deployments":       "projects/%s/deployments",
	"deployments.logs":           "deployments/%s/logs",
	"deployments.get":            "deployments/%s",
	"apps.by_name.get":           "apps/by-name/%s",
	"apps.by_name.releases":      "cli/apps/by-name/%s/releases",
	"apps.by_name.deployments":   "cli/apps/by-name/%s/deployments",
	"apps.by_name.config.export": "cli/apps/by-name/%s/config/export",
}

// apiURL 根据注册的端点 key 和参数构建完整的 API URL。
// 如果 key 未注册，则将 name 作为原始路径格式处理。
func apiURL(name string, args ...any) string {
	pattern, ok := endpoints[name]
	if !ok {
		// 灵活处理，回退为原始路径格式
		return apiURLf(name, args...)
	}
	return apiURLf(pattern, args...)
}

// apiURLf 通过格式化给定路径（不带前导 /api）构建完整 API URL。
// 使用 fmt.Sprintf，并加上标准化后的 API 根路径。
// 示例：apiURLf("projects/%s/variables", projectID)
func apiURLf(pathFmt string, args ...any) string {
	p := fmt.Sprintf(pathFmt, args...)
	p = strings.TrimLeft(p, "/")
	return apiRoot() + "/" + p
}
