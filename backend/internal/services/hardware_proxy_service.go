package services

import (
	"log"

	"github.com/softivite/puxbay/internal/models"
)

type HardwareProxyService interface {
	PrintReceipt(order *models.Order) error
	OpenCashDrawer() error
}

type LocalHardwareProxyStub struct {
	Enabled bool
}

func NewLocalHardwareProxyStub() *LocalHardwareProxyStub {
	return &LocalHardwareProxyStub{Enabled: true}
}

func (h *LocalHardwareProxyStub) PrintReceipt(order *models.Order) error {
	if !h.Enabled {
		return nil
	}
	log.Printf("[HARDWARE-PROXY-STUB] Sending Print Job for Receipt #%s to Local Printer...", order.OrderNumber)
	return nil
}

func (h *LocalHardwareProxyStub) OpenCashDrawer() error {
	if !h.Enabled {
		return nil
	}
	log.Printf("[HARDWARE-PROXY-STUB] Sending pulse to open Cash Drawer...")
	return nil
}
