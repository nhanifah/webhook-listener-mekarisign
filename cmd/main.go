package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"webhook-listener-mekarisign/config"
	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/logger"
	"webhook-listener-mekarisign/router"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system ENV...")
	}

	// Load config
	cfg := config.LoadConfig()

	// Setup logger
	logger := logger.NewLogger()
	defer logger.Sync()

	// Connect to MongoDB
	db, err := database.ConnectMongoDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Client.Disconnect(context.Background())

	// Setup Echo server
	e := echo.New()
	router.SetupRoutes(e, db)

	// Start server
	logger.Info("Server running", zap.String("port", cfg.ServerPort))
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
