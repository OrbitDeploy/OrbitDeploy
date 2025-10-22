package handlers

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// --- 2FA Setup ---

// TwoFASetupResponse defines the structure for the 2FA setup response.
type TwoFASetupResponse struct {
	Secret          string `json:"secret"`
	EncryptedSecret string `json:"encrypted_secret"`
	QRCodeURL       string `json:"qr_code_url"`
}

// to2FASetupResponse converts a TOTP key into a TwoFASetupResponse.
func to2FASetupResponse(key *otp.Key, encryptedSecret string) *TwoFASetupResponse {
	return &TwoFASetupResponse{
		Secret:          key.Secret(),
		EncryptedSecret: encryptedSecret,
		QRCodeURL:       key.URL(),
	}
}

// Setup2FA generates a new 2FA secret for the user.
func Setup2FA(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Unauthorized")
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "User not found.")
	}

	// Security check: Prevent re-enabling 2FA if it's already active.
	if user.TwoFactorEnabled {
		return SendError(c, http.StatusBadRequest, "2FA is already enabled.")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OrbitDeploy",
		AccountName: user.Username,
	})
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate 2FA key.")
	}

	encryptedSecret, err := utils.EncryptValue(key.Secret())
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to encrypt 2FA secret.")
	}

	return SendSuccess(c, to2FASetupResponse(key, encryptedSecret))
}

// --- 2FA Verification ---

// Verify2FARequest defines the request body for verifying a 2FA setup.
type Verify2FARequest struct {
	OTP             string `json:"otp"`
	EncryptedSecret string `json:"encrypted_secret"`
}

// Verify2FA verifies the OTP and enables 2FA for the user.
func Verify2FA(c echo.Context) error {
	var req Verify2FARequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body.")
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Unauthorized")
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "User not found.")
	}

	decryptedSecret, err := utils.DecryptValue(req.EncryptedSecret)
	if err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid secret.")
	}

	valid := totp.Validate(req.OTP, decryptedSecret)
	if !valid {
		return SendError(c, http.StatusBadRequest, "Invalid OTP.")
	}

	// Generate recovery codes before updating the user, in case of failure.
	recoveryCodes, hashedCodes, err := generateRecoveryCodes(10, 12)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate recovery codes.")
	}

	// Persist changes to the database
	// Enable 2FA for the logged-in user
	err = models.Enable2FAForUser(user.ID, req.EncryptedSecret, hashedCodes)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to enable 2FA.")
	}

	return SendSuccess(c, echo.Map{"recovery_codes": recoveryCodes})
}

// --- 2FA Login & Disable ---

// Login2FARequest defines the request body for the second step of 2FA login.

type Login2FARequest struct {
	Temp2FAToken string `json:"temp_2fa_token"`
	OTP          string `json:"otp"`
}

// Login2FA validates the OTP during the login process.
func Login2FA(c echo.Context) error {
	var req Login2FARequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body.")
	}

	jwtService := services.GetJWTService()
	claims, err := jwtService.Verify2FAToken(req.Temp2FAToken)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Invalid or expired 2FA token.")
	}

	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "User not found.")
	}

	decryptedSecret, err := utils.DecryptValue(user.TwoFactorSecret)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to process 2FA secret.")
	}

	// First, try to validate as a TOTP
	otpValid := totp.Validate(req.OTP, decryptedSecret)

	// If OTP is not valid, try to validate as a recovery code
	if !otpValid {
		recoveryValid, err := models.UseRecoveryCode(user.ID, req.OTP)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, "Failed to process recovery code.")
		}
		if !recoveryValid {
			return SendError(c, http.StatusBadRequest, "Invalid OTP or recovery code.")
		}
	}

	// --- Token Generation (copied from original Login handler) ---
	tokens, err := jwtService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	refreshTokenHash := jwtService.HashRefreshToken(tokens.RefreshToken)
	_, err = models.CreateAuthToken(refreshTokenHash, "Web Browser", time.Now().Add(720*time.Hour)) // 30 days
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to store refresh token")
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(720 * time.Hour), // 30 days
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return SendSuccess(c, map[string]interface{}{
		"access_token": tokens.AccessToken,
	})
}

// Disable2FARequest defines the request body for disabling 2FA.
type Disable2FARequest struct {
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

// Disable2FA disables 2FA for the user after verifying password and OTP.
func Disable2FA(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return SendError(c, http.StatusUnauthorized, "Unauthorized")
	}

	var req Disable2FARequest
	if err := c.Bind(&req); err != nil {
		return SendError(c, http.StatusBadRequest, "Invalid request body.")
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		return SendError(c, http.StatusNotFound, "User not found.")
	}

	// Security check: Verify user's current password.
	if !user.CheckPassword(req.Password) {
		return SendError(c, http.StatusUnauthorized, "Invalid password.")
	}

	// Security check: Verify a current OTP.
	decryptedSecret, err := utils.DecryptValue(user.TwoFactorSecret)
	if err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to process 2FA secret.")
	}

	valid := totp.Validate(req.OTP, decryptedSecret)
	if !valid {
		return SendError(c, http.StatusBadRequest, "Invalid OTP.")
	}

	// Proceed with disabling 2FA.
	if err := models.Disable2FAForUser(userID); err != nil {
		return SendError(c, http.StatusInternalServerError, "Failed to disable 2FA.")
	}

	return SendSuccess(c, echo.Map{"message": "2FA disabled successfully."})
}

// --- Helpers ---

const recoveryCodeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRecoveryCodes creates a set of unique recovery codes.
func generateRecoveryCodes(count, length int) (plainCodes []string, hashedCodes []string, err error) {
	plainCodes = make([]string, count)
	hashedCodes = make([]string, count)
	codeBytes := make([]byte, length)

	for i := 0; i < count; i++ {
		_, err = rand.Read(codeBytes)
		if err != nil {
			return nil, nil, err
		}
		for j := 0; j < length; j++ {
			codeBytes[j] = recoveryCodeChars[int(codeBytes[j])%len(recoveryCodeChars)]
		}
		plainCodes[i] = string(codeBytes)

		hashed, hashErr := bcrypt.GenerateFromPassword(codeBytes, bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, nil, hashErr
		}
		hashedCodes[i] = string(hashed)
	}

	return plainCodes, hashedCodes, nil
}
