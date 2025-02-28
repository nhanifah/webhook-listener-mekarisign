package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"webhook-listener-mekarisign/database"

	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	db *database.Database
}

type WebhookLog struct {
	Method    string              `json:"method"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	Timestamp string              `json:"timestamp"`
}

func NewWebhookHandler(db *database.Database) *WebhookHandler {
	return &WebhookHandler{db: db}
}

func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read request body"})
	}
	defer c.Request().Body.Close()

	webhookData := WebhookLog{
		Method:    c.Request().Method,
		Headers:   c.Request().Header,
		Body:      string(body),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection.InsertOne(context.Background(), webhookData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert data into DB"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Webhook received"})
}
