package router

import (
	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/handler"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, db *database.Database) {
	h := handler.NewWebhookHandler(db)
	e.POST("/webhook", h.HandleWebhook)
	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "ok")
	})

}
