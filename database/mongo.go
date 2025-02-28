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
	Client     *mongo.Client
	Collection *mongo.Collection
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

	// Ambil koleksi yang dibutuhkan
	collection := client.Database(cfg.DatabaseName).Collection(cfg.Collection)

	return &Database{Client: client, Collection: collection}, nil
}
