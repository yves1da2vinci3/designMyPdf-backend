package amqp

import (
	"encoding/json"
	"fmt"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

const queueName = "pdf_jobs"

type Client struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

// NewClient connects to RabbitMQ and declares the durable pdf_jobs queue.
func NewClient(amqpURL string) (*Client, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("amqp channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("amqp queue declare: %w", err)
	}

	return &Client{conn: conn, ch: ch}, nil
}

type jobMessage struct {
	JobID string `json:"job_id"`
}

// Publish enqueues a job ID for the worker to process.
func (c *Client) Publish(jobID string) error {
	body, err := json.Marshal(jobMessage{JobID: jobID})
	if err != nil {
		return err
	}
	return c.ch.Publish(
		"",        // default exchange
		queueName, // routing key = queue name
		false,     // mandatory
		false,     // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}

// Consume returns a channel of deliveries from the pdf_jobs queue.
func (c *Client) Consume() (<-chan amqp091.Delivery, error) {
	return c.ch.Consume(
		queueName,
		"",    // consumer tag — server-generated
		false, // auto-ack: we ack manually after processing
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
}

// Close shuts down the channel and connection.
func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
