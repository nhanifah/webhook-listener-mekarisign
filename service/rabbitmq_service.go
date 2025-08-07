package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQService struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   amqp091.Queue
}

// NewRabbitMQService membuat koneksi ke RabbitMQ
func NewRabbitMQService(url string, queueName string) (*RabbitMQService, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		queueName, // Nama queue
		true,      // Durable
		false,     // Auto-delete
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQService{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

// Publish mengirim pesan ke queue
func (r *RabbitMQService) Publish(body string) error {
	ctx := context.Background()
	err := r.channel.PublishWithContext(
		ctx,
		"",           // Exchange
		r.queue.Name, // Routing key (queue name)
		false,        // Mandatory
		false,        // Immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}
	log.Printf(" [x] Sent %s", body)
	return nil
}

// PublishJSON mengirim data JSON ke queue
func (r *RabbitMQService) PublishJSON(payload EmailPayload) error {
	ctx := context.Background()

	// Encode ke JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",           // Exchange
		r.queue.Name, // Routing key (queue name)
		false,        // Mandatory
		false,        // Immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf(" [x] Sent JSON: %s", body)
	return nil
}

// Consume menerima pesan dari queue dengan manual ack
func (r *RabbitMQService) Consume() (<-chan amqp091.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name, // Queue
		"",           // Consumer
		false,        // Auto-ack diubah menjadi false
		false,        // Exclusive
		false,        // No-local
		false,        // No-wait
		nil,          // Args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}
	return msgs, nil
}

// ProcessMessages memproses pesan dari queue dengan rollback jika gagal
func (r *RabbitMQService) ProcessMessages(msgs <-chan amqp091.Delivery) {
	for msg := range msgs {
		log.Printf("Received message: %s", msg.Body)

		// Simulasi error handling
		if string(msg.Body) == "error" {
			log.Println("Processing failed, rolling back...")
			msg.Nack(false, true) // Kirim ulang pesan ke queue
			continue
		}

		// Jika berhasil, ack pesan
		msg.Ack(false)
		log.Println("Message processed successfully")
	}
}

// Close menutup koneksi RabbitMQ
func (r *RabbitMQService) Close() {
	r.channel.Close()
	r.conn.Close()
}
