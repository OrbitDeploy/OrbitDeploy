package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/logman"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/utils"
)

// BuildService 负责处理应用构建相关的核心业务逻辑
type BuildService struct{}

// NewBuildService 创建新的构建服务实例
func NewBuildService() *BuildService {
	return &BuildService{}
}

// BuildFromGitHubRequest GitHub构建请求结构
type BuildFromGitHubRequest struct {
	RepoURL     string            `json:"repo_url"`
	Dockerfile  string            `json:"dockerfile"`
	ContextPath string            `json:"context_path"`
	BuildArgs   map[string]string `json:"build_args"`
}

// BuildFromApplicationRequest 应用构建请求结构
type BuildFromApplicationRequest struct {
	ApplicationID uuid.UUID         `json:"application_id"`
	Dockerfile    string            `json:"dockerfile"`
	ContextPath   string            `json:"context_path"`
	BuildArgs     map[string]string `json:"build_args"`
}

// RepositoryInfo 仓库信息结构，用于从ProviderAuth提取仓库信息
type RepositoryInfo struct {
	URL       string // 仓库URL
	IsPrivate bool   // 是否为私有仓库
	Platform  string // 平台类型（github, gitlab等）
	AuthToken string // 认证令牌
	Username  string // 用户名（某些平台需要）
}

// BuildImageFromApplication 从应用构建镜像的核心逻辑
// 这个函数与HTTP层完全解耦，可以被其他任何代码复用
func (bs *BuildService) BuildImageFromApplication(req BuildFromApplicationRequest) (string, error) {
	logman.Info("开始从应用构建镜像", "application_id", req.ApplicationID)

	// 1. 参数校验和设置默认值
	if req.ApplicationID == uuid.Nil {
		logman.Error("应用ID不能为空")
		return "", fmt.Errorf("应用ID不能为空")
	}
	if req.Dockerfile == "" {
		req.Dockerfile = "Dockerfile"
	}
	if req.ContextPath == "" {
		req.ContextPath = "."
	}
	logman.Info("参数校验完成", "dockerfile", req.Dockerfile, "context_path", req.ContextPath)

	// 2. 获取应用信息
	application, err := models.GetApplicationByID(req.ApplicationID)
	if err != nil {
		logman.Error("获取应用信息失败", "error", err)
		return "", fmt.Errorf("获取应用信息失败: %w", err)
	}
	logman.Info("获取应用信息成功", "application_name", application.Name)

	// 3. 获取项目信息
	project, err := models.GetProjectByID(application.ProjectID)
	if err != nil {
		logman.Error("获取项目信息失败", "error", err)
		return "", fmt.Errorf("获取项目信息失败: %w", err)
	}
	logman.Info("获取项目信息成功", "project_name", project.Name)

	// 4. 获取仓库信息（从ProviderAuth或本地CLI推送）
	repoInfo, err := bs.getRepositoryInfo(application)
	if err != nil {
		logman.Error("获取仓库信息失败", "error", err)
		return "", fmt.Errorf("获取仓库信息失败: %w", err)
	}

	// 5. 如果没有关联的ProviderAuth，说明是本地CLI推送，不支持构建
	if repoInfo == nil {
		logman.Error("应用未关联第三方仓库授权，无法从代码仓库构建")
		return "", fmt.Errorf("应用未关联第三方仓库授权，无法从代码仓库构建。请先关联ProviderAuth或使用本地CLI推送镜像")
	}

	logman.Info("获取仓库信息成功", "repo_url", repoInfo.URL, "platform", repoInfo.Platform)

	// 6. 构建请求结构体
	buildReq := BuildFromGitHubRequest{
		RepoURL:     repoInfo.URL,
		Dockerfile:  req.Dockerfile,
		ContextPath: req.ContextPath,
		BuildArgs:   req.BuildArgs,
	}
	logman.Info("构建请求结构体完成", "repo_url", repoInfo.URL)

	// 7. 使用统一的认证构建函数
	imageName, err := bs.buildImageFromRepoWithAuth(buildReq, application, project, repoInfo)
	if err != nil {
		return "", fmt.Errorf("从仓库构建镜像失败: %w", err)
	}

	return imageName, nil
}

// getRepositoryInfo 从应用的ProviderAuth获取仓库信息
func (bs *BuildService) getRepositoryInfo(application *models.Application) (*RepositoryInfo, error) {
	// 如果应用没有关联ProviderAuth，返回nil（支持本地CLI推送）
	if application.ProviderAuthID == nil {
		return nil, nil
	}

	// 获取ProviderAuth信息
	providerAuth, err := models.GetProviderAuthByID(*application.ProviderAuthID)
	if err != nil {
		return nil, fmt.Errorf("获取ProviderAuth失败: %w", err)
	}

	if !providerAuth.IsActive {
		return nil, fmt.Errorf("ProviderAuth已禁用")
	}

	// 检查RepoURL是否为空
	if application.RepoURL == nil || *application.RepoURL == "" {
		return nil, fmt.Errorf("Application中未设置仓库URL ")
	}

	repoURL := *application.RepoURL
	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
		return nil, fmt.Errorf("仓库URL必须是完整的可访问URL，例如 https://github.com/owner/repo 或 https://your-gitea.com/owner/repo")
	}

	// 根据平台类型构造仓库信息并生成认证令牌
	repoInfo := &RepositoryInfo{
		URL:       repoURL,
		Platform:  providerAuth.Platform,
		IsPrivate: true, // 现在所有仓库都通过授权处理
	}

	switch providerAuth.Platform {
	case "github":
		// GitHub Apps方式，生成安装令牌
		if providerAuth.AppID == "" || providerAuth.PrivateKey == "" || providerAuth.InstallationID == 0 {
			return nil, fmt.Errorf("GitHub平台需要完整的AppID、PrivateKey和InstallationID")
		}
		token, err := utils.GenerateGitHubAppInstallationToken(providerAuth.AppID, providerAuth.PrivateKey, providerAuth.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("生成GitHub安装令牌失败: %w", err)
		}
		repoInfo.AuthToken = token
		repoInfo.Username = "x-access-token" // GitHub推荐的用户名
	case "gitlab":
		repoInfo.AuthToken = providerAuth.ClientSecret
		repoInfo.Username = providerAuth.Username
	case "bitbucket":
		repoInfo.Username = providerAuth.Username
		repoInfo.AuthToken = providerAuth.AppPassword
	case "gitea":
		repoInfo.AuthToken = providerAuth.ClientSecret
		repoInfo.Username = providerAuth.Username
	default:
		return nil, fmt.Errorf("不支持的平台类型: %s", providerAuth.Platform)
	}

	return repoInfo, nil
}

// RewriteDockerfileShortNames rewrites short image names in the Dockerfile to include docker.io/ prefix
func (bs *BuildService) RewriteDockerfileShortNames(dockerfilePath string) (string, error) {
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	changed := false

	rewrite := func(img string) string {
		// 新增：判断是否为变量（以$开头），如果是则直接返回，不进行处理
		if strings.HasPrefix(img, "$") {
			return img
		}

		if strings.Contains(img, "/") {
			parts := strings.Split(img, "/")
			if len(parts) > 1 && strings.Contains(parts[0], ".") {
				return img // already has registry
			} else {
				return "docker.io/" + img // add docker.io for user/repo
			}
		} else {
			return "docker.io/library/" + img // add library for short names
		}
	}

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trim), "FROM ") {
			// Handle syntax: FROM image[:tag] [AS stage]
			parts := strings.Fields(trim)
			if len(parts) >= 2 {
				img := parts[1]
				newImg := rewrite(img)
				if newImg != img {
					changed = true
					parts[1] = newImg
					lines[i] = strings.Join(parts, " ")
				}
			}
		}
	}

	if !changed {
		return dockerfilePath, nil
	}

	// Write to a temp file next to original
	tmpPath := dockerfilePath + ".resolved"
	if err := os.WriteFile(tmpPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return "", err
	}
	return tmpPath, nil
}

// getDefaultBranch fetches the default branch from the platform's API.
// Supports self-hosted by extracting host from repoURL.
func (bs *BuildService) getDefaultBranch(repoURL string, repoInfo *RepositoryInfo) string {
	// Extract host and repoPath from full repoURL (e.g., https://github.com/owner/repo -> host=github.com, repoPath=owner/repo)
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) < 3 {
		logman.Warn("Invalid repo URL format, falling back to 'main'", "repo_url", repoURL)
		return "main"
	}
	host := parts[2]
	repoPath := strings.Join(parts[3:], "/")

	var apiURL string
	switch repoInfo.Platform {
	case "github":
		if host == "github.com" {
			apiURL = "https://api.github.com/repos/" + repoPath
		} else {
			apiURL = "https://" + host + "/api/v3/repos/" + repoPath // GitHub Enterprise
		}
	case "gitlab":
		apiURL = "https://" + host + "/api/v4/projects/" + url.QueryEscape(repoPath)
	case "gitea":
		apiURL = "https://" + host + "/api/v1/repos/" + repoPath
	case "bitbucket":
		if host == "bitbucket.org" {
			apiURL = "https://api.bitbucket.org/2.0/repositories/" + repoPath
		} else {
			logman.Warn("Self-hosted Bitbucket not fully supported, falling back to 'main'", "host", host)
			return "main"
		}
	default:
		logman.Warn("Unsupported platform, falling back to 'main'", "platform", repoInfo.Platform)
		return "main"
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logman.Warn("Failed to create API request, falling back to 'main'", "error", err)
		return "main"
	}

	// Set auth based on platform
	switch repoInfo.Platform {
	case "github", "gitea":
		req.Header.Set("Authorization", "token "+repoInfo.AuthToken)
	case "gitlab":
		req.Header.Set("Private-Token", repoInfo.AuthToken)
	case "bitbucket":
		req.SetBasicAuth(repoInfo.Username, repoInfo.AuthToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logman.Warn("API request failed, falling back to 'main'", "error", err)
		return "main"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logman.Warn("API returned non-200 status, falling back to 'main'", "status", resp.StatusCode)
		return "main"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logman.Warn("Failed to read API response, falling back to 'main'", "error", err)
		return "main"
	}

	var defaultBranch string
	switch repoInfo.Platform {
	case "github", "gitlab", "gitea":
		var repoInfoResp struct {
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.Unmarshal(body, &repoInfoResp); err != nil {
			logman.Warn("Failed to parse API response, falling back to 'main'", "error", err)
			return "main"
		}
		defaultBranch = repoInfoResp.DefaultBranch
	case "bitbucket":
		var repoInfoResp struct {
			MainBranch struct {
				Name string `json:"name"`
			} `json:"mainbranch"`
		}
		if err := json.Unmarshal(body, &repoInfoResp); err != nil {
			logman.Warn("Failed to parse API response, falling back to 'main'", "error", err)
			return "main"
		}
		defaultBranch = repoInfoResp.MainBranch.Name
	}

	logman.Info("Fetched default branch from API", "repo", repoPath, "branch", defaultBranch, "platform", repoInfo.Platform)
	return defaultBranch
}

// buildImageFromRepoWithAuth 统一的仓库构建函数，使用认证
func (bs *BuildService) buildImageFromRepoWithAuth(req BuildFromGitHubRequest, application *models.Application, project *models.Project, repoInfo *RepositoryInfo) (string, error) {
	// 1. 参数校验和设置默认值
	if strings.TrimSpace(req.RepoURL) == "" {
		return "", fmt.Errorf("仓库URL不能为空")
	}
	if req.Dockerfile == "" {
		req.Dockerfile = "Dockerfile"
	}
	if req.ContextPath == "" {
		req.ContextPath = "."
	}

	// 2. 生成临时目录
	tempDir, err := os.MkdirTemp("", "github-clone-auth-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir) // 确保清理

	// 3. 使用认证克隆仓库，使用指定的分支
	branch := "main" // Temporary default, will be overridden
	if application.Branch != nil && *application.Branch != "" {
		branch = *application.Branch
	} else {
		// Fetch default branch using repoInfo
		branch = bs.getDefaultBranch(req.RepoURL, repoInfo)
	}

	// authenticatedURL construction (req.RepoURL is now full)
	authenticatedURL := strings.Replace(req.RepoURL, "https://", fmt.Sprintf("https://%s:%s@", repoInfo.Username, repoInfo.AuthToken), 1)
	cloneCmd := exec.Command("git", "clone", "--depth=1", "--branch", branch, authenticatedURL, tempDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		fmt.Println(req.RepoURL, branch, authenticatedURL) // Debug output
		return "", fmt.Errorf("认证克隆仓库失败 (分支: %s): %s", branch, string(output))
	}

	if *application.BuildType != "dockerfile" {
		return "", fmt.Errorf("应用的构建类型不是Dockerfile，当前仅支持Dockerfile构建")
	}

	// 4. 验证Dockerfile存在
	contextDir := filepath.Join(tempDir, req.ContextPath)
	resolvedDockerfilePath := filepath.Join(contextDir, req.Dockerfile)
	if _, err := os.Stat(resolvedDockerfilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("dockerfile不存在: %s", resolvedDockerfilePath)
	}

	// 4.5. 重写Dockerfile中的镜像短名称
	resolvedDockerfilePath, err = bs.RewriteDockerfileShortNames(resolvedDockerfilePath)
	if err != nil {
		return "", fmt.Errorf("重写Dockerfile失败: %w", err)
	}

	// 5. 生成镜像名称
	timestamp := time.Now().Format("20060102-150405")
	repoName := filepath.Base(strings.TrimSuffix(req.RepoURL, ".git"))
	imageName := fmt.Sprintf("%s:%s-%s", strings.ToLower(repoName), branch, timestamp)

	// 6. 构造并执行podman build命令
	args := []string{"build", "--pull-always", "-t", imageName, "-f", resolvedDockerfilePath}
	for k, v := range req.BuildArgs {
		if strings.TrimSpace(k) == "" {
			continue
		}
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, contextDir)

	logman.Info("使用podman构建认证应用镜像", "image", imageName, "dockerfile", resolvedDockerfilePath, "context", contextDir, "branch", branch)
	buildCmd := exec.Command("podman", args...)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		// 构建失败，返回具体的Podman输出信息
		return "", fmt.Errorf("认证镜像构建失败: %s", string(output))
	}

	logman.Info("认证应用镜像构建成功", "image", imageName)
	return imageName, nil

}

// // buildImageFromPrivateRepoForApp 处理私有仓库的构建，使用认证（基于应用）
// func (bs *BuildService) buildImageFromPrivateRepoForApp(req BuildFromGitHubRequest, application *models.Application, project *models.Project, repoInfo *RepositoryInfo) (string, error) {
// 	// 1. 获取项目关联的GitHub token
// 	token, decryptedToken, err := models.GetGitHubTokenForProject(project.ID)
// 	if err != nil {
// 		return "", fmt.Errorf("获取私有仓库GitHub token失败: %w", err)
// 	}

// 	// 2. 构建认证配置
// 	authConfig := &utils.GitAuthConfig{
// 		Token:      decryptedToken,
// 		Username:   "x-access-token", // GitHub推荐的用户名
// 		AuthMethod: "token",
// 	}

// 	// 3. 调用带认证的构建函数
// 	imageName, err := bs.buildImageFromGitHubInternalWithAuthForApp(req, authConfig, application, project)
// 	if err != nil {
// 		return "", fmt.Errorf("从私有仓库构建镜像失败: %w", err)
// 	}

// 	// 4. 更新token最后使用时间
// 	if updateErr := models.UpdateTokenLastUsed(token.ID); updateErr != nil {
// 		logman.Warn("更新token最后使用时间失败", "token_id", token.ID, "error", updateErr)
// 	}

// 	return imageName, nil
// }
