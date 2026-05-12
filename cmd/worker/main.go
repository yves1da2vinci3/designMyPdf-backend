package main

import (
	"designmypdf/config/database"
	_ "designmypdf/config/env"
	"designmypdf/pkg/amqp"
	"designmypdf/pkg/pdfjob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	database.Initialize()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Fatal("RABBITMQ_URL not set")
	}

	amqpClient, err := amqp.NewClient(rabbitmqURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer amqpClient.Close()

	jobSvc := pdfjob.NewService(amqpClient)

	// Warm up the browser pool so the first job doesn't pay Chrome start cost.
	_ = pdfjob.GetBrowserPool()
	defer pdfjob.GetBrowserPool().Close()

	deliveries, err := amqpClient.Consume()
	if err != nil {
		log.Fatalf("failed to start consumer: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker started. Waiting for jobs...")

	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				log.Println("Delivery channel closed, shutting down")
				return
			}

			var msg struct {
				JobID string `json:"job_id"`
			}
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				fmt.Printf("worker: invalid message body: %v\n", err)
				d.Nack(false, false) // discard malformed message
				continue
			}

			if err := jobSvc.ProcessJob(msg.JobID); err != nil {
				fmt.Printf("worker: job %s failed: %v\n", msg.JobID, err)
				// Job is already marked failed in DB by ProcessJob; ack to
				// remove from queue (no requeue — failure is recorded in DB).
			}
			d.Ack(false)

		case <-quit:
			log.Println("Shutdown signal received")
			return
		}
	}
}
