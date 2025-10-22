package utils

// GitAuthConfig contains configuration for Git authentication
type GitAuthConfig struct {
	Token      string
	Username   string // 通常为GitHub用户名或 "x-access-token"
	AuthMethod string // "token", "ssh"
}

// BuildAuthenticatedGitURL constructs a Git URL with embedded authentication
// func BuildAuthenticatedGitURL(repoURL string, config GitAuthConfig) (string, error) {
// 	// 解析原始URL
// 	parsedURL, err := url.Parse(repoURL)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid repository URL: %w", err)
// 	}

// 	// 仅支持HTTPS的GitHub URL
// 	if parsedURL.Scheme != "https" {
// 		return "", fmt.Errorf("only HTTPS GitHub URLs are supported for token authentication")
// 	}

// 	if !strings.Contains(parsedURL.Host, "github.com") {
// 		return "", fmt.Errorf("only GitHub repositories are supported")
// 	}

// 	// 构建带认证的URL
// 	switch config.AuthMethod {
// 	case "token":
// 		if config.Token == "" {
// 			return "", fmt.Errorf("token is required for token authentication")
// 		}
// 		// 格式：https://username:token@github.com/user/repo.git
// 		username := config.Username
// 		if username == "" {
// 			username = "x-access-token"
// 		}
// 		authenticatedURL := fmt.Sprintf("https://%s:%s@%s%s",
// 			username, config.Token, parsedURL.Host, parsedURL.Path)
// 		return authenticatedURL, nil

// 	default:
// 		return "", fmt.Errorf("unsupported authentication method: %s", config.AuthMethod)
// 	}
// }

// SanitizeGitURL removes authentication information from a Git URL for logging
// func SanitizeGitURL(gitURL string) string {
// 	parsedURL, err := url.Parse(gitURL)
// 	if err != nil {
// 		return gitURL // 如果解析失败，返回原URL
// 	}

// 	// 移除用户信息
// 	parsedURL.User = nil
// 	return parsedURL.String()
// }

// // ValidateGitHubURL validates if a URL is a valid GitHub repository URL
// func ValidateGitHubURL(repoURL string) error {
// 	if repoURL == "" {
// 		return fmt.Errorf("repository URL is required")
// 	}

// 	parsedURL, err := url.Parse(repoURL)
// 	if err != nil {
// 		return fmt.Errorf("invalid URL format: %w", err)
// 	}

// 	if parsedURL.Scheme != "https" {
// 		return fmt.Errorf("only HTTPS URLs are supported")
// 	}

// 	if !strings.Contains(parsedURL.Host, "github.com") {
// 		return fmt.Errorf("only GitHub repositories are supported")
// 	}

// 	// 检查路径格式：应该是 /user/repo 或 /user/repo.git
// 	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
// 	if len(pathParts) != 2 {
// 		return fmt.Errorf("invalid GitHub repository path format, expected /user/repo")
// 	}

// 	return nil
// }

// // ExtractRepoInfo extracts owner and repository name from GitHub URL
// func ExtractRepoInfo(repoURL string) (owner, repo string, err error) {
// 	parsedURL, err := url.Parse(repoURL)
// 	if err != nil {
// 		return "", "", fmt.Errorf("invalid URL format: %w", err)
// 	}

// 	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
// 	if len(pathParts) != 2 {
// 		return "", "", fmt.Errorf("invalid GitHub repository path format")
// 	}

// 	owner = pathParts[0]
// 	repo = strings.TrimSuffix(pathParts[1], ".git")

// 	return owner, repo, nil
// }
