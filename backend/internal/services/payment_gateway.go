package services

import (
	"context"
	"fmt"
)

// PaymentGateway defines the interface for processing payments
type PaymentGateway interface {
	Charge(ctx context.Context, amount float64, currency, sourceToken string) (string, error)
	Refund(ctx context.Context, chargeID string, amount float64) error
}

// MockPaymentGateway is a stub for testing payment flows without hitting real APIs.
type MockPaymentGateway struct {
	ShouldFail bool
}

func NewMockPaymentGateway(shouldFail bool) *MockPaymentGateway {
	return &MockPaymentGateway{ShouldFail: shouldFail}
}

func (m *MockPaymentGateway) Charge(ctx context.Context, amount float64, currency, sourceToken string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock payment gateway declined the charge")
	}
	return "ch_mock_1234567890", nil
}

func (m *MockPaymentGateway) Refund(ctx context.Context, chargeID string, amount float64) error {
	if m.ShouldFail {
		return fmt.Errorf("mock payment gateway declined the refund")
	}
	return nil
}
