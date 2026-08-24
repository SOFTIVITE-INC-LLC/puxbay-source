package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/softivite/puxbay/internal/config"
)

type SMSService struct {
	config config.SMSConfig
}

func NewSMSService(cfg config.SMSConfig) *SMSService {
	return &SMSService{config: cfg}
}

type ArkeselPayload struct {
	Sender     string   `json:"sender"`
	Message    string   `json:"message"`
	Recipients []string `json:"recipients"`
}

// SendSMS sends an SMS message via Arkesel V2 API.
func (s *SMSService) SendSMS(recipients []string, message string) error {
	if s.config.APIKey == "" {
		log.Println("Warning: SMS API Key is not configured. Skipping SMS.")
		return nil
	}

	payload := ArkeselPayload{
		Sender:     s.config.SenderID,
		Message:    message,
		Recipients: recipients,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://sms.arkesel.com/api/v2/sms/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("arkesel api error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
