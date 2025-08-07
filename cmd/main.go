package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"webhook-listener-mekarisign/config"
	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/logger"
	"webhook-listener-mekarisign/router"
	"webhook-listener-mekarisign/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system ENV...")
	}
	log.Println("App ver 1.0-rc")

	// Load config
	cfg := config.LoadConfig()

	// Setup logger
	logger := logger.NewLogger()
	defer logger.Sync()

	// Print config
	logger.Info("Config", zap.Any("config", cfg))

	// Inisialisasi RabbitMQ
	rmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMqUser, cfg.RabbitMqPassword, cfg.RabbitMqHost, cfg.RabbitMqPort)
	rabbitMQ, err := service.NewRabbitMQService(rmqUrl, cfg.RabbitMqQueue)
	if err != nil {
		log.Fatalf("Could not initialize RabbitMQ: %v", err)
	} else {
		log.Println("Connected to RabbitMQ")
	}
	defer rabbitMQ.Close()

	// Connect to MongoDB
	db, err := database.ConnectMongoDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Client.Disconnect(context.Background())

	// Connect to MySQL
	database.ConnectMySQL(cfg)
	defer database.DB.Close()

	// Setup Echo server
	e := echo.New()
	router.SetupRoutes(e, db, cfg, rabbitMQ)

	// Start server
	logger.Info("Server running", zap.String("port", cfg.ServerPort))
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
