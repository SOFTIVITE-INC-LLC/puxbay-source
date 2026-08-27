package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type FinancialService struct {
	db *gorm.DB
}

func NewFinancialService(db *gorm.DB) *FinancialService {
	return &FinancialService{db: db}
}

func (s *FinancialService) ListExpenses(tenantID uuid.UUID, branchID string, startDate, endDate string) ([]models.Expense, error) {
	var expenses []models.Expense
	query := s.db.Preload("Category").Order("date desc")
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			query = query.Where("date >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			// Add 1 day to make it inclusive if it's just a date
			query = query.Where("date <= ?", t.Add(24*time.Hour))
		}
	}
	if err := query.Find(&expenses).Error; err != nil {
		return nil, err
	}
	return expenses, nil
}

type ExpenseCreateInput struct {
	CategoryID         string
	Amount             float64
	Date               string
	Description        string
	IsRecurring        bool
	RecurrenceInterval string
	ReceiptURL         string
	CreatedByID        uuid.UUID
	BranchID           *uuid.UUID
}

func parseExpenseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 first, then simple date format
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid date format, expected YYYY-MM-DD or RFC3339")
}

func (s *FinancialService) CreateExpense(input ExpenseCreateInput) (*models.Expense, error) {
	date, err := parseExpenseDate(input.Date)
	if err != nil {
		return nil, err
	}

	categoryUUID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	expense := models.Expense{
		CategoryID:         categoryUUID,
		Amount:             input.Amount,
		Date:               date,
		Description:        &input.Description,
		IsRecurring:        input.IsRecurring,
		RecurrenceInterval: input.RecurrenceInterval,
		CreatedByID:        &input.CreatedByID,
	}
	if input.ReceiptURL != "" {
		expense.ReceiptURL = &input.ReceiptURL
	}
	expense.BranchID = input.BranchID

	if err := s.db.Create(&expense).Error; err != nil {
		return nil, err
	}

	return &expense, nil
}

type ExpenseUpdateInput struct {
	CategoryID         string
	Amount             float64
	Date               string
	Description        string
	IsRecurring        *bool
	RecurrenceInterval *string
	ReceiptURL         *string
}

func (s *FinancialService) UpdateExpense(id uuid.UUID, input ExpenseUpdateInput) (*models.Expense, error) {
	var expense models.Expense
	if err := s.db.First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]any{
		"amount":      input.Amount,
		"description": input.Description,
	}

	if input.CategoryID != "" {
		catID, err := uuid.Parse(input.CategoryID)
		if err == nil {
			updates["category_id"] = catID
		}
	}

	if input.Date != "" {
		date, err := parseExpenseDate(input.Date)
		if err == nil {
			updates["date"] = date
		}
	}

	if input.IsRecurring != nil {
		updates["is_recurring"] = *input.IsRecurring
	}
	if input.RecurrenceInterval != nil {
		updates["recurrence_interval"] = *input.RecurrenceInterval
	}
	if input.ReceiptURL != nil {
		updates["receipt_url"] = *input.ReceiptURL
	}

	if err := s.db.Model(&expense).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &expense, nil
}

func (s *FinancialService) DeleteExpense(id uuid.UUID) error {
	return s.db.Delete(&models.Expense{}, "id = ?", id).Error
}

type MonthlyFinancialData struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Expense float64 `json:"expense"`
}

type ProfitAndLossData struct {
	GrossRevenue            float64                `json:"gross_revenue"`
	COGS                    float64                `json:"cogs"`
	GrossProfit             float64                `json:"gross_profit"`
	TotalExpenses           float64                `json:"total_expenses"`
	NetProfit               float64                `json:"net_profit"`
	TaxCollected            float64                `json:"tax_collected"`
	OperatingCashFlow       float64                `json:"operating_cash_flow"`
	CreditSales             float64                `json:"credit_sales"`
	CashSales               float64                `json:"cash_sales"`
	DebtCollections         float64                `json:"debt_collections"`
	TotalAccountsReceivable float64                `json:"total_accounts_receivable"`
	OverdueReceivables      float64                `json:"overdue_receivables"`
	MonthlyData             []MonthlyFinancialData `json:"monthly_data"`
}

func (s *FinancialService) GetProfitAndLoss(tenantID uuid.UUID, branchID string, startDate, endDate string) (*ProfitAndLossData, error) {
	type TotalResult struct {
		Total float64
	}

	db := s.db.Session(&gorm.Session{NewDB: true})

	var revenueData struct {
		TotalRevenue float64
		TotalTax     float64
		OrderCount   int64
	}
	revQ := db.Model(&models.Order{}).Where("status = ?", "completed")
	if branchID != "" {
		revQ = revQ.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			revQ = revQ.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			revQ = revQ.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}
	revQ.Select("COALESCE(SUM(total), 0) as total_revenue, COALESCE(SUM(tax), 0) as total_tax, COUNT(id) as order_count").
		Scan(&revenueData)

	// Credit Sales vs Cash Sales
	var creditSales float64
	creditQ := db.Model(&models.Order{}).Where("status = ? AND payment_method IN ('credit', 'bnpl', 'store_credit')", "completed")
	if branchID != "" {
		creditQ = creditQ.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			creditQ = creditQ.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			creditQ = creditQ.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}
	creditQ.Select("COALESCE(SUM(total), 0)").Scan(&creditSales)
	cashSales := revenueData.TotalRevenue - creditSales

	// Customer Debt Collections (repayments in period)
	var debtCollections float64
	repayQ := db.Model(&models.CreditTransaction{}).Where("transaction_type = 'repayment'")
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			repayQ = repayQ.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			repayQ = repayQ.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}
	repayQ.Select("COALESCE(SUM(abs(amount)), 0)").Scan(&debtCollections)

	// Total Accounts Receivable (current snapshot)
	var totalReceivables float64
	db.Model(&models.Customer{}).Select("COALESCE(SUM(debt_balance), 0)").Scan(&totalReceivables)

	// Overdue Debt
	var overdueReceivables float64
	db.Model(&models.BNPLInstalment{}).Where("status != 'paid' AND due_date < CURRENT_TIMESTAMP").
		Select("COALESCE(SUM(amount - amount_paid), 0)").Scan(&overdueReceivables)

	var cogs float64
	cogsQ := db.Session(&gorm.Session{NewDB: true}).Table("order_items").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status = ? AND orders.deleted_at IS NULL", "completed")
	if branchID != "" {
		cogsQ = cogsQ.Where("orders.branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			cogsQ = cogsQ.Where("orders.created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			cogsQ = cogsQ.Where("orders.created_at <= ?", t.Add(24*time.Hour))
		}
	}
	cogsQ.Select("COALESCE(SUM(order_items.quantity * order_items.cost_price), 0)").
		Scan(&cogs)

	var expenseResult TotalResult
	expQ := db.Session(&gorm.Session{NewDB: true}).Model(&models.Expense{})
	if branchID != "" {
		expQ = expQ.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			expQ = expQ.Where("date >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			expQ = expQ.Where("date <= ?", t.Add(24*time.Hour))
		}
	}
	expQ.Select("COALESCE(sum(amount), 0) as total").Scan(&expenseResult)

	grossRevenue := revenueData.TotalRevenue - revenueData.TotalTax
	grossProfit := grossRevenue - cogs
	netProfit := grossProfit - expenseResult.Total
	operatingCashFlow := (cashSales + debtCollections - revenueData.TotalTax) - expenseResult.Total

	// Monthly Data
	var monthlyRevenues []struct {
		Month int
		Total float64
	}
	revMonthlyQ := s.db.Session(&gorm.Session{NewDB: true}).Model(&models.Order{}).Where("status = ? AND created_at >= date_trunc('year', CURRENT_DATE)", "completed")
	if branchID != "" {
		revMonthlyQ = revMonthlyQ.Where("branch_id = ?", branchID)
	}
	revMonthlyQ.Select("EXTRACT(MONTH FROM created_at) as month, SUM(total - tax) as total").
		Group("EXTRACT(MONTH FROM created_at)").Scan(&monthlyRevenues)

	var monthlyExpenses []struct {
		Month int
		Total float64
	}
	expMonthlyQ := s.db.Session(&gorm.Session{NewDB: true}).Model(&models.Expense{}).Where("date >= date_trunc('year', CURRENT_DATE)")
	if branchID != "" {
		expMonthlyQ = expMonthlyQ.Where("branch_id = ?", branchID)
	}
	expMonthlyQ.Select("EXTRACT(MONTH FROM date) as month, SUM(amount) as total").
		Group("EXTRACT(MONTH FROM date)").Scan(&monthlyExpenses)

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	monthlyData := make([]MonthlyFinancialData, 12)
	for i := 0; i < 12; i++ {
		monthlyData[i].Month = months[i]
	}
	for _, r := range monthlyRevenues {
		if r.Month >= 1 && r.Month <= 12 {
			monthlyData[r.Month-1].Revenue = r.Total
		}
	}
	for _, e := range monthlyExpenses {
		if e.Month >= 1 && e.Month <= 12 {
			monthlyData[e.Month-1].Expense = e.Total
		}
	}

	return &ProfitAndLossData{
		GrossRevenue:            grossRevenue,
		COGS:                    cogs,
		GrossProfit:             grossProfit,
		TotalExpenses:           expenseResult.Total,
		NetProfit:               netProfit,
		TaxCollected:            revenueData.TotalTax,
		OperatingCashFlow:       operatingCashFlow,
		CreditSales:             creditSales,
		CashSales:               cashSales,
		DebtCollections:         debtCollections,
		TotalAccountsReceivable: totalReceivables,
		OverdueReceivables:      overdueReceivables,
		MonthlyData:             monthlyData,
	}, nil
}

func (s *FinancialService) GetTaxConfig() (*models.TaxConfiguration, error) {
	var taxConfig models.TaxConfiguration
	if err := s.db.First(&taxConfig).Error; err != nil {
		return nil, err
	}
	return &taxConfig, nil
}

type TaxSummaryData struct {
	TotalSales        float64 `json:"total_sales"`
	TotalTaxCollected float64 `json:"total_tax_collected"`
	TaxableAmount     float64 `json:"taxable_amount"`
	OrderCount        int64   `json:"order_count"`
}

func (s *FinancialService) GetTaxReport(branchID string, startDate, endDate string) (*TaxSummaryData, error) {
	var taxSummary TaxSummaryData
	query := s.db.Session(&gorm.Session{NewDB: true}).Model(&models.Order{}).Where("status = ?", "completed")
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		if t, err := parseExpenseDate(startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := parseExpenseDate(endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}
	query.Select("COALESCE(SUM(total), 0) as total_sales, COALESCE(SUM(tax), 0) as total_tax_collected, COALESCE(SUM(subtotal), 0) as taxable_amount, COUNT(id) as order_count").
		Scan(&taxSummary)
	return &taxSummary, nil
}

// ------------------------------------------------------------------------
// Ledger & Accounting
// ------------------------------------------------------------------------

func (s *FinancialService) ListLedgerAccounts() ([]models.LedgerAccount, error) {
	var accounts []models.LedgerAccount
	if err := s.db.Order("code asc").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

type CreateLedgerAccountInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // Asset, Liability, Equity, Revenue, Expense
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (s *FinancialService) CreateLedgerAccount(input CreateLedgerAccountInput) (*models.LedgerAccount, error) {
	acc := models.LedgerAccount{
		Name:         input.Name,
		Type:         input.Type,
		Code:         input.Code,
		Description:  input.Description,
	}
	if err := s.db.Create(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (s *FinancialService) ListJournalEntries() ([]models.JournalEntry, error) {
	var entries []models.JournalEntry
	if err := s.db.Preload("Lines").Preload("Lines.Account").Order("created_at desc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

type JournalEntryLineInput struct {
	AccountID uuid.UUID `json:"account_id"`
	Amount    float64   `json:"amount"`
	IsDebit   bool      `json:"is_debit"`
}

type CreateJournalEntryInput struct {
	ReferenceID   *uuid.UUID              `json:"reference_id"`
	ReferenceType string                  `json:"reference_type"`
	Description   string                  `json:"description"`
	Lines         []JournalEntryLineInput `json:"lines"`
}

func (s *FinancialService) CreateJournalEntry(input CreateJournalEntryInput) (*models.JournalEntry, error) {
	var totalDebits float64
	var totalCredits float64

	for _, line := range input.Lines {
		if line.IsDebit {
			totalDebits += line.Amount
		} else {
			totalCredits += line.Amount
		}
	}

	// Validate double-entry
	if totalDebits != totalCredits {
		return nil, errors.New("journal entry must balance (total debits must equal total credits)")
	}

	entry := models.JournalEntry{
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
		Description:   input.Description,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		for _, l := range input.Lines {
			line := models.LedgerLine{
				JournalEntryID: entry.ID,
				AccountID:      l.AccountID,
				Amount:         l.Amount,
				IsDebit:        l.IsDebit,
			}
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &entry, nil
}
