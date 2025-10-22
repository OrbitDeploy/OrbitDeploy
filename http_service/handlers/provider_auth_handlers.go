package handlers

import (
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/labstack/echo/v4"
)

// ProviderAuthRequest 创建/更新ProviderAuth的请求结构
type ProviderAuthRequest struct {
	Platform       string `json:"platform"`
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
	RedirectURI    string `json:"redirectUri"`
	Username       string `json:"username"`
	AppPassword    string `json:"appPassword"`
	AppID          string `json:"appId"` // For GitHub Apps (required for GitHub platform)
	Slug           string `json:"slug"`  // For GitHub Apps
	PrivateKey     string `json:"privateKey"`
	WebhookSecret  string `json:"webhookSecret"`
	InstallationID uint   `json:"installationId"`
	Scopes         string `json:"scopes"`
	IsActive       bool   `json:"isActive"`
}

// ProviderAuthResponse ProviderAuth响应结构
type ProviderAuthResponse struct {
	Uid      string `json:"uid"`
	Platform string `json:"platform"`
	ClientID string `json:"clientId"`
	// 注意：不返回敏感信息ClientSecret、AppPassword、PrivateKey
	RedirectURI    string               `json:"redirectUri"`
	Username       string               `json:"username"`
	AppID          string               `json:"appId"`
	Slug           string               `json:"slug"`
	WebhookSecret  string               `json:"webhookSecret"`
	InstallationID uint                 `json:"installationId"`
	Scopes         string               `json:"scopes"`
	IsActive       bool                 `json:"isActive"`
	CreatedAt      string               `json:"createdAt"`
	UpdatedAt      string               `json:"updatedAt"`
	Application    *ApplicationResponse `json:"application,omitempty"`
}

// CreateProviderAuthHandler 创建新的第三方平台授权
func CreateProviderAuthHandler(c echo.Context) error {
	var req ProviderAuthRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "无效的请求格式: "+err.Error())
	}

	// 参数验证
	if req.Platform == "" {
		return SendError(c, http.StatusBadRequest, "平台类型不能为空")
	}

	// 根据平台类型验证必要字段
	switch req.Platform {
	case "github":
		// GitHub仅支持GitHub Apps方式（不再支持OAuth Apps）
		if req.AppID == "" || req.PrivateKey == "" {
			return SendError(c, http.StatusBadRequest, "GitHub平台仅支持GitHub Apps，需要提供AppID和PrivateKey")
		}
	case "gitlab", "gitea":
		if req.ClientID == "" || req.ClientSecret == "" {
			return SendError(c, http.StatusBadRequest, "OAuth平台需要提供ClientID和ClientSecret")
		}
	case "bitbucket":
		if req.Username == "" || req.AppPassword == "" {
			return SendError(c, http.StatusBadRequest, "Bitbucket需要提供用户名和应用密码")
		}
	default:
		return SendError(c, http.StatusBadRequest, "不支持的平台类型: "+req.Platform)
	}

	// 创建ProviderAuth
	providerAuth, err := models.CreateProviderAuth(
		req.Platform,
		req.ClientID,
		req.ClientSecret,
		req.RedirectURI,
		req.Username,
		req.AppPassword,
		req.AppID,
		req.Slug,
		req.PrivateKey,
		req.WebhookSecret,
		req.InstallationID,
		req.Scopes,
	)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "创建ProviderAuth失败: "+err.Error())
	}

	// 构造响应
	response := &ProviderAuthResponse{
		Uid:            EncodeFriendlyID(PrefixProviderAuth, providerAuth.ID),
		Platform:       providerAuth.Platform,
		ClientID:       providerAuth.ClientID,
		RedirectURI:    providerAuth.RedirectURI,
		Username:       providerAuth.Username,
		AppID:          providerAuth.AppID,
		Slug:           providerAuth.Slug,
		WebhookSecret:  providerAuth.WebhookSecret,
		InstallationID: providerAuth.InstallationID,
		Scopes:         providerAuth.Scopes,
		IsActive:       providerAuth.IsActive,
		CreatedAt:      providerAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      providerAuth.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return SendCreated(c, response)
}

// ListProviderAuthsHandler 获取用户的所有授权记录
func ListProviderAuthsHandler(c echo.Context) error {
	// 获取平台过滤参数（可选）
	platform := c.QueryParam("platform")

	var providerAuths []*models.ProviderAuth
	var err error
	if platform != "" {
		providerAuths, err = models.ListProviderAuthsByPlatform(platform)
	} else {
		providerAuths, err = models.ListProviderAuths()
	}

	if err != nil {
		return SendError(c, http.StatusInternalServerError, "获取授权列表失败: "+err.Error())
	}

	// 构造响应列表
	var responses []*ProviderAuthResponse
	for _, pa := range providerAuths {
		response := &ProviderAuthResponse{
			Uid:            EncodeFriendlyID(PrefixProviderAuth, pa.ID),
			Platform:       pa.Platform,
			ClientID:       pa.ClientID,
			RedirectURI:    pa.RedirectURI,
			Username:       pa.Username,
			AppID:          pa.AppID,
			Slug:           pa.Slug,
			WebhookSecret:  pa.WebhookSecret,
			InstallationID: pa.InstallationID,
			Scopes:         pa.Scopes,
			IsActive:       pa.IsActive,
			CreatedAt:      pa.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:      pa.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		responses = append(responses, response)
	}

	return SendSuccess(c, responses)
}

// GetProviderAuthHandler 获取特定的授权记录
func GetProviderAuthHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "无效的ID格式")
	}

	providerAuth, err := models.GetProviderAuthByID(id)
	if err != nil {
		return SendError(c, http.StatusNotFound, "ProviderAuth不存在")
	}

	// 构造响应
	response := &ProviderAuthResponse{
		Uid:            EncodeFriendlyID(PrefixProviderAuth, providerAuth.ID),
		Platform:       providerAuth.Platform,
		ClientID:       providerAuth.ClientID,
		RedirectURI:    providerAuth.RedirectURI,
		Username:       providerAuth.Username,
		AppID:          providerAuth.AppID,
		Slug:           providerAuth.Slug,
		WebhookSecret:  providerAuth.WebhookSecret,
		InstallationID: providerAuth.InstallationID,
		Scopes:         providerAuth.Scopes,
		IsActive:       providerAuth.IsActive,
		CreatedAt:      providerAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      providerAuth.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return SendSuccess(c, response)
}

// UpdateProviderAuthHandler 更新授权记录
func UpdateProviderAuthHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "无效的ID格式")
	}

	var req ProviderAuthRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "无效的请求格式: "+err.Error())
	}

	// 更新ProviderAuth
	providerAuth, err := models.UpdateProviderAuth(
		id,
		req.ClientID,
		req.ClientSecret,
		req.RedirectURI,
		req.Username,
		req.AppPassword,
		req.AppID,
		req.Slug,
		req.PrivateKey,
		req.WebhookSecret,
		req.InstallationID,
		req.Scopes,
		req.IsActive,
	)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "更新ProviderAuth失败: "+err.Error())
	}

	// 构造响应
	response := &ProviderAuthResponse{
		Uid:            EncodeFriendlyID(PrefixProviderAuth, providerAuth.ID),
		Platform:       providerAuth.Platform,
		ClientID:       providerAuth.ClientID,
		RedirectURI:    providerAuth.RedirectURI,
		Username:       providerAuth.Username,
		AppID:          providerAuth.AppID,
		Slug:           providerAuth.Slug,
		WebhookSecret:  providerAuth.WebhookSecret,
		InstallationID: providerAuth.InstallationID,
		Scopes:         providerAuth.Scopes,
		IsActive:       providerAuth.IsActive,
		CreatedAt:      providerAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      providerAuth.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return SendSuccess(c, response)
}

// DeleteProviderAuthHandler 删除授权记录
func DeleteProviderAuthHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "无效的ID格式")
	}

	if err := models.DeleteProviderAuth(id); err != nil {
		return SendError(c, http.StatusInternalServerError, "删除ProviderAuth失败: "+err.Error())
	}

	return SendSuccess(c, map[string]string{"message": "ProviderAuth删除成功"})
}

// ActivateProviderAuthHandler 激活授权记录
func ActivateProviderAuthHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "无效的ID格式")
	}

	if err := models.ActivateProviderAuth(id); err != nil {
		return SendError(c, http.StatusInternalServerError, "激活ProviderAuth失败: "+err.Error())
	}

	return SendSuccess(c, map[string]string{"message": "ProviderAuth激活成功"})
}

// DeactivateProviderAuthHandler 停用授权记录
func DeactivateProviderAuthHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := DecodeFriendlyID(PrefixProviderAuth, idStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "无效的ID格式")
	}

	if err := models.DeactivateProviderAuth(id); err != nil {
		return SendError(c, http.StatusInternalServerError, "停用ProviderAuth失败: "+err.Error())
	}

	return SendSuccess(c, map[string]string{"message": "ProviderAuth停用成功"})
}
