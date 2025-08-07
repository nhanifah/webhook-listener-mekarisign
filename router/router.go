package router

import (
	"webhook-listener-mekarisign/config"
	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/handler"
	"webhook-listener-mekarisign/service"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, db *database.Database, cfg *config.Config, rabbitMQ *service.RabbitMQService) {
	// Webhook Handler
	h := handler.NewWebhookHandler(db, service.NewXenditService(cfg), cfg, rabbitMQ)
	e.POST("/webhook", h.HandleWebhook)
	e.GET("/webhook", h.HandleWebhook)

	// Invoice Handler
	xs := service.NewXenditService(cfg)
	invoiceHandler := handler.NewInvoiceHandler(xs)
	e.POST("/invoice", invoiceHandler.CreateInvoice)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "ok")
	})
}
