// database/mongo.go
package database

import (
	"context"
	"log"
	"time"

	"webhook-listener-mekarisign/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func ConnectMongoDB(cfg *config.Config) (*Database, error) {
	clientOptions := options.Client().ApplyURI(cfg.DatabaseURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")

	return &Database{Client: client, DB: client.Database(cfg.DatabaseName)}, nil
}
