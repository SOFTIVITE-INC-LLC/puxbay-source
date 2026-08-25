package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication, user registration, JWT generation and password hashing.
type AuthService struct {
	db            *gorm.DB
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	tokenStore    TokenStore
	rootDomain    string
	permsCache    sync.Map
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access"`
	RefreshToken string `json:"refresh"`
}

// Claims represents JWT token claims.
type Claims struct {
	UserID     uuid.UUID  `json:"user_id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	BranchID   *uuid.UUID `json:"branch_id,omitempty"`
	SupplierID *uuid.UUID `json:"supplier_id,omitempty"` // Added for Supplier Portal
	RoleID     *uuid.UUID `json:"role_id,omitempty"`
	Role       string     `json:"role"`
	TokenType  string     `json:"token_type"` // "access" or "refresh"
	Version    int        `json:"version"`
	jwt.RegisteredClaims
}



func NewAuthService(cfg *config.JWTConfig, db *gorm.DB, tokenStore TokenStore, rootDomain string) *AuthService {
	if tokenStore == nil {
		tokenStore = &NoopTokenStore{}
	}
	return &AuthService{
		db:            db,
		jwtSecret:     []byte(cfg.Secret),
		accessExpiry:  cfg.AccessExpiry,
		refreshExpiry: cfg.RefreshExpiry,
		tokenStore:    tokenStore,
		rootDomain:    rootDomain,
	}
}

// GetRolePermissions fetches and caches permissions for a role ID
func (s *AuthService) GetRolePermissions(roleID *uuid.UUID) ([]string, error) {
	if roleID == nil {
		return []string{}, nil
	}

	if val, ok := s.permsCache.Load(*roleID); ok {
		return val.([]string), nil
	}

	var permissions []string
	err := s.db.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", *roleID).
		Pluck("permissions.code", &permissions).Error

	if err == nil {
		s.permsCache.Store(*roleID, permissions)
	}

	return permissions, err
}

// ClearPermissionsCache clears the in-memory permissions cache,
// forcing subsequent requests to re-fetch from the database.
func (s *AuthService) ClearPermissionsCache() {
	s.permsCache.Range(func(key, value any) bool {
		s.permsCache.Delete(key)
		return true
	})
}

// GenerateTokenPair creates a new access + refresh token pair.
func (s *AuthService) GenerateTokenPair(userID, tenantID uuid.UUID, branchID *uuid.UUID, role string, roleID *uuid.UUID, version int) (*TokenPair, error) {
	accessToken, err := s.generateToken(userID, tenantID, branchID, role, roleID, "access", version, s.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, tenantID, branchID, role, roleID, "refresh", version, s.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken parses and validates a JWT token.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token.
func (s *AuthService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	// Token Rotation: Deny the old refresh token
	if claims.ID != "" {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			_ = s.tokenStore.DenyToken(context.Background(), claims.ID, ttl)
		}
	}

	// Global Logout check: Verify the token version against the DB
	var user models.User
	if err := s.db.Select("token_version").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}
	if user.TokenVersion != claims.Version {
		return nil, errors.New("token version invalidated")
	}

	return s.GenerateTokenPair(claims.UserID, claims.TenantID, claims.BranchID, claims.Role, claims.RoleID, user.TokenVersion)
}

// Logout invalidates the access token by adding its JTI to the Redis denylist.
// The entry TTL matches the token's remaining lifetime.
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// Token already invalid — treat as successful logout
		return nil
	}
	if claims.ID == "" {
		return errors.New("token has no JTI")
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Already expired — nothing to denylist
		return nil
	}
	return s.tokenStore.DenyToken(ctx, claims.ID, ttl)
}

// IsTokenDenied checks whether a token's JTI has been added to the denylist.
func (s *AuthService) IsTokenDenied(ctx context.Context, jti string) bool {
	return s.tokenStore.IsDenied(ctx, jti)
}

// HashPassword hashes a plaintext password using bcrypt.
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// validateTOTPCode performs TOTP validation (RFC 6238).
func (s *AuthService) validateTOTPCode(secret, code string) bool {
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil || len(secretBytes) == 0 {
		return false
	}

	now := time.Now()
	for _, offset := range []int64{-30, 0, 30} {
		t := now.Add(time.Duration(offset) * time.Second)
		counter := uint64(t.Unix()) / 30
		expected := s.hmacBasedOTP(secretBytes, counter)
		if expected == code {
			return true
		}
	}
	return false
}

func (s *AuthService) hmacBasedOTP(key []byte, counter uint64) string {
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(counter & 0xff)
		counter >>= 8
	}

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	binCode := (uint32(h[offset])&0x7f)<<24 |
		(uint32(h[offset+1])&0xff)<<16 |
		(uint32(h[offset+2])&0xff)<<8 |
		(uint32(h[offset+3]) & 0xff)
	otp := binCode % 1000000

	return fmt.Sprintf("%06d", otp)
}

func (s *AuthService) generateToken(userID, tenantID uuid.UUID, branchID *uuid.UUID, role string, roleID *uuid.UUID, tokenType string, version int, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID,
		TenantID:    tenantID,
		BranchID:    branchID,
		RoleID:      roleID,
		Role:        role,
		TokenType:   tokenType,
		Version:     version,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(300 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "puxbay",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateSupplierToken creates a short-lived access token for a supplier portal session.
func (s *AuthService) GenerateSupplierToken(supplierID, tenantID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     uuid.Nil,
		TenantID:   tenantID,
		SupplierID: &supplierID,
		Role:       "supplier",
		TokenType:  "access",
		Version:    1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "puxbay",
			Subject:   supplierID.String(),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ---------------------------------------------------------
// Login & Registration
// ---------------------------------------------------------

// CurrentUser retrieves a user's profile details.
func (s *AuthService) CurrentUser(userID, tenantID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	result := s.db.
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Preload("User").
		Preload("Tenant").
		Preload("Role").
		First(&profile)

	if result.Error != nil {
		return nil, fmt.Errorf("profile not found: %w", result.Error)
	}

	return &profile, nil
}

// Authenticate verifies credentials and returns the user's profile and tokens.
func (s *AuthService) Authenticate(username, password, totpCode string) (*models.UserProfile, *TokenPair, error) {
	var user models.User
	if err := s.db.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("invalid credentials: %w", err)
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is inactive")
	}

	if !s.CheckPassword(password, user.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if user.RequirePasswordChange {
		return nil, nil, errors.New("password_change_required")
	}

	var profile models.UserProfile
	if user.IsSuperuser {
		// Mock profile for superadmin who doesn't have a specific tenant
		profile = models.UserProfile{
			UserID:   user.ID,
			TenantID: uuid.Nil,
			Role:     &models.Role{Name: "superadmin"},
			User:     user,
		}
	} else {
		if err := s.db.Where("user_id = ?", user.ID).Preload("User").Preload("Tenant").Preload("Role").First(&profile).Error; err != nil {
			return nil, nil, fmt.Errorf("no profile found for this user: %w", err)
		}
	}

	if profile.Is2FAEnabled {
		if totpCode == "" {
			return nil, nil, errors.New("2fa_required")
		}
		if profile.OTPSecret == nil || *profile.OTPSecret == "" {
			return nil, nil, errors.New("2fa is enabled but secret is missing")
		}
		if !s.validateTOTPCode(*profile.OTPSecret, totpCode) {
			return nil, nil, errors.New("invalid 2FA code")
		}
	}

	roleStr := ""
	if profile.Role != nil {
		roleStr = profile.Role.Name
	}
	tokens, err := s.GenerateTokenPair(user.ID, profile.TenantID, profile.BranchID, roleStr, &profile.RoleID, user.TokenVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Update LastLogin
	s.db.Model(&user).Update("last_login", time.Now())

	return &profile, tokens, nil
}

// ChangeTemporaryPassword allows a user to change their password on first login
func (s *AuthService) ChangeTemporaryPassword(username, temporaryPassword, newPassword string) (*models.UserProfile, *TokenPair, error) {
	var user models.User
	if err := s.db.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is inactive")
	}

	if !s.CheckPassword(temporaryPassword, user.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if !user.RequirePasswordChange {
		return nil, nil, fmt.Errorf("password change not required")
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return nil, nil, err
	}

	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"password":                hashedPassword,
		"require_password_change": false,
	}).Error; err != nil {
		return nil, nil, err
	}

	// Re-run authenticate now that the password is changed
	return s.Authenticate(username, newPassword, "")
}

// RegisterInput contains data for new tenant registration.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	FirstName   string
	LastName    string
	CompanyName string
	Subdomain   string
}

// Register creates a new tenant, user, profile, and default branch.
func (s *AuthService) Register(input RegisterInput) error {
	// Validate subdomain early — it will become the PostgreSQL schema name.
	// This prevents SQL injection via crafted subdomain values.
	if err := database.ValidateSchemaName(input.Subdomain); err != nil {
		return fmt.Errorf("invalid subdomain: %w", err)
	}

	var existingUser models.User
	if s.db.Where("username = ? OR email = ?", input.Username, input.Email).First(&existingUser).Error == nil {
		return errors.New("username or email already exists")
	}

	var existingTenant models.Tenant
	if s.db.Where("subdomain = ?", input.Subdomain).First(&existingTenant).Error == nil {
		return errors.New("subdomain is already taken")
	}

	hashedPassword, err := s.HashPassword(input.Password)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		tenant := models.Tenant{Name: input.CompanyName, Subdomain: input.Subdomain}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		user := models.User{
			Username:  input.Username,
			Email:     input.Email,
			Password:  hashedPassword,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			IsActive:  true,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		var adminRole models.Role
		if err := tx.Where("name = ? AND is_system = ?", "Admin", true).First(&adminRole).Error; err != nil {
			return fmt.Errorf("failed to find default admin role: %w", err)
		}

		profile := models.UserProfile{
			UserID:   user.ID,
			TenantID: tenant.ID,
			RoleID:   adminRole.ID,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		domain := models.Domain{TenantID: tenant.ID, Domain: input.Subdomain + "." + s.rootDomain, IsPrimary: true}
		if err := tx.Create(&domain).Error; err != nil {
			return err
		}

		metrics := models.TenantMetrics{TenantID: tenant.ID, TotalBranches: 0}
		if err := tx.Create(&metrics).Error; err != nil {
			return err
		}

		trialEnd := time.Now().AddDate(0, 0, 7)
		subscription := models.Subscription{
			TenantID:         tenant.ID,
			Status:           "trialing",
			CurrentPeriodEnd: &trialEnd,
		}
		if err := tx.Create(&subscription).Error; err != nil {
			return err
		}

		// 1. Create the PostgreSQL schema for the tenant (validated + quoted)
		if err := database.CreateTenantSchema(tx, tenant.SchemaName); err != nil {
			return err
		}

		// 2. Set search path for THIS transaction only to the new schema
		if err := database.SetTenantSchema(tx, tenant.SchemaName).Error; err != nil {
			return err
		}

		// 3. Migrate tenant models into this new schema
		if err := models.MigrateTenantModels(tx); err != nil {
			return err
		}

		// 4. Seed default Ledger Accounts for the tenant
		defaultAccounts := []models.LedgerAccount{
			{Name: "Cash on Hand", Type: "Asset", Code: "1000", Description: "Physical cash on hand"},
			{Name: "Checking Account", Type: "Asset", Code: "1010", Description: "Primary bank account"},
			{Name: "Accounts Receivable", Type: "Asset", Code: "1200", Description: "Money owed by customers"},
			{Name: "Inventory Asset", Type: "Asset", Code: "1300", Description: "Value of inventory on hand"},
			{Name: "Accounts Payable", Type: "Liability", Code: "2000", Description: "Money owed to suppliers"},
			{Name: "Sales Tax Payable", Type: "Liability", Code: "2200", Description: "Tax collected and payable"},
			{Name: "Owner's Equity", Type: "Equity", Code: "3000", Description: "Owner's investment"},
			{Name: "Retained Earnings", Type: "Equity", Code: "3900", Description: "Accumulated profits"},
			{Name: "Sales Revenue", Type: "Revenue", Code: "4000", Description: "Revenue from product sales"},
			{Name: "Service Revenue", Type: "Revenue", Code: "4100", Description: "Revenue from services"},
			{Name: "Cost of Goods Sold", Type: "Expense", Code: "5000", Description: "Direct cost of products sold"},
			{Name: "Rent Expense", Type: "Expense", Code: "6000", Description: "Facility rent"},
			{Name: "Payroll Expense", Type: "Expense", Code: "6100", Description: "Employee salaries and wages"},
			{Name: "Utilities Expense", Type: "Expense", Code: "6200", Description: "Electricity, water, internet"},
		}
		
		for _, acc := range defaultAccounts {
			if err := tx.Create(&acc).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
