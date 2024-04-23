package main

import (
	"discard/message-service/pkg/api"
	"discard/message-service/pkg/configuration"
	logger "discard/message-service/pkg/models/logger"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ADDRESS                 string = "0.0.0.0"
	PORT                    string = "8080"
	RABBITMQ_SERVER_ADDRESS string = "amqp://guest:guest@rabbitmq:5672/"
	DATABASE_URL            string = "message-db"
	DATABASE_KEYSPACE       string = "messages"
)

func main() {
	// Start GIN API server + DB connection
	configuration := configuration.Configuration{
		APISettings: configuration.APISettings{
			Address: ADDRESS,
			Port:    PORT,
		},
		DatabaseSettings: configuration.DatabaseSettings{
			Url:      DATABASE_URL,
			Keyspace: DATABASE_KEYSPACE,
		},
	}

	go api.InitializeAPI(configuration)

	apiReady := false
	for !apiReady {
		request, err := http.NewRequest("GET", "http://"+ADDRESS+":"+PORT+"/ping", nil)
		if err != nil {
			logger.WARN.Println("API not ready, retrying in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}

		_, err = http.DefaultClient.Do(request)
		if err != nil {
			logger.WARN.Println("API not ready, retrying in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}

		apiReady = true
	}

	// Start RabbitMQ connection
	rabbitConnected := false
	var activeConnection *amqp.Connection = nil
	for !rabbitConnected {
		conn, err := amqp.Dial(RABBITMQ_SERVER_ADDRESS)
		if err != nil {
			logger.WARN.Println("Failed to connect to RabbitMQ, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		} else {
			activeConnection = conn
			rabbitConnected = true
		}
	}
	defer activeConnection.Close()
	logger.LOG.Println("Successfully connected to RabbitMQ!")

	ch, err := activeConnection.Channel()
	logger.FailOnError(err, "Failed to create a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"delete-user", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	logger.FailOnError(err, "Failed to declare a queue")

	messages, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	logger.FailOnError(err, "Failed to start consuming messages")

	forever := make(chan bool)

	// Keep consuming messages
	go func() {
		for d := range messages {
			logger.LOG.Printf("Received a message: %s\n", d.Body)
		}
	}()
	logger.LOG.Printf("Waiting for messages... To exit press CTRL+C")

	<-forever
}
