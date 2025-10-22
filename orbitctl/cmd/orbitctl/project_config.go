package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ConfigSessionInfo 表示配置会话信息
type ConfigSessionInfo struct {
	SessionID        string `json:"session_id"`
	ConfigurationURI string `json:"configuration_uri"`
	ExpiresIn        int    `json:"expires_in"`
}

// ProjectConfigData 表示从服务端获取的项目配置
type ProjectConfigData struct {
	ProjectID   string                 `json:"project_id"`
	ProjectName string                 `json:"project_name"`
	Spec        map[string]interface{} `json:"spec"`
	Environment map[string]interface{} `json:"environment"`
}

// initiateConfigSession 创建配置会话
func initiateConfigSession(tomlData string) (*ConfigSessionInfo, error) {
	url := apiURL("cli.configure.initiate")

	payload := map[string]interface{}{}
	if tomlData != "" {
		// 编码TOML数据为base64
		payload["toml_data"] = base64.StdEncoding.EncodeToString([]byte(tomlData))
	}

	resp, err := httpPostJSON(url, payload, false) // 不需要认证
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessionResp apiResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, err
	}

	if !sessionResp.Success {
		return nil, fmt.Errorf("服务器错误: %s", sessionResp.Message)
	}

	// 解析响应数据
	data := sessionResp.Data
	return &ConfigSessionInfo{
		SessionID:        getString(data, "session_id"),
		ConfigurationURI: getString(data, "configuration_uri"),
		ExpiresIn:        int(getFloat64(data, "expires_in")),
	}, nil
}

// waitForConfiguration 等待用户完成配置
// waitForConfigurationWithContext 使用 select 同时处理 ticker, timeout 和 cancellation
func waitForConfigurationWithContext(ctx context.Context, sessionID string, expiresIn int) (*ProjectConfigData, error) {
url := apiURL("cli.configure.status", sessionID)

	// 轮询间隔
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// 使用 time.After 创建一个在超时后发送事件的 channel
	timeout := time.Duration(expiresIn-30) * time.Second
	timeoutChan := time.After(timeout)

	fmt.Println("   提示: 您可以随时按 Ctrl+C 取消等待")

	// 首次立即检查一次，无需等待第一个 Ticker
	if config, done, err := checkStatus(url); done {
		return config, err
	}

	for {
		// select 会等待第一个就绪的 channel
		select {
		case <-timeoutChan:
			return nil, fmt.Errorf("配置会话已超时")

		case <-ctx.Done():
			// 如果外部上下文被取消 (例如按了 Ctrl+C)，则退出
			return nil, fmt.Errorf("操作被用户取消")

		case <-ticker.C:
			if config, done, err := checkStatus(url); done {
				return config, err
			}
			fmt.Print(".") // 打印点来表示正在等待
		}
	}
}

// checkStatus 是从循环中提取出的单次状态检查逻辑
func checkStatus(url string) (config *ProjectConfigData, done bool, err error) {
	resp, err := httpGetJSON(url, false)
	if err != nil {
		fmt.Printf("⚠️  检查配置状态失败: %v\n", err)
		return nil, false, nil // 返回非致命错误，让轮询继续
	}
	defer resp.Body.Close()

	var statusResp apiResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		fmt.Printf("⚠️  解析响应失败: %v\n", err)
		return nil, false, nil // 继续轮询
	}

	if !statusResp.Success {
		fmt.Printf("⚠️  服务器错误: %s\n", statusResp.Message)
		return nil, false, nil // 继续轮询
	}

	status := getString(statusResp.Data, "status")
	switch status {
	case "SUCCESS":
		fmt.Println("\n✅ 配置提交成功！")
		configData := &ProjectConfigData{
			ProjectID:   getString(statusResp.Data, "project_id"),
			ProjectName: getString(statusResp.Data, "project_name"),
			Spec:        getMap(statusResp.Data, "spec"),
			Environment: getMap(statusResp.Data, "environment"),
		}
		return configData, true, nil // (配置, 完成, 无错误)
	case "PENDING":
		return nil, false, nil // (无配置, 未完成, 无错误)
	case "EXPIRED":
		return nil, true, fmt.Errorf("配置会话已过期") // (无配置, 完成, 有错误)
	default:
		fmt.Printf("⚠️  未知状态: %s\n", status)
		return nil, false, nil // 继续轮询
	}
}

// generateLocalConfig 生成本地配置文件
func generateLocalConfig(filename string, config *ProjectConfigData) error {
	// 构建orbitctl.toml内容
	content := fmt.Sprintf(`# OrbitCtl 配置文件
# 由 orbitctl init 自动生成

# 服务端信息
api_base = "%s"

# 项目信息
project_id = "%s"
project_name = "%s"

# 部署规格和环境配置从服务端获取
# 使用 orbitctl deploy 时会自动同步最新配置
`,
		getAPIBase(),
		config.ProjectID,
		config.ProjectName,
	)

	// 写入文件
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}
