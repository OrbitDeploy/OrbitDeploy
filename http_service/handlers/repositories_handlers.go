package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/labstack/echo/v4"
)

// ListRepositoriesHandler fetches repositories for a given provider auth ID
func ListRepositoriesHandler(c echo.Context) error {
	// Get provider auth ID from URL
	idStr := c.Param("uid")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid provider auth ID format")
	}

	// Retrieve provider auth from database
	providerAuth, err := models.GetProviderAuthByID(id)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Provider auth not found")
	}

	// Log provider auth fields for debugging
	fmt.Printf("ProviderAuth ID=%s: AppID='%s', PrivateKey length=%d, InstallationID=%d, IsActive=%t\n",
		EncodeFriendlyID(PrefixProviderAuth, providerAuth.ID), providerAuth.AppID, len(providerAuth.PrivateKey), providerAuth.InstallationID, providerAuth.IsActive)

	// Check if the auth is active
	if !providerAuth.IsActive {
		return SendError(c, http.StatusForbidden, "Provider auth is not active")
	}

	// Fetch repositories based on platform (e.g., GitHub)
	var repositories []map[string]interface{}
	switch providerAuth.Platform {
	case "github":
		var token string
		if providerAuth.AppID != "" && providerAuth.PrivateKey != "" && providerAuth.InstallationID != 0 {
			// Generate installation token for GitHub App
			token, err = utils.GenerateGitHubAppInstallationToken(providerAuth.AppID, providerAuth.PrivateKey, providerAuth.InstallationID)
			if err != nil {
				return SendError(c, http.StatusInternalServerError, "Failed to generate GitHub App token: "+err.Error())
			}
		} else {
			// Enhanced error message for debugging
			missingFields := []string{}
			if providerAuth.AppID == "" {
				missingFields = append(missingFields, "AppID")
			}
			if providerAuth.PrivateKey == "" {
				missingFields = append(missingFields, "PrivateKey")
			}
			if providerAuth.InstallationID == 0 {
				missingFields = append(missingFields, "InstallationID")
			}
			return SendError(c, http.StatusBadRequest, fmt.Sprintf("GitHub App credentials incomplete. Missing: %v. Please ensure the app is installed and InstallationID is updated.", missingFields))
		}
		repositories, err = fetchGitHubRepos(token)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to fetch repositories: "+err.Error())
		}
	default:
		return SendError(c, http.StatusBadRequest, "Unsupported platform")
	}

	return SendSuccess(c, repositories)
}

// fetchGitHubRepos 使用给定的 token 从 GitHub API 获取用户的仓库列表。
// 它返回一个 map 切片（每个 map 代表一个仓库）和一个 error（如果发生错误）。
func fetchGitHubRepos(token string) ([]map[string]interface{}, error) {
	// 定义 GitHub API 端点
	apiUrl := "https://api.github.com/installation/repositories" // Changed to installation endpoint for GitHub App

	// 创建一个新的 HTTP GET 请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("无法创建 GitHub API 请求: %w", err)
	}

	// 设置认证和 API 版本头
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 执行请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行 GitHub API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查 API 返回的状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误状态: %s", resp.Status)
	}

	// 定义一个临时结构体来安全地解析 JSON 响应
	type GitHubReposResponse struct {
		TotalCount   int `json:"total_count"`
		Repositories []struct {
			FullName string `json:"full_name"`
			URL      string `json:"html_url"`
		} `json:"repositories"`
	}

	// 将响应体解码到结构体中
	var response GitHubReposResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解码 GitHub API 响应失败: %w", err)
	}

	// 将解析后的数据转换为要求的 []map[string]interface{} 格式
	result := make([]map[string]interface{}, len(response.Repositories))
	for i, repo := range response.Repositories {
		result[i] = map[string]interface{}{
			"fullName": repo.FullName,
			"url":      repo.URL,
		}
	}

	// 成功，返回结果和 nil error
	return result, nil
}

// ListBranchesHandler fetches branches for a given repository
func ListBranchesHandler(c echo.Context) error {
	// Get provider auth ID from URL
	idStr := c.Param("uid")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid provider auth ID format")
	}

	// Get repo full name from query
	repo := c.QueryParam("repo")
	if repo == "" {
		return SendError(c, http.StatusBadRequest, "Repo full name is required")
	}

	// Retrieve provider auth from database
	providerAuth, err := models.GetProviderAuthByID(id)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Provider auth not found")
	}

	// Check if the auth is active
	if !providerAuth.IsActive {
		return SendError(c, http.StatusForbidden, "Provider auth is not active")
	}

	// Fetch branches based on platform
	var branches []map[string]interface{}
	switch providerAuth.Platform {
	case "github":
		if providerAuth.AppID == "" || providerAuth.PrivateKey == "" || providerAuth.InstallationID == 0 {
			return SendError(c, http.StatusBadRequest, "GitHub App credentials incomplete")
		}
		token, err := utils.GenerateGitHubAppInstallationToken(providerAuth.AppID, providerAuth.PrivateKey, providerAuth.InstallationID)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to generate GitHub App token: "+err.Error())
		}
		branches, err = fetchGitHubBranches(token, repo)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to fetch branches: "+err.Error())
		}
	default:
		return SendError(c, http.StatusBadRequest, "Unsupported platform")
	}

	return SendSuccess(c, branches)
}

// fetchGitHubBranches fetches branches for a specific repo
func fetchGitHubBranches(token, repo string) ([]map[string]interface{}, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/branches", repo)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("无法创建 GitHub API 请求: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行 GitHub API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误状态: %s", resp.Status)
	}

	var githubBranches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubBranches); err != nil {
		return nil, fmt.Errorf("解码 GitHub API 响应失败: %w", err)
	}

	result := make([]map[string]interface{}, len(githubBranches))
	for i, branch := range githubBranches {
		result[i] = map[string]interface{}{
			"name": branch.Name,
		}
	}
	return result, nil
}
