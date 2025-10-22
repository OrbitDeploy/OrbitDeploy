package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 项目配置会话相关结构
type projectConfigSessionCreateResp struct {
	SessionID string `json:"session_id"`
	ConfigURL string `json:"config_url"`
	ExpiresIn int    `json:"expires_in"`
}

type projectConfigSessionStatusResp struct {
	Status      string `json:"status"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
}

// 镜像上传响应
type imageUploadResp struct {
	ImageID string `json:"image_id"`
	Size    int64  `json:"size"`
	Tag     string `json:"tag"`
}

// 部署响应
type deploymentCreateResp struct {
	DeploymentID string `json:"deployment_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type deploymentStatusResp struct {
	DeploymentID string   `json:"deployment_id"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
	FinishedAt   string   `json:"finished_at,omitempty"`
	URLs         []string `json:"urls,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// CLI 专用类型定义
type releaseUploadResp struct {
	ReleaseID   string `json:"release_id"`
	Version     string `json:"version"`
	Description string `json:"description"`
	ImageSize   int64  `json:"image_size"`
	Status      string `json:"status"`
	AppName     string `json:"app_name"`
	AppID       uint   `json:"app_id"`
}

type applicationInfo struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	TargetPort      int    `json:"target_port"`
	Status          string `json:"status"`
	ProjectID       uint   `json:"project_id"`
	ActiveReleaseID *uint  `json:"active_release_id"`
}

// 环境变量相关结构
type envVariable struct {
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Secret    bool   `json:"secret"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type envVariablesResp struct {
	Variables []envVariable `json:"variables"`
}

type envUpdateResp struct {
	Message string `json:"message"`
}

// cmdInit 启动项目初始化，通过配置网页获取服务端配置
func cmdInit(name, project, env string) error {
	const filename = "orbitctl.toml"

	// 检查文件是否已存在
	if _, err := os.Stat(filename); err == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("配置文件 %s 已存在，是否覆盖? (y/N): ", filename)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("取消初始化")
			return nil
		}
	}

	fmt.Println("🚀 启动项目初始化...")

	// 1. 准备现有的TOML配置(如果有的话)
	var existingTomlData string
	if _, err := os.Stat("orbitdeploy.toml"); err == nil {
		if data, err := os.ReadFile("orbitdeploy.toml"); err == nil {
			existingTomlData = string(data)
			fmt.Println("📋 发现现有的 orbitdeploy.toml，将作为预填充数据")
		}
	}

	// 2. 请求配置网页链接
	fmt.Println("📡 请求配置会话...")
	sessionInfo, err := initiateConfigSession(existingTomlData)
	if err != nil {
		return fmt.Errorf("创建配置会话失败: %w", err)
	}

	fmt.Printf("✅ 配置会话已创建，ID: %s\n", sessionInfo.SessionID)
	fmt.Printf("⏱️  会话有效期: %d 分钟\n", sessionInfo.ExpiresIn/60)

	// 3. 打开配置网页
	fmt.Println("🌐 正在打开配置网页...")
	if err := openBrowser(sessionInfo.ConfigurationURI); err != nil {
		fmt.Printf("⚠️  自动打开浏览器失败: %v\n", err)
		fmt.Printf("请手动打开以下链接进行配置:\n%s\n\n", sessionInfo.ConfigurationURI)
	} else {
		fmt.Printf("✅ 配置网页已在浏览器中打开\n")
		fmt.Printf("📋 配置链接: %s\n\n", sessionInfo.ConfigurationURI)
	}

	// 4. 等待用户完成配置
	fmt.Println("⏳ 等待您在网页中完成项目配置...")
	fmt.Println("   请在打开的网页中:")
	fmt.Println("   1. 填写项目名称、描述等基本信息")
	fmt.Println("   2. 配置环境变量和域名")
	fmt.Println("   3. 点击提交按钮")
	fmt.Println("   4. 配置完成后网页会自动关闭")
	fmt.Println()

	// 5. 轮询等待配置完成
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	config, err := waitForConfigurationWithContext(ctx, sessionInfo.SessionID, sessionInfo.ExpiresIn)
	if err != nil {
		return fmt.Errorf("等待配置完成失败: %w", err)
	}

	// 6. 生成本地配置文件
	fmt.Println("💾 正在生成本地配置文件...")
	if err := generateLocalConfig(filename, config); err != nil {
		return fmt.Errorf("生成配置文件失败: %w", err)
	}

	fmt.Printf("✅ 项目初始化完成！\n")
	fmt.Printf("📄 已生成配置文件: %s\n", filename)
	if config.ProjectName != "" {
		fmt.Printf("📋 项目名称: %s\n", config.ProjectName)
	}
	if config.ProjectID != "" {
		fmt.Printf("🆔 项目ID: %s\n", config.ProjectID)
	}
	fmt.Println("\n下一步:")
	fmt.Printf("  1. 编辑 %s 查看完整配置\n", filename)
	fmt.Printf("  2. 运行 orbitctl spec-validate 验证配置\n")
	fmt.Printf("  3. 运行 orbitctl deploy 部署应用\n")

	return nil
}

// cmdDeploy 部署应用
func cmdDeploy(project, env string, dryRun bool) error {
	// 读取配置文件
	spec, err := loadSpecFromFile("orbitdeploy.toml")
	if err != nil {
		return err
	}

	// 使用命令行参数覆盖配置文件中的值
	if project != "" {
		spec.Project = project
	}
	if env != "" {
		spec.Environment = env
	}

	fmt.Printf("🚀 准备部署应用\n")
	fmt.Printf("   项目: %s\n", spec.Project)
	fmt.Printf("   环境: %s\n", spec.Environment)
	fmt.Printf("   应用: %s\n", spec.Name)
	fmt.Printf("   策略: %s\n", spec.Strategy)
	fmt.Printf("   副本: %d\n", spec.Replicas)

	if dryRun {
		fmt.Println("\n📋 部署计划 (--dry-run 模式):")
		fmt.Println("   [模拟] 1. 验证配置文件...")
		fmt.Println("   [模拟] 2. 创建项目配置会话...")
		fmt.Println("   [模拟] 3. 等待用户配置项目...")
		fmt.Println("   [模拟] 4. 构建并上传镜像...")
		fmt.Println("   [模拟] 5. 触发部署...")
		fmt.Println("   [模拟] 6. 监控部署进度...")
		fmt.Println("\n✨ 部署计划验证通过，使用 --dry-run=false 执行实际部署")
	} else {
		return performRealDeployment(spec)
	}

	return nil
}

// performRealDeployment 执行实际部署流程
func performRealDeployment(spec *specTOML) error {
	// 1. 验证应用存在
	fmt.Println("\n📋 步骤 1: 验证应用配置...")
	appName := spec.Name
	if appName == "" {
		return fmt.Errorf("应用名称不能为空")
	}

	app, err := getApplicationByName(appName)
	if err != nil {
		return fmt.Errorf("获取应用信息失败: %w", err)
	}
	fmt.Printf("   ✅ 找到应用: %s (ID: %d)\n", app.Name, app.ID)

	// 2. 构建并上传镜像
	fmt.Println("\n📦 步骤 2: 构建并上传镜像...")
	releaseID, err := buildAndUploadImageToApp(appName, spec)
	if err != nil {
		return fmt.Errorf("构建上传镜像失败: %w", err)
	}

	// 3. 触发部署
	fmt.Println("\n🚀 步骤 3: 触发部署...")
	deploymentID, err := triggerAppDeployment(appName, releaseID)
	if err != nil {
		return fmt.Errorf("触发部署失败: %w", err)
	}

	// 4. 监控部署进度
	fmt.Println("\n📊 步骤 4: 监控部署进度...")
	return monitorDeployment(deploymentID)
}

// cmdEnvList 列出环境变量
func cmdEnvList(project, env string) error {
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("📋 环境变量列表\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)

	url := apiURL("projects.variables", projectID)
	resp, err := httpGetJSON(url, true)
	if err != nil {
		return fmt.Errorf("获取环境变量失败: %w", err)
	}
	defer resp.Body.Close()

	var envResp apiResponse[envVariablesResp]
	if err := json.NewDecoder(resp.Body).Decode(&envResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !envResp.Success {
		return fmt.Errorf("获取环境变量失败: %s", envResp.Message)
	}

	fmt.Println("\n环境变量:")
	if len(envResp.Data.Variables) == 0 {
		fmt.Println("   (无环境变量)")
	} else {
		for _, v := range envResp.Data.Variables {
			if v.Secret {
				fmt.Printf("   %s = [隐藏] (密钥)\n", v.Key)
			} else {
				fmt.Printf("   %s = %s\n", v.Key, v.Value)
			}
			if v.UpdatedAt != "" {
				fmt.Printf("     更新时间: %s\n", v.UpdatedAt)
			}
		}
	}

	return nil
}

// cmdEnvSet 设置环境变量
func cmdEnvSet(keyValue, project, env string) error {
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("格式错误，应为 KEY=VALUE")
	}

	key, value := parts[0], parts[1]
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("🔧 设置环境变量\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)
	fmt.Printf("   变量: %s=%s\n", key, value)

	url := apiURL("projects.variables", projectID)
	payload := map[string]map[string]string{
		"variables": {
			key: value,
		},
	}

	resp, err := httpPostJSON(url, payload, true)
	if err != nil {
		return fmt.Errorf("设置环境变量失败: %w", err)
	}
	defer resp.Body.Close()

	var updateResp apiResponse[envUpdateResp]
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !updateResp.Success {
		return fmt.Errorf("设置环境变量失败: %s", updateResp.Message)
	}

	fmt.Printf("✅ %s\n", updateResp.Data.Message)
	return nil
}

// cmdEnvUnset 删除环境变量
func cmdEnvUnset(key, project, env string) error {
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("🗑️  删除环境变量\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)
	fmt.Printf("   变量: %s\n", key)

	// 通过设置空值来删除环境变量
	url := apiURL("projects.variables", projectID)
	payload := map[string]map[string]interface{}{
		"variables": {
			key: nil, // 设置为 nil 表示删除
		},
	}

	resp, err := httpPostJSON(url, payload, true)
	if err != nil {
		return fmt.Errorf("删除环境变量失败: %w", err)
	}
	defer resp.Body.Close()

	var updateResp apiResponse[envUpdateResp]
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !updateResp.Success {
		return fmt.Errorf("删除环境变量失败: %s", updateResp.Message)
	}

	fmt.Printf("✅ 环境变量 %s 已删除\n", key)
	return nil
}

// cmdScale 扩缩容
func cmdScale(replicasStr, project, env string) error {
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil {
		return fmt.Errorf("副本数必须为整数: %s", replicasStr)
	}
	if replicas < 0 {
		return fmt.Errorf("副本数不能为负数: %d", replicas)
	}

	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("📏 扩缩容操作\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)
	fmt.Printf("   目标副本数: %d\n", replicas)
	fmt.Println("\n⚠️  扩缩容功能需要部署API支持")
	fmt.Println("   将来会调用后端 API: POST /api/projects/:id/scale")
	return nil
}

// cmdStatus 查看状态
func cmdStatus(project, env string) error {
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("📊 应用状态\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)
	fmt.Println("\n⚠️  状态查询功能需要部署API支持")
	fmt.Println("   将来会调用后端 API: GET /api/projects/:id/status")
	return nil
}

// cmdLogs 查看日志
func cmdLogs(follow bool, project, env string) error {
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("📝 应用日志\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)
	fmt.Printf("   跟踪模式: %t\n", follow)
	fmt.Println("\n⚠️  日志查询功能需要部署API支持")
	fmt.Println("   将来会调用后端 API: GET /api/projects/:id/logs")
	if follow {
		fmt.Println("   并支持实时跟踪日志流")
	}
	return nil
}

// cmdInspect 检查配置
func cmdInspect(project, env string) error {
	projectID := getOrDefault(project, getProjectFromConfig())
	if projectID == "" {
		return fmt.Errorf("项目名称不能为空，请使用 --project 参数或在 orbitdeploy.toml 中指定")
	}

	fmt.Printf("🔍 配置检查\n")
	fmt.Printf("   项目: %s\n", projectID)
	fmt.Printf("   环境: %s\n", env)

	// 显示本地配置
	fmt.Println("\n📄 本地配置 (orbitdeploy.toml):")
	spec, err := loadSpecFromFile("orbitdeploy.toml")
	if err != nil {
		fmt.Printf("   错误: %v\n", err)
	} else {
		fmt.Printf("   项目: %s\n", spec.Project)
		fmt.Printf("   环境: %s\n", spec.Environment)
		fmt.Printf("   应用: %s\n", spec.Name)
		fmt.Printf("   策略: %s\n", spec.Strategy)
		fmt.Printf("   副本: %d\n", spec.Replicas)
		if len(spec.Containers) > 0 {
			fmt.Printf("   容器数: %d\n", len(spec.Containers))
		}
	}

	fmt.Println("\n⚠️  远程配置查询功能需要部署API支持")
	fmt.Println("   将来会显示:")
	fmt.Println("   1. 后端运行时配置")
	fmt.Println("   2. 合并后的最终配置")
	fmt.Println("   3. 部署历史和版本信息")
	return nil
}

// getApplicationByName 根据应用名称获取应用信息
func getApplicationByName(appName string) (*applicationInfo, error) {
	url := apiURL("apps.by_name.get", appName)
	resp, err := httpGetJSON(url, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var appResp apiResponse[applicationInfo]
	if err := json.NewDecoder(resp.Body).Decode(&appResp); err != nil {
		return nil, err
	}

	if !appResp.Success {
		return nil, fmt.Errorf("获取应用失败: %s", appResp.Message)
	}

	return &appResp.Data, nil
}

// buildAndUploadImageToApp 构建并上传镜像到指定应用
func buildAndUploadImageToApp(appName string, spec *specTOML) (string, error) {
	// 检查是否有 Dockerfile
	dockerfilePath := "Dockerfile"
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		// 如果没有 Dockerfile，检查配置中是否指定了镜像
		if len(spec.Containers) > 0 && spec.Containers[0].Image != nil && spec.Containers[0].Image.Ref != "" {
			fmt.Printf("   使用现有镜像: %s\n", spec.Containers[0].Image.Ref)
			return "existing-image", nil
		}
		return "", fmt.Errorf("未找到 Dockerfile 且未指定镜像")
	}

	// 构建镜像
	imageName := fmt.Sprintf("%s:cli-upload-%d", appName, time.Now().Unix())
	fmt.Printf("   构建镜像: %s\n", imageName)

	buildCmd := exec.Command("docker", "build", "-t", imageName, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("Docker构建失败: %w", err)
	}

	// 导出镜像为tar包
	tarPath := fmt.Sprintf("/tmp/%s.tar", appName)
	fmt.Printf("   导出镜像: %s\n", tarPath)

	saveCmd := exec.Command("docker", "save", "-o", tarPath, imageName)
	if err := saveCmd.Run(); err != nil {
		return "", fmt.Errorf("导出镜像失败: %w", err)
	}
	defer os.Remove(tarPath) // 清理临时文件

	// 上传镜像到应用
	fmt.Printf("   上传镜像到应用: %s\n", appName)
	releaseID, err := uploadImageToApplication(appName, tarPath, imageName)
	if err != nil {
		return "", fmt.Errorf("上传镜像失败: %w", err)
	}

	fmt.Printf("   ✅ 镜像上传成功: %s\n", releaseID)
	return releaseID, nil
}

// uploadImageToApplication 上传镜像文件到应用
func uploadImageToApplication(appName, filePath, version string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建multipart请求
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(fileWriter, file); err != nil {
		return "", err
	}

	// 添加版本信息
	if err := writer.WriteField("version", version); err != nil {
		return "", err
	}

	if err := writer.WriteField("description", "CLI upload"); err != nil {
		return "", err
	}

	writer.Close()

	// 发送请求
	url := apiURL("apps.by_name.releases", appName)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token := loadAccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var uploadResp apiResponse[releaseUploadResp]
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", err
	}

	if !uploadResp.Success {
		return "", fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	return uploadResp.Data.ReleaseID, nil
}

// triggerAppDeployment 触发应用部署
func triggerAppDeployment(appName, releaseID string) (string, error) {
	url := apiURL("apps.by_name.deployments", appName)
	payload := map[string]interface{}{
		"release_id": releaseID,
		"source":     "cli",
		"metadata": map[string]interface{}{
			"cli_version": "v0.1.0",
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	}

	resp, err := httpPostJSON(url, payload, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var deployResp apiResponse[deploymentCreateResp]
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return "", err
	}

	if !deployResp.Success {
		return "", fmt.Errorf("触发部署失败: %s", deployResp.Message)
	}

	fmt.Printf("   ✅ 部署已触发: %s\n", deployResp.Data.DeploymentID)
	return deployResp.Data.DeploymentID, nil
}

// buildAndUploadImage 构建并上传镜像
func buildAndUploadImage(projectID string, spec *specTOML) (string, error) {
	// 检查是否有 Dockerfile
	dockerfilePath := "Dockerfile"
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		// 如果没有 Dockerfile，检查配置中是否指定了镜像
		if len(spec.Containers) > 0 && spec.Containers[0].Image != nil && spec.Containers[0].Image.Ref != "" {
			fmt.Printf("   使用现有镜像: %s\n", spec.Containers[0].Image.Ref)
			return "existing-image", nil
		}
		return "", fmt.Errorf("未找到 Dockerfile 且未指定镜像")
	}

	// 构建镜像
	imageName := fmt.Sprintf("%s:cli-upload-%d", spec.Name, time.Now().Unix())
	fmt.Printf("   构建镜像: %s\n", imageName)

	buildCmd := exec.Command("docker", "build", "-t", imageName, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("Docker构建失败: %w", err)
	}

	// 导出镜像为tar包
	tarPath := fmt.Sprintf("/tmp/%s.tar", spec.Name)
	fmt.Printf("   导出镜像: %s\n", tarPath)

	saveCmd := exec.Command("docker", "save", "-o", tarPath, imageName)
	if err := saveCmd.Run(); err != nil {
		return "", fmt.Errorf("导出镜像失败: %w", err)
	}
	defer os.Remove(tarPath) // 清理临时文件

	// 上传镜像
	fmt.Printf("   上传镜像...\n")
	imageID, err := uploadImageFile(projectID, tarPath)
	if err != nil {
		return "", fmt.Errorf("上传镜像失败: %w", err)
	}

	fmt.Printf("   ✅ 镜像上传成功: %s\n", imageID)
	return imageID, nil
}

// uploadImageFile 上传镜像文件
func uploadImageFile(projectID, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建multipart请求
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(fileWriter, file); err != nil {
		return "", err
	}

	writer.Close()

	// 发送请求
url := apiURL("projects.images", projectID)
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token := loadAccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var uploadResp apiResponse[imageUploadResp]
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", err
	}

	if !uploadResp.Success {
		return "", fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	return uploadResp.Data.ImageID, nil
}

// triggerDeployment 触发部署
func triggerDeployment(projectID, imageID string) (string, error) {
url := apiURL("projects.deployments", projectID)
	payload := map[string]string{
		"image_id": imageID,
	}

	resp, err := httpPostJSON(url, payload, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var deployResp apiResponse[deploymentCreateResp]
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return "", err
	}

	if !deployResp.Success {
		return "", fmt.Errorf("触发部署失败: %s", deployResp.Message)
	}

	fmt.Printf("   ✅ 部署已触发: %s\n", deployResp.Data.DeploymentID)
	return deployResp.Data.DeploymentID, nil
}

// monitorDeployment 监控部署进度
func monitorDeployment(deploymentID string) error {
	fmt.Printf("   部署ID: %s\n", deploymentID)
	fmt.Println("   正在监控部署进度...")

	// 启动日志流监控
	go func() {
		if err := streamDeploymentLogs(deploymentID); err != nil {
			fmt.Printf("⚠️  获取部署日志失败: %v\n", err)
		}
	}()

	// 轮询部署状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := getDeploymentStatus(deploymentID)
			if err != nil {
				fmt.Printf("⚠️  获取部署状态失败: %v\n", err)
				continue
			}

			switch status.Status {
			case "SUCCESS":
				fmt.Println("\n✅ 部署成功！")
				if len(status.URLs) > 0 {
					fmt.Println("🌐 访问地址:")
					for _, url := range status.URLs {
						fmt.Printf("   %s\n", url)
					}
				}
				return nil
			case "FAILED":
				fmt.Println("\n❌ 部署失败")
				if status.ErrorMessage != "" {
					fmt.Printf("错误信息: %s\n", status.ErrorMessage)
				}
				return fmt.Errorf("部署失败")
			case "PENDING", "RUNNING":
				fmt.Print(".")
				continue
			default:
				fmt.Printf("⚠️  未知状态: %s\n", status.Status)
				continue
			}
		}
	}
}

// streamDeploymentLogs 流式获取部署日志 (SSE)
func streamDeploymentLogs(deploymentID string) error {
url := apiURL("deployments.logs", deploymentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if token := loadAccessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取SSE流
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "" {
				continue
			}

			// 解析日志数据
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &logData); err == nil {
				if msg, ok := logData["message"].(string); ok {
					fmt.Printf("📝 %s\n", msg)
				}
			}
		} else if strings.HasPrefix(line, "event: complete") {
			break
		}
	}

	return scanner.Err()
}

// getDeploymentStatus 获取部署状态
func getDeploymentStatus(deploymentID string) (*deploymentStatusResp, error) {
url := apiURL("deployments.get", deploymentID)
	resp, err := httpGetJSON(url, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statusResp apiResponse[deploymentStatusResp]
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	if !statusResp.Success {
		return nil, fmt.Errorf("获取状态失败: %s", statusResp.Message)
	}

	return &statusResp.Data, nil
}
