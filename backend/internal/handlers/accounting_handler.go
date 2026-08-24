package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type AccountingHandler struct {
	db *gorm.DB
}

func NewAccountingHandler(db *gorm.DB) *AccountingHandler {
	return &AccountingHandler{db: db}
}

func (h *AccountingHandler) service(c *gin.Context) *services.FinancialService {
	return services.NewFinancialService(getDB(c, h.db))
}

func (h *AccountingHandler) ListLedgerAccounts(c *gin.Context) {
	accounts, err := h.service(c).ListLedgerAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ledger accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

func (h *AccountingHandler) CreateLedgerAccount(c *gin.Context) {
	var input services.CreateLedgerAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.service(c).CreateLedgerAccount(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ledger account", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *AccountingHandler) ListJournalEntries(c *gin.Context) {
	entries, err := h.service(c).ListJournalEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch journal entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *AccountingHandler) CreateJournalEntry(c *gin.Context) {
	var input services.CreateJournalEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service(c).CreateJournalEntry(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}
