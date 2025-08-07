package config

import (
	"os"
)

type Config struct {
	DatabaseURL                string
	DatabaseName               string
	Collection                 string
	ServerPort                 string
	XenditSecretKey            string
	XenditPublicKey            string
	XenditCallbackURL          string
	XenditInvDuration          string
	XenditCallbackToken        string
	DirectorName               string
	DirectorMail               string
	RabbitMqHost               string
	RabbitMqPort               string
	RabbitMqUser               string
	RabbitMqPassword           string
	RabbitMqQueue              string
	WhatsappToken              string
	WhatsappNotificationNumber string
	DatabaseMysqlHost          string
	DatabaseMysqlPort          string
	DatabaseMysqlUser          string
	DatabaseMysqlPassword      string
	DatabaseMysqlDatabase      string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		DatabaseName:               os.Getenv("DB_NAME"),
		Collection:                 os.Getenv("COLLECTION_NAME"),
		ServerPort:                 os.Getenv("SERVER_PORT"),
		XenditSecretKey:            os.Getenv("XENDIT_SECRET_KEY"),
		XenditPublicKey:            os.Getenv("XENDIT_PUBLIC_KEY"),
		XenditCallbackURL:          os.Getenv("XENDIT_CALLBACK_URL"),
		XenditInvDuration:          os.Getenv("INVOICE_DURATION"),
		XenditCallbackToken:        os.Getenv("XENDIT_CALLBACK_TOKEN"),
		DirectorName:               os.Getenv("DIRECTOR_NAME"),
		DirectorMail:               os.Getenv("DIRECTOR_EMAIL"),
		RabbitMqHost:               os.Getenv("RABBITMQ_HOST"),
		RabbitMqPort:               os.Getenv("RABBITMQ_PORT"),
		RabbitMqUser:               os.Getenv("RABBITMQ_USER"),
		RabbitMqPassword:           os.Getenv("RABBITMQ_PASSWORD"),
		RabbitMqQueue:              os.Getenv("RABBITMQ_QUEUE_NAME"),
		WhatsappToken:              os.Getenv("WHATSAPP_TOKEN"),
		WhatsappNotificationNumber: os.Getenv("WHATSAPP_NOTIFICATION_NUMBER"),
		DatabaseMysqlHost:          os.Getenv("DB_MySQL_HOST"),
		DatabaseMysqlPort:          os.Getenv("DB_MySQL_PORT"),
		DatabaseMysqlUser:          os.Getenv("DB_MySQL_USER"),
		DatabaseMysqlPassword:      os.Getenv("DB_MySQL_PASSWORD"),
		DatabaseMysqlDatabase:      os.Getenv("DB_MySQL_DATABASE"),
	}
}
