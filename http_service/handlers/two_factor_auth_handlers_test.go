package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/dborm"
	"github.com/stretchr/testify/assert"
	"github.com/youfun/OrbitDeploy/models"
)

// setupTestDB configures a temporary in-memory sqlite database for tests.
func setupTestDB(t *testing.T) {
	tmpDir := t.TempDir()
	config := &dborm.Config{
		Type:   "sqlite",
		DbName: tmpDir + "/test.db",
	}
	if dborm.Connect(config) == nil {
		t.Fatal("failed to connect to test database")
	}

	// Migrate all necessary models
	err := dborm.Db.AutoMigrate(&models.User{}, &models.TwoFactorRecoveryCode{})
	assert.NoError(t, err)
}

// TestSetup2FA_Success tests the successful generation of a 2FA secret.
func TestSetup2FA_Success(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer dborm.Destroy()

	// Create a test user
	testUser, err := models.CreateUser("testuser", "password123")
	assert.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/setup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock authenticated user in context
	c.Set("userID", testUser.ID)

	// Execute the handler
	err = Setup2FA(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Define a struct to parse the whole response
	type APIResponse struct {
		Success bool                 `json:"success"`
		Data    TwoFASetupResponse `json:"data"`
	}

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check success field and data content
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data.Secret)
	assert.NotEmpty(t, response.Data.EncryptedSecret)
	assert.NotEmpty(t, response.Data.QRCodeURL)
	assert.Contains(t, response.Data.QRCodeURL, "otpauth://totp/OrbitDeploy:testuser")
}

// TestVerify2FA_Success tests the successful verification of a 2FA setup.
func TestVerify2FA_Success(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer dborm.Destroy()

	user, err := models.CreateUser("testuser2fa", "password123")
	assert.NoError(t, err)

	// 1. Simulate the `Setup` step to get a secret
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "OrbitDeploy", AccountName: user.Username})
	assert.NoError(t, err)
	encryptedSecret, err := utils.EncryptValue(key.Secret())
	assert.NoError(t, err)

	// 2. Generate a valid OTP from the secret
	validOTP, err := totp.GenerateCode(key.Secret(), time.Now())
	assert.NoError(t, err)

	// 3. Prepare request for `Verify` step
	body := fmt.Sprintf(`{"otp":"%s", "encrypted_secret":"%s"}`, validOTP, encryptedSecret)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/verify", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", user.ID)

	// Execute handler
	assert.NoError(t, Verify2FA(c))

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	recoveryCodes := data["recovery_codes"].([]interface{})
	assert.Len(t, recoveryCodes, 10) // Check if 10 recovery codes are returned

	// 4. Verify database state
	updatedUser, err := models.GetUserByID(user.ID)
	assert.NoError(t, err)
	assert.True(t, updatedUser.TwoFactorEnabled)
	assert.Equal(t, encryptedSecret, updatedUser.TwoFactorSecret)

	var recoveryCodeCount int64
	dborm.Db.Model(&models.TwoFactorRecoveryCode{}).Where("user_id = ?", user.ID).Count(&recoveryCodeCount)
	assert.Equal(t, int64(10), recoveryCodeCount)
}

// TestVerify2FA_InvalidOTP tests the case where an invalid OTP is provided.
func TestVerify2FA_InvalidOTP(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer dborm.Destroy()

	user, err := models.CreateUser("testuserinvalid", "password123")
	assert.NoError(t, err)

	// 1. Simulate the `Setup` step to get a secret
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "OrbitDeploy", AccountName: user.Username})
	assert.NoError(t, err)
	encryptedSecret, err := utils.EncryptValue(key.Secret())
	assert.NoError(t, err)

	// 2. Prepare request with an invalid OTP
	body := fmt.Sprintf(`{"otp":"000000", "encrypted_secret":"%s"}`, encryptedSecret)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/verify", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", user.ID)

	// Execute handler
	assert.NoError(t, Verify2FA(c))

	// Assertions
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Invalid OTP.", response["message"])

	// 4. Verify database state has not changed
	updatedUser, err := models.GetUserByID(user.ID)
	assert.NoError(t, err)
	assert.False(t, updatedUser.TwoFactorEnabled)
}

// TestSetup2FA_UserNotFound tests the case where the user ID in context is invalid.
func TestSetup2FA_UserNotFound(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer dborm.Destroy()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/2fa/setup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock an invalid user ID in context
	c.Set("userID", uint(999))

	// Execute the handler
	err = Setup2FA(c)

	// Assertions
	assert.NoError(t, err) // The helper handles the error response, so handler returns nil
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Check response body for error message
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User not found.", response["message"])
}