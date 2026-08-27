package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// setAuthCookies writes the access and refresh JWT tokens as HttpOnly, domain-wide cookies.
// Using HttpOnly prevents JavaScript from reading the tokens (XSS immunity).
// Setting the domain to the root domain (e.g. .puxbay.com) enables cross-subdomain SSO.
func setAuthCookies(c *gin.Context, tokens *services.TokenPair, rootDomain string, jwtCfg config.JWTConfig) {
	isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"

	// Cookie domain: prepend "." to share across all subdomains.
	// For localhost we use an empty string so the browser accepts it.
	cookieDomain := ""
	if isProduction && rootDomain != "" {
		// Strip port if present, then prepend dot.
		domain := rootDomain
		if idx := strings.LastIndex(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
		cookieDomain = "." + domain
	}

	accessMaxAge := int(jwtCfg.AccessExpiry.Seconds())
	refreshMaxAge := int(jwtCfg.RefreshExpiry.Seconds())

	if isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	// Access token cookie — shorter lived
	c.SetCookie("pux_session", tokens.AccessToken, accessMaxAge, "/", cookieDomain, isProduction, true)
	// Refresh token cookie — longer lived
	c.SetCookie("pux_refresh", tokens.RefreshToken, refreshMaxAge, "/", cookieDomain, isProduction, true)
}

// clearAuthCookies removes the session cookies from the browser.
func clearAuthCookies(c *gin.Context, rootDomain string) {
	isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"
	cookieDomain := ""
	if isProduction && rootDomain != "" {
		domain := rootDomain
		if idx := strings.LastIndex(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
		cookieDomain = "." + domain
	}

	if isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie("pux_session", "", -1, "/", cookieDomain, isProduction, true)
	c.SetCookie("pux_refresh", "", -1, "/", cookieDomain, isProduction, true)
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	db           *gorm.DB
	authService  *services.AuthService
	smtpCfg      config.SMTPConfig
	rootDomain   string
	jwtCfg       config.JWTConfig
	emailService *services.EmailService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(db *gorm.DB, authService *services.AuthService, smtpCfg config.SMTPConfig, rootDomain string, jwtCfg config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		db:           db,
		authService:  authService,
		smtpCfg:      smtpCfg,
		rootDomain:   rootDomain,
		jwtCfg:       jwtCfg,
		emailService: services.NewEmailService(db, smtpCfg),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totp_code,omitempty"`
}

// RegisterRequest is the request body for tenant registration.
type RegisterRequest struct {
	// User fields
	Username  string `json:"username" binding:"required,min=3,max=150"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`

	// Tenant fields
	CompanyName string `json:"company_name" binding:"required,min=2,max=100"`
	Subdomain   string `json:"subdomain" binding:"required,min=3,max=100,alphanum"`
}

// Login authenticates a user and returns JWT tokens.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	profile, tokens, err := h.authService.Authenticate(req.Username, req.Password, req.TOTPCode)
	if err != nil {
		if err.Error() == "2fa_required" {
			c.Error(middleware.NewAppError(http.StatusForbidden, "Two-factor authentication code required", err))
			return
		}
		c.Error(middleware.NewAppError(http.StatusUnauthorized, err.Error(), err))
		return
	}

	roleName := "staff"
	var perms []string = []string{}

	if profile.User.IsSuperuser {
		roleName = "superadmin"
		perms = []string{"*"}
		// Also fetch explicit admin user permissions if they exist
		var adminUser models.AdminUser
		if err := h.db.Where("user_id = ?", profile.User.ID).First(&adminUser).Error; err == nil {
			if adminUser.Permissions != "" && adminUser.Permissions != "{}" {
				var directPerms []string
				if err := json.Unmarshal([]byte(adminUser.Permissions), &directPerms); err == nil {
					perms = append(perms, directPerms...)
				}
			}
		}
	} else {
		// Check if this is an AdminUser from admin_users table
		var adminUser models.AdminUser
		if err := h.db.Preload("Role").Where("user_id = ?", profile.User.ID).First(&adminUser).Error; err == nil {
			if adminUser.Role != nil {
				roleName = adminUser.Role.Name
				if adminUser.Role.Permissions != "" {
					var rolePerms []string
					if err := json.Unmarshal([]byte(adminUser.Role.Permissions), &rolePerms); err == nil {
						perms = append(perms, rolePerms...)
					}
				}
			} else {
				roleName = "admin"
			}
			if adminUser.Permissions != "" && adminUser.Permissions != "{}" {
				var directPerms []string
				if err := json.Unmarshal([]byte(adminUser.Permissions), &directPerms); err == nil {
					perms = append(perms, directPerms...)
				}
			}
		} else if profile.Role != nil {
			roleName = profile.Role.Name
			perms, _ = h.authService.GetRolePermissions(&profile.Role.ID)
		}
	}

	// Set HttpOnly session cookies — tokens are NOT returned in the body.
	setAuthCookies(c, tokens, h.rootDomain, h.jwtCfg)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           profile.ID,
			"user_id":      profile.UserID,
			"tenant_id":    profile.TenantID,
			"branch_id":    profile.BranchID,
			"role":         roleName,
			"permissions":  perms,
			"is_superuser": profile.User.IsSuperuser,
			"username":     profile.User.Username,
			"email":        profile.User.Email,
			"first_name":   profile.User.FirstName,
			"last_name":    profile.User.LastName,
			"subdomain":    profile.Tenant.Subdomain,
		},
	})
}

// ChangeTemporaryPassword
// POST /api/v1/auth/change-temporary-password
func (h *AuthHandler) ChangeTemporaryPassword(c *gin.Context) {
	var req struct {
		Username          string `json:"username" binding:"required"`
		TemporaryPassword string `json:"temporary_password" binding:"required"`
		NewPassword       string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, tokens, err := h.authService.ChangeTemporaryPassword(req.Username, req.TemporaryPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set HttpOnly session cookies
	setAuthCookies(c, tokens, h.rootDomain, h.jwtCfg)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         profile.ID,
			"user_id":    profile.UserID,
			"tenant_id":  profile.TenantID,
			"branch_id":  profile.BranchID,
			"role":       profile.Role.Name,
			"username":   profile.User.Username,
			"email":      profile.User.Email,
			"first_name": profile.User.FirstName,
			"last_name":  profile.User.LastName,
			"subdomain":  profile.Tenant.Subdomain,
		},
	})
}

// RefreshToken silently issues a new access token using the pux_refresh HttpOnly cookie.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Read refresh token from HttpOnly cookie
	refreshToken, err := c.Cookie("pux_refresh")
	if err != nil || refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token cookie missing"})
		return
	}

	tokens, err := h.authService.RefreshAccessToken(refreshToken)
	if err != nil {
		clearAuthCookies(c, h.rootDomain)
		c.Error(middleware.NewAppError(http.StatusUnauthorized, "Invalid or expired refresh token", err))
		return
	}

	// Rotate cookies with fresh tokens
	setAuthCookies(c, tokens, h.rootDomain, h.jwtCfg)
	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
}

// Logout invalidates the current token.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Attempt to blacklist the session token if present
	if tokenString, err := c.Cookie("pux_session"); err == nil && tokenString != "" {
		_ = h.authService.Logout(c.Request.Context(), tokenString)
	}
	// Clear cookies from the browser
	clearAuthCookies(c, h.rootDomain)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// GetSession returns the authenticated user's profile if a valid pux_session cookie exists.
// This is used by the frontend to restore the session on page load / subdomain navigation.
// GET /api/v1/auth/session
func (h *AuthHandler) GetSession(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	profile, err := h.authService.CurrentUser(userID.(uuid.UUID), tenantID.(uuid.UUID))
	if err != nil {
		clearAuthCookies(c, h.rootDomain)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session invalid"})
		return
	}

	perms := c.GetStringSlice(middleware.ContextKeyPermissions)
	if profile.User.IsSuperuser {
		perms = []string{"*"}
	} else if len(perms) == 0 {
		var adminUser models.AdminUser
		if err := h.db.Preload("Role").Where("user_id = ?", profile.User.ID).First(&adminUser).Error; err == nil {
			if adminUser.Role != nil && adminUser.Role.Permissions != "" {
				var rolePerms []string
				if err := json.Unmarshal([]byte(adminUser.Role.Permissions), &rolePerms); err == nil {
					perms = append(perms, rolePerms...)
				}
			}
			if adminUser.Permissions != "" && adminUser.Permissions != "{}" {
				var directPerms []string
				if err := json.Unmarshal([]byte(adminUser.Permissions), &directPerms); err == nil {
					perms = append(perms, directPerms...)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        profile.ID,
		"user_id":   profile.UserID,
		"tenant_id": profile.TenantID,
		"branch_id": profile.BranchID,
		"role": func() string {
			if profile.Role != nil {
				return profile.Role.Name
			}
			return ""
		}(),
		"permissions":       perms,
		"is_superuser":      profile.User.IsSuperuser,
		"username":          profile.User.Username,
		"email":             profile.User.Email,
		"first_name":        profile.User.FirstName,
		"last_name":         profile.User.LastName,
		"is_2fa_enabled":    profile.Is2FAEnabled,
		"is_email_verified": profile.IsEmailVerified,
		"subdomain":         profile.Tenant.Subdomain,
	})
}

// CurrentUser returns the authenticated user's details.
// GET /api/v1/auth/user
func (h *AuthHandler) CurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	profile, err := h.authService.CurrentUser(userID.(uuid.UUID), tenantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        profile.ID,
		"user_id":   profile.UserID,
		"tenant_id": profile.TenantID,
		"branch_id": profile.BranchID,
		"role": func() string {
			if profile.Role != nil {
				return profile.Role.Name
			}
			return ""
		}(),
		"permissions":       c.GetStringSlice(middleware.ContextKeyPermissions),
		"username":          profile.User.Username,
		"email":             profile.User.Email,
		"first_name":        profile.User.FirstName,
		"last_name":         profile.User.LastName,
		"is_2fa_enabled":    profile.Is2FAEnabled,
		"is_email_verified": profile.IsEmailVerified,
		"subdomain":         profile.Tenant.Subdomain,
	})
}

// Register creates a new tenant with an admin user.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	input := services.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		CompanyName: req.CompanyName,
		Subdomain:   req.Subdomain,
	}

	if err := h.authService.Register(input); err != nil {
		// Differentiate between conflicts and server errors based on error message
		if err.Error() == "username or email already exists" || err.Error() == "subdomain is already taken" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created. Please check your email for a verification code.",
	})
}

// VerifyEmail confirms the user's email using a link token (from the button in the email).
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Also accept token as query param for magic-link clicks from email clients
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
			return
		}
		req.Token = token
	}

	var user models.User
	if err := h.db.Where("email_verification_token = ? AND email_verification_expiry > ?", req.Token, time.Now()).
		First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification link. Please request a new one."})
		return
	}

	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"is_email_verified":        true,
		"email_verification_token": nil,
		"email_verification_expiry": nil,
		"email_verification_code":  nil,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}

// VerifyEmailOTP confirms the user's email using the 6-digit OTP code.
// POST /api/v1/auth/verify-email-otp
func (h *AuthHandler) VerifyEmailOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and 6-digit code are required"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ? AND email_verification_code = ? AND email_verification_expiry > ?",
		req.Email, req.Code, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification code"})
		return
	}

	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"is_email_verified":         true,
		"email_verification_token":  nil,
		"email_verification_expiry": nil,
		"email_verification_code":   nil,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}

// ResendVerificationEmail resends the verification email/OTP to the user.
// POST /api/v1/auth/resend-verification
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "If that account exists, a new verification code has been sent."})
		return
	}

	if user.IsEmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This email is already verified"})
		return
	}

	// Generate fresh token + OTP
	randBytes := make([]byte, 32)
	if _, err := rand.Read(randBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	newToken := hex.EncodeToString(randBytes)
	otpN, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	newOTP := fmt.Sprintf("%06d", otpN.Int64())
	expiry := time.Now().Add(24 * time.Hour)

	h.db.Model(&user).Updates(map[string]interface{}{
		"email_verification_token":  newToken,
		"email_verification_code":   newOTP,
		"email_verification_expiry": expiry,
	})

	if h.emailService != nil {
		go h.emailService.SendEmailVerification(user.Email, user.FirstName, newToken, newOTP, "")
	}

	c.JSON(http.StatusOK, gin.H{"message": "A new verification code has been sent to your email."})
}

// ForgotPassword handles sending password reset emails.
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	emailService := services.NewEmailService(getDB(c, h.db), h.smtpCfg)
	_ = emailService.RequestPasswordReset(req.Email)

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
}

// ResetPassword handles resetting the password.
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	emailService := services.NewEmailService(getDB(c, h.db), h.smtpCfg)
	if err := emailService.ResetPassword(req.Token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// UpdateMeRequest defines the fields a user can update about themselves.
type UpdateMeRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UpdateMe allows the authenticated user to update their own profile info.
// PUT /api/v1/auth/me
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var user models.User
	if err := getDB(c, h.db).First(&user, "id = ?", userID.(uuid.UUID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update name fields if provided
	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}

	// Handle password change
	if req.NewPassword != "" {
		if req.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required to set a new password"})
			return
		}
		if !h.authService.CheckPassword(req.CurrentPassword, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		if len(req.NewPassword) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 8 characters"})
			return
		}
		hashed, err := h.authService.HashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		updates["password"] = hashed
		// Invalidate all existing tokens by incrementing token version
		updates["token_version"] = user.TokenVersion + 1
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updates provided"})
		return
	}

	if err := getDB(c, h.db).Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Profile updated successfully",
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"username":   user.Username,
	})
}
