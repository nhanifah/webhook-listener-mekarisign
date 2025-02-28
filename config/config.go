package config

import (
	"os"
)

type Config struct {
	DatabaseURL  string
	DatabaseName string
	Collection   string
	ServerPort   string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		DatabaseName: os.Getenv("DB_NAME"),
		Collection:   os.Getenv("COLLECTION_NAME"),
		ServerPort:   os.Getenv("SERVER_PORT"),
	}
}
