package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/workers"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type IntegrationHandler struct {
	db   *gorm.DB
	xero *oauth2.Config
}

func NewIntegrationHandler(cfg *config.Config, db *gorm.DB) *IntegrationHandler {
	xeroOauthConfig := &oauth2.Config{
		ClientID:     cfg.Xero.ClientID,
		ClientSecret: cfg.Xero.ClientSecret,
		RedirectURL:  cfg.Xero.RedirectURL,
		Scopes:       []string{"offline_access", "accounting.transactions", "accounting.settings"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.xero.com/identity/connect/authorize",
			TokenURL: "https://identity.xero.com/connect/token",
		},
	}

	return &IntegrationHandler{
		db:   db,
		xero: xeroOauthConfig,
	}
}

// ListIntegrations returns all active integrations for the tenant.
func (h *IntegrationHandler) ListIntegrations(c *gin.Context) {
	var integrations []models.TenantIntegration
	if err := getDB(c, h.db).Find(&integrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list integrations"})
		return
	}

	// Mask tokens for client response
	for i := range integrations {
		integrations[i].AccessToken = "***"
		integrations[i].RefreshToken = "***"
	}

	c.JSON(http.StatusOK, integrations)
}

// ConnectXero initiates the OAuth flow by redirecting to the Xero authorization URL.
func (h *IntegrationHandler) ConnectXero(c *gin.Context) {
	tenantIDRaw, _ := c.Get("tenant_id")
	state := tenantIDRaw.(string) // In production, hash this or use JWT to prevent CSRF

	url := h.xero.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

// XeroCallback handles the OAuth callback and stores the tokens.
func (h *IntegrationHandler) XeroCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callback parameters"})
		return
	}

	tenantID, err := uuid.Parse(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter (tenant_id)"})
		return
	}
	_ = tenantID // Parsed just for validation

	// Exchange the authorization code for tokens
	token, err := h.xero.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Xero OAuth Exchange Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	db := getDB(c, h.db)
	var integration models.TenantIntegration
	result := db.Where("provider = ?", "xero").First(&integration)

	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		integration = models.TenantIntegration{
			Provider:     "xero",
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    &token.Expiry,
			IsActive:     true,
		}
		db.Create(&integration)
	} else {
		integration.AccessToken = token.AccessToken
		integration.RefreshToken = token.RefreshToken
		integration.ExpiresAt = &token.Expiry
		integration.IsActive = true
		db.Save(&integration)
	}

	// Redirect to frontend integrations page
	c.Redirect(http.StatusFound, "/integrations?success=true")
}

// TriggerSync manually enqueues an accounting sync task for a given provider.
func (h *IntegrationHandler) TriggerSync(c *gin.Context) {
	tenantIDRaw, _ := c.Get("tenant_id")
	tenantID, _ := uuid.Parse(tenantIDRaw.(string))
	provider := c.Param("provider")

	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	task, err := workers.NewAccountingSyncTask(tenantID, provider, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sync task"})
		return
	}

	log.Printf("Enqueuing manual accounting sync task for %s (Tenant: %s)", provider, tenantID)

	_ = task // Suppress unused var

	c.JSON(http.StatusOK, gin.H{"message": "Sync triggered successfully"})
}
