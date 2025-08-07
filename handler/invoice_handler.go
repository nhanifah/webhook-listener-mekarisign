package handler

import (
	"net/http"

	"webhook-listener-mekarisign/service"

	"github.com/labstack/echo/v4"
)

type InvoiceHandler struct {
	xs *service.XenditService
}

func NewInvoiceHandler(xs *service.XenditService) *InvoiceHandler {
	return &InvoiceHandler{xs: xs}
}

func (h *InvoiceHandler) CreateInvoice(c echo.Context) error {
	type Request struct {
		ExternalID  string  `json:"external_id"`
		PayerEmail  string  `json:"payer_email"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		ForUserID   string  `json:"for_user_id"` // Optional
	}

	req := new(Request)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Create customer object
	customer := &service.CustomerObject{
		Id:           "CUST-789",
		PhoneNumber:  "+628111222333",
		GivenNames:   "Budi",
		Surname:      "Santoso",
		Email:        "budi@example.com",
		MobileNumber: "+628111222333",
		CustomerId:   "CUST-789",
	}

	InvoiceDuration := "432000" // 5 days

	invoice, err := h.xs.CreateInvoice(req.ExternalID, req.PayerEmail, req.Description, customer, InvoiceDuration, req.Amount, req.ForUserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, invoice)
}
