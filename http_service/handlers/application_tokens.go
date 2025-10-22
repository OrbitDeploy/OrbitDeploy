package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/youfun/OrbitDeploy/models"
)

// ApplicationTokenRequest represents the request body for creating/updating application tokens
type ApplicationTokenRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateApplicationToken handles POST /api/apps/:appId/tokens
func CreateApplicationToken(c echo.Context) error {
	// 获取应用ID
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	// 验证应用是否存在
	app, err := models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// 解析请求体
	var req ApplicationTokenRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// 验证请求
	if req.Name == "" {
		return SendError(c, http.StatusBadRequest, "Token name is required")
	}

	// 创建token
	token, plainToken, err := models.CreateApplicationToken(app.ID, req.Name, req.ExpiresAt)
	if err != nil {
		logman.Error("Failed to create application token", err)
		return SendError(c, http.StatusInternalServerError, "Failed to create token")
	}

	// 构造响应
	response := CreateApplicationTokenResponse{
		ApplicationTokenResponse: ApplicationTokenResponse{
			Uid:            EncodeFriendlyID(PrefixAppToken, token.ID),
			ApplicationUid: EncodeFriendlyID(PrefixApplication, token.ApplicationID),
			Name:           token.Name,
			ExpiresAt:      token.ExpiresAt,
			LastUsedAt:     token.LastUsedAt,
			IsActive:       token.IsActive,
			CreatedAt:      token.CreatedAt,
			UpdatedAt:      token.UpdatedAt,
		},
		Token: plainToken,
	}

	return SendSuccess(c, response)
}

// ListApplicationTokens handles GET /api/apps/:appId/tokens
func ListApplicationTokens(c echo.Context) error {
	// 获取应用ID
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	// 验证应用是否存在
	_, err = models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// 获取tokens
	tokens, err := models.GetApplicationTokensByAppID(appID)
	if err != nil {
		logman.Error("Failed to list application tokens", err)
		return SendError(c, http.StatusInternalServerError, "Failed to list tokens")
	}

	// 转换为响应格式
	var responses []ApplicationTokenResponse
	for _, token := range tokens {
		responses = append(responses, ApplicationTokenResponse{
			Uid:            EncodeFriendlyID(PrefixAppToken, token.ID),
			ApplicationUid: EncodeFriendlyID(PrefixApplication, token.ApplicationID),
			Name:           token.Name,
			ExpiresAt:      token.ExpiresAt,
			LastUsedAt:     token.LastUsedAt,
			IsActive:       token.IsActive,
			CreatedAt:      token.CreatedAt,
			UpdatedAt:      token.UpdatedAt,
		})
	}

	return SendSuccess(c, responses)
}

// UpdateApplicationToken handles PUT /api/apps/:appId/tokens/:tokenId
func UpdateApplicationToken(c echo.Context) error {
	// 获取应用ID和tokenID
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	tokenIDStr := c.Param("tokenId")
	tokenID, err := DecodeFriendlyID(PrefixAppToken, tokenIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid token ID format")
	}

	// 验证应用是否存在
	_, err = models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// 解析请求体
	var req ApplicationTokenRequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// 验证请求
	if req.Name == "" {
		return SendError(c, http.StatusBadRequest, "Token name is required")
	}

	// 更新token
	if err := models.UpdateApplicationToken(tokenID, appID, req.Name, req.ExpiresAt); err != nil {
		logman.Error("Failed to update application token", err)
		return SendError(c, http.StatusInternalServerError, "Failed to update token")
	}

	// 获取更新后的token
	token, _, err := models.GetApplicationTokenByID(tokenID, appID, false)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Token not found")
	}

	response := ApplicationTokenResponse{
		Uid:            EncodeFriendlyID(PrefixAppToken, token.ID),
		ApplicationUid: EncodeFriendlyID(PrefixApplication, token.ApplicationID),
		Name:           token.Name,
		ExpiresAt:      token.ExpiresAt,
		LastUsedAt:     token.LastUsedAt,
		IsActive:       token.IsActive,
		CreatedAt:      token.CreatedAt,
		UpdatedAt:      token.UpdatedAt,
	}

	return SendSuccess(c, response)
}

// DeleteApplicationToken handles DELETE /api/apps/:appId/tokens/:tokenId
func DeleteApplicationToken(c echo.Context) error {
	// 获取应用ID和tokenID
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	tokenIDStr := c.Param("tokenId")
	tokenID, err := DecodeFriendlyID(PrefixAppToken, tokenIDStr)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid token ID format")
	}

	// 验证应用是否存在
	_, err = models.GetApplicationByID(appID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "Application not found")
	}

	// 删除token
	if err := models.DeleteApplicationToken(tokenID, appID); err != nil {
		logman.Error("Failed to delete application token", err)
		return SendError(c, http.StatusInternalServerError, "Failed to delete token")
	}

	return SendSuccess(c, map[string]string{"message": "Token deleted successfully"})
}
