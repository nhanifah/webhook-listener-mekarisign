package service

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
)

type EmailSendStruct struct {
	To            string `json:"to"`
	Subject       string `json:"subject"`
	RecipientName string `json:"recipient_name"`
	Link          string `json:"link"`
	DueDate       string `json:"due_date"`
}

type PaymentNotificationStruct struct {
	To            string `json:"to"`
	Subject       string `json:"subject"`
	RecipientName string `json:"recipient_name"`
	InvoiceID     string `json:"invoice_id"`
	Amount        string `json:"amount"`
}

type EmailPayload struct {
	Type    string `json:"type"`
	Subject string `json:"subject"`
	To      string `json:"to"`
	Format  string `json:"format"`
	Msg     string `json:"msg"`
}

// LoadEmailTemplate memuat template HTML dan menggantikan placeholder dengan data
func LoadEmailTemplate(templateFile string, data EmailSendStruct) (string, error) {
	// Buka file template
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return "", fmt.Errorf("failed to load email template: %v", err)
	}

	// Render template dengan data
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to render email template: %v", err)
	}

	return rendered.String(), nil
}

func LoadEmailPaymentTemplate(templateFile string, data PaymentNotificationStruct) (string, error) {
	// Buka file template
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return "", fmt.Errorf("failed to load email template: %v", err)
	}

	// Render template dengan data
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to render email template: %v", err)
	}

	return rendered.String(), nil
}

// SendEmail contoh fungsi untuk mengirim email
func (r *RabbitMQService) SendEmail(to string, subject string, templateFile string, templateData EmailSendStruct) error {
	// Render template HTML
	emailBody, err := LoadEmailTemplate(templateFile, templateData)
	if err != nil {
		log.Printf("Error rendering email template: %v", err)
		return err
	}

	// payload
	payload := EmailPayload{
		Type:    "email",
		Subject: subject,
		To:      to,
		Format:  "html",
		Msg:     emailBody,
	}

	r.PublishJSON(payload)

	// Simulasi pengiriman email (gantilah dengan SMTP atau layanan email)
	// log.Printf("Sending email to: %s\nSubject: %s\nBody:\n%s", to, subject, emailBody)
	return nil
}

func (r *RabbitMQService) SendPaymentNotification(to string, subject string, templateFile string, templateData PaymentNotificationStruct) error {
	// Render template HTML
	emailBody, err := LoadEmailPaymentTemplate(templateFile, templateData)
	if err != nil {
		log.Printf("Error rendering email template: %v", err)
		return err
	}

	// payload
	payload := EmailPayload{
		Type:    "email",
		Subject: subject,
		To:      to,
		Format:  "html",
		Msg:     emailBody,
	}

	r.PublishJSON(payload)

	// Simulasi pengiriman email (gantilah dengan SMTP atau layanan email)
	// log.Printf("Sending email to: %s\nSubject: %s\nBody:\n%s", to, subject, emailBody)
	return nil
}
