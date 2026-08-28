package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type CreditService struct {
	db  *gorm.DB
	sms *SMSService
}

func NewCreditService(db *gorm.DB, sms *SMSService) *CreditService {
	return &CreditService{db: db, sms: sms}
}

func (s *CreditService) GetSMSService() *SMSService {
	return s.sms
}

// GetOrCreateCreditAccount returns the customer's credit account, creating one if none exists.
func (s *CreditService) GetOrCreateCreditAccount(tenantID, customerID uuid.UUID) (*models.CreditAccount, error) {
	var acc models.CreditAccount
	err := s.db.Where("customer_id = ?", customerID).First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			acc = models.CreditAccount{
				CustomerID:  customerID,
				CreditLimit: 0,
				Balance:     0,
				Status:      "active",
				DaysToRepay: 30,
			}
			if createErr := s.db.Create(&acc).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			return nil, err
		}
	}
	return &acc, nil
}

// SetCreditLimit updates the credit limit, repayment term, and notes for a customer.
func (s *CreditService) SetCreditLimit(tenantID, customerID uuid.UUID, limit float64, daysToRepay int, notes string) (*models.CreditAccount, error) {
	acc, err := s.GetOrCreateCreditAccount(tenantID, customerID)
	if err != nil {
		return nil, err
	}

	acc.CreditLimit = limit
	if daysToRepay > 0 {
		acc.DaysToRepay = daysToRepay
	}
	if notes != "" {
		acc.Notes = notes
	}

	if err := s.db.Save(acc).Error; err != nil {
		return nil, err
	}

	// Sync with Customer table in tenant schema
	s.db.Model(&models.Customer{}).Where("id = ?", customerID).
		Updates(map[string]interface{}{
			"credit_limit": limit,
		})

	return acc, nil
}

// DrawdownCredit charges an order or amount to the customer's store credit account.
// If instalmentsCount > 1, it automatically creates scheduled BNPL instalments.
func (s *CreditService) DrawdownCredit(tenantID, customerID uuid.UUID, orderID *uuid.UUID, amount float64, instalmentsCount int, createdByID *uuid.UUID, notes string) (*models.CreditTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("drawdown amount must be greater than zero")
	}

	var txRecord *models.CreditTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var acc models.CreditAccount
		if err := tx.Where("customer_id = ?", customerID).First(&acc).Error; err != nil {
			return errors.New("credit account not found for customer")
		}

		if acc.Status != "active" {
			return fmt.Errorf("credit account is currently %s", acc.Status)
		}

		availableCredit := acc.CreditLimit - acc.Balance
		if acc.CreditLimit > 0 && amount > availableCredit {
			return fmt.Errorf("credit limit exceeded. Available: %.2f, Requested: %.2f", availableCredit, amount)
		}

		newBalance := acc.Balance + amount
		now := time.Now()
		days := acc.DaysToRepay
		if days <= 0 {
			days = 30
		}
		dueDate := now.AddDate(0, 0, days)

		ref := fmt.Sprintf("BNPL-%s-%d", customerID.String()[:6], now.Unix())

		creditTx := models.CreditTransaction{
			CreditAccountID: acc.ID,
			CustomerID:      customerID,
			OrderID:         orderID,
			Amount:          amount,
			BalanceAfter:    newBalance,
			TransactionType: "drawdown",
			Reference:       ref,
			DueDate:         &dueDate,
			Status:          "pending",
			Notes:           notes,
			CreatedByID:     createdByID,
		}

		if err := tx.Create(&creditTx).Error; err != nil {
			return err
		}

		// Update Credit Account balance
		acc.Balance = newBalance
		acc.LastDrawdownAt = &now
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}

		// Update Customer debt_balance
		if err := tx.Model(&models.Customer{}).Where("id = ?", customerID).
			Update("debt_balance", newBalance).Error; err != nil {
			return err
		}

		// Generate BNPL instalments if requested
		if instalmentsCount > 1 {
			instalmentAmount := amount / float64(instalmentsCount)
			daysInterval := days / instalmentsCount
			if daysInterval < 7 {
				daysInterval = 7 // weekly minimum
			}

			for i := 1; i <= instalmentsCount; i++ {
				instDueDate := now.AddDate(0, 0, daysInterval*i)
				inst := models.BNPLInstalment{
					CreditTransactionID: creditTx.ID,
					CustomerID:          customerID,
					OrderID:             orderID,
					InstalmentNumber:    i,
					TotalInstalments:    instalmentsCount,
					Amount:              instalmentAmount,
					DueDate:             instDueDate,
					Status:              "pending",
				}
				if err := tx.Create(&inst).Error; err != nil {
					return err
				}
			}
		}

		txRecord = &creditTx
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send SMS confirmation in background if SMS is enabled and customer has phone
	if s.sms != nil {
		go func() {
			var cust models.Customer
			if err := s.db.Where("id = ?", customerID).First(&cust).Error; err == nil && cust.Phone != nil && *cust.Phone != "" {
				msg := fmt.Sprintf("Dear %s, your purchase of GHS %.2f on Store Credit/BNPL has been recorded. Current balance: GHS %.2f. Thank you!", cust.Name, amount, txRecord.BalanceAfter)
				_ = s.sms.SendTenantSMS(s.db, []string{*cust.Phone}, msg, "Store Credit / BNPL Drawdown SMS")
			}
		}()
	}

	return txRecord, nil
}

// RecordRepayment logs a cash, MoMo, or bank repayment and reduces the customer's credit balance.
func (s *CreditService) RecordRepayment(tenantID, customerID uuid.UUID, amount float64, paymentMethod, reference string, createdByID *uuid.UUID, notes string) (*models.CreditTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("repayment amount must be greater than zero")
	}

	var txRecord *models.CreditTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var acc models.CreditAccount
		if err := tx.Where("customer_id = ?", customerID).First(&acc).Error; err != nil {
			return errors.New("credit account not found")
		}

		if amount > acc.Balance {
			return fmt.Errorf("repayment amount (%.2f) exceeds current balance (%.2f)", amount, acc.Balance)
		}

		newBalance := acc.Balance - amount
		now := time.Now()

		ref := reference
		if ref == "" {
			ref = fmt.Sprintf("REP-%d", now.Unix())
		}

		creditTx := models.CreditTransaction{
			CreditAccountID: acc.ID,
			CustomerID:      customerID,
			Amount:          amount,
			BalanceAfter:    newBalance,
			TransactionType: "repayment",
			PaymentMethod:   paymentMethod,
			Reference:       ref,
			Notes:           notes,
			CreatedByID:     createdByID,
		}

		if err := tx.Create(&creditTx).Error; err != nil {
			return err
		}

		acc.Balance = newBalance
		acc.LastRepaymentAt = &now
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}

		// Update Customer debt_balance
		if err := tx.Model(&models.Customer{}).Where("id = ?", customerID).
			Update("debt_balance", newBalance).Error; err != nil {
			return err
		}

		// Automatically settle earliest pending BNPL instalments
		var pendingInstalments []models.BNPLInstalment
		tx.Where("customer_id = ? AND status = ?", customerID, "pending").
			Order("due_date asc").Find(&pendingInstalments)

		remainingRepayment := amount
		for _, inst := range pendingInstalments {
			if remainingRepayment <= 0 {
				break
			}
			if remainingRepayment >= inst.Amount {
				inst.Status = "paid"
				inst.PaidAt = &now
				tx.Save(&inst)
				remainingRepayment -= inst.Amount
			} else {
				// Partial instalment reduction
				inst.Amount -= remainingRepayment
				tx.Save(&inst)
				remainingRepayment = 0
			}
		}

		txRecord = &creditTx
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send SMS receipt in background
	if s.sms != nil {
		go func() {
			var cust models.Customer
			if err := s.db.Where("id = ?", customerID).First(&cust).Error; err == nil && cust.Phone != nil && *cust.Phone != "" {
				msg := fmt.Sprintf("Dear %s, your repayment of GHS %.2f via %s received. Remaining credit balance: GHS %.2f. Ref: %s.", cust.Name, amount, paymentMethod, txRecord.BalanceAfter, txRecord.Reference)
				_ = s.sms.SendTenantSMS(s.db, []string{*cust.Phone}, msg, "Credit Repayment Receipt SMS")
			}
		}()
	}

	return txRecord, nil
}

// GetCreditAccountDetails fetches the credit account, recent transactions, and scheduled BNPL instalments.
func (s *CreditService) GetCreditAccountDetails(tenantID, customerID uuid.UUID) (map[string]interface{}, error) {
	acc, err := s.GetOrCreateCreditAccount(tenantID, customerID)
	if err != nil {
		return nil, err
	}

	var transactions []models.CreditTransaction
	s.db.Where("customer_id = ?", customerID).
		Order("created_at desc").Limit(50).Find(&transactions)

	var instalments []models.BNPLInstalment
	s.db.Where("customer_id = ?", customerID).
		Order("due_date asc").Find(&instalments)

	available := acc.CreditLimit - acc.Balance
	if available < 0 {
		available = 0
	}

	return map[string]interface{}{
		"account":          acc,
		"available_credit": available,
		"transactions":     transactions,
		"instalments":      instalments,
	}, nil
}

// GetOverdueAccounts returns all customer credit accounts with past-due instalments or balances.
func (s *CreditService) GetOverdueAccounts(tenantID uuid.UUID) ([]map[string]interface{}, error) {
	var accounts []models.CreditAccount
	if err := s.db.Preload("Customer").Where("balance > 0").Find(&accounts).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	var results []map[string]interface{}

	for _, acc := range accounts {
		var overdueInstalments []models.BNPLInstalment
		s.db.Where("customer_id = ? AND status != 'paid' AND due_date < ?", acc.CustomerID, now).
			Find(&overdueInstalments)

		var overdueAmount float64
		for _, inst := range overdueInstalments {
			overdueAmount += (inst.Amount - inst.AmountPaid)
		}

		if overdueAmount > 0 || (acc.LastDrawdownAt != nil && now.Sub(*acc.LastDrawdownAt).Hours() > float64(acc.DaysToRepay*24)) {
			results = append(results, map[string]interface{}{
				"account":        acc,
				"customer":       acc.Customer,
				"total_balance":  acc.Balance,
				"overdue_amount": overdueAmount,
				"overdue_count":  len(overdueInstalments),
				"days_overdue":   int(now.Sub(*acc.LastDrawdownAt).Hours() / 24),
			})
		}
	}

	return results, nil
}
