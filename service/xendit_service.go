package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"webhook-listener-mekarisign/logger"

	"webhook-listener-mekarisign/config"

	xendit "github.com/xendit/xendit-go/v6"
	invoice "github.com/xendit/xendit-go/v6/invoice"
)

type XenditService struct {
	client *xendit.APIClient
	apiKey string
}

type CustomerObject struct {
	// The unique identifier for the customer.
	Id string `json:"id,omitempty"`
	// The customer's phone number.
	PhoneNumber string `json:"phone_number,omitempty"`
	// The customer's given names or first names.
	GivenNames string `json:"given_names,omitempty"`
	// The customer's surname or last name.
	Surname string `json:"surname,omitempty"`
	// The customer's email address.
	Email string `json:"email,omitempty"`
	// The customer's mobile phone number.
	MobileNumber string `json:"mobile_number,omitempty"`
	// An additional identifier for the customer.
	CustomerId string `json:"customer_id,omitempty"`
}

func StringPtr(s string) *string {
	if s == "" {
		return nil // Menghindari pointer ke string kosong
	}
	return &s
}

func NewXenditService(cfg *config.Config) *XenditService {
	client := xendit.NewClient(cfg.XenditSecretKey)
	return &XenditService{
		client: client,
		apiKey: cfg.XenditSecretKey,
	}
}

func (xs *XenditService) CreateInvoice(externalID, payerEmail string, description string, customerData *CustomerObject, invDuration string, amount float64, forUserID string) (*invoice.Invoice, error) {
	logger := logger.NewLogger()
	defer logger.Sync()

	customer := &invoice.CustomerObject{
		PhoneNumber:  *invoice.NewNullableString(StringPtr(customerData.PhoneNumber)),
		Id:           *invoice.NewNullableString(StringPtr(customerData.Id)),
		GivenNames:   *invoice.NewNullableString(StringPtr(customerData.GivenNames)),
		Surname:      *invoice.NewNullableString(StringPtr(customerData.Surname)),
		Email:        *invoice.NewNullableString(StringPtr(customerData.Email)),
		MobileNumber: *invoice.NewNullableString(StringPtr(customerData.MobileNumber)),
		CustomerId:   *invoice.NewNullableString(StringPtr(customerData.CustomerId)),
		Addresses:    []invoice.AddressObject{}, // Sesuaikan jika ada alamat
	}

	currency := "IDR"
	sendEmail := true

	createInvoiceRequest := *invoice.NewCreateInvoiceRequest(externalID, amount)
	createInvoiceRequest.PayerEmail = &payerEmail
	createInvoiceRequest.Description = &description
	createInvoiceRequest.Currency = &currency
	createInvoiceRequest.ShouldSendEmail = &sendEmail
	createInvoiceRequest.Customer = customer
	createInvoiceRequest.InvoiceDuration = &invDuration

	resp, r, err := xs.client.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(createInvoiceRequest).
		ForUserId(forUserID). // [OPTIONAL] Business ID for sub-account merchants
		Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InvoiceApi.CreateInvoice`: %v\n", err.Error())

		b, _ := json.Marshal(err.FullError())
		fmt.Fprintf(os.Stderr, "Full Error Struct: %v\n", string(b))

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}

	return resp, nil
}
