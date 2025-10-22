package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/dborm"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/youfun/OrbitDeploy/models"
	"github.com/youfun/OrbitDeploy/services"
	"github.com/youfun/OrbitDeploy/utils"
)

// TestLogin_Non2FAUser_Success tests standard login for a user without 2FA.
func TestLogin_Non2FAUser_Success(t *testing.T) {
	setupTestDB(t)
	defer dborm.Destroy()

	_, err := models.CreateUser("normaluser", "password123")
	assert.NoError(t, err)

	body := `{"username":"normaluser", "password":"password123"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, Login(c))

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.Nil(t, data["two_factor_required"]) // Ensure 2FA fields are not present
}

// TestLogin_2FAUser_Step1_Success tests the first step of logging in for a 2FA-enabled user.
func TestLogin_2FAUser_Step1_Success(t *testing.T) {
	setupTestDB(t)
	defer dborm.Destroy()

	user, err := models.CreateUser("2fauser", "password123")
	assert.NoError(t, err)
	// Manually enable 2FA for the test user
	user.TwoFactorEnabled = true
	user.TwoFactorSecret = "dummy_secret"
	dborm.Db.Save(user)

	body := `{"username":"2fauser", "password":"password123"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, Login(c))

	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.True(t, data["two_factor_required"].(bool))
	assert.NotEmpty(t, data["temp_2fa_token"])
	assert.Nil(t, data["access_token"]) // Ensure final token is not present
}

// TestLogin2FA_Step2_ValidOTP_Success tests the second step of 2FA login with a valid OTP.
func TestLogin2FA_Step2_ValidOTP_Success(t *testing.T) {
	setupTestDB(t)
	defer dborm.Destroy()

	// 1. Create user and enable 2FA
	user, err := models.CreateUser("2fauser_step2", "password123")
	assert.NoError(t, err)
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "OrbitDeploy", AccountName: user.Username})
	assert.NoError(t, err)
	encryptedSecret, err := utils.EncryptValue(key.Secret())
	assert.NoError(t, err)
	assert.NoError(t, models.Enable2FAForUser(user.ID, encryptedSecret, nil))

	// 2. Get a temp token from Step 1
	jwtService := services.GetJWTService()
	tempToken, err := jwtService.Generate2FAToken(user.ID, user.Username)
	assert.NoError(t, err)

	// 3. Generate a valid OTP
	validOTP, err := totp.GenerateCode(key.Secret(), time.Now())
	assert.NoError(t, err)

	// 4. Call Login2FA handler
	body := fmt.Sprintf(`{"temp_2fa_token":"%s", "otp":"%s"}`, tempToken, validOTP)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, Login2FA(c))

	// 5. Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotNil(t, rec.Result().Header.Get("Set-Cookie")) // Check for refresh token cookie
}

// TestLogin2FA_Step2_ValidRecoveryCode_Success tests the second step with a recovery code.
func TestLogin2FA_Step2_ValidRecoveryCode_Success(t *testing.T) {
	setupTestDB(t)
	defer dborm.Destroy()

	// 1. Create user and enable 2FA with recovery codes
	user, err := models.CreateUser("2fauser_recovery", "password123")
	assert.NoError(t, err)
	plainCodes, hashedCodes, err := generateRecoveryCodes(1, 12) // Generate one code for test
	assert.NoError(t, err)
	assert.NoError(t, models.Enable2FAForUser(user.ID, "dummy_secret", hashedCodes))

	// 2. Get a temp token
	jwtService := services.GetJWTService()
	tempToken, err := jwtService.Generate2FAToken(user.ID, user.Username)
	assert.NoError(t, err)

	// 3. Call Login2FA handler with the recovery code
	body := fmt.Sprintf(`{"temp_2fa_token":"%s", "otp":"%s"}`, tempToken, plainCodes[0])
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, Login2FA(c))

	// 4. Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])

	// 5. Verify recovery code is marked as used
	var usedCode models.TwoFactorRecoveryCode
	dborm.Db.First(&usedCode, "user_id = ?", user.ID)
	assert.True(t, usedCode.Used)
}
