package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"webhook-listener-mekarisign/config"
)

// WhatsAppPayload struct untuk request body
type WhatsAppPayload struct {
	Sender          string   `json:"sender"`
	Recipient       string   `json:"recipient"`
	RecipientName   string   `json:"recipient_name"`
	TemplateID      string   `json:"template_id"`
	Attrb           []string `json:"attrb"`
	EnabledSchedule int      `json:"enabled_schedule"`
}

// SendWhatsAppMessage mengirim pesan ke API WhatsApp
func SendWhatsAppMessage(cfg *config.Config, payload WhatsAppPayload) error {
	url := "https://wapi.wkwk-japanese.com/api/queue"

	// Konversi payload ke JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Buat request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set header
	req.Header.Set("X-WEBHOOK-TOKEN", cfg.WhatsappToken)
	req.Header.Set("Content-Type", "application/json")

	// Kirim request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Periksa status response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}
