package main

import (
	"discard/message-service/pkg/controllers"
	logger "discard/message-service/pkg/models/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ADDRESS                 string = "0.0.0.0"
	PORT                    string = "8080"
	RABBITMQ_SERVER_ADDRESS string = "amqp://guest:guest@rabbitmq:5672/"
)

func main() {
	connected := false
	var activeConnection *amqp.Connection = nil
	for !connected {
		conn, err := amqp.Dial(RABBITMQ_SERVER_ADDRESS)
		if err != nil {
			logger.WARN.Println("Failed to connect to RabbitMQ, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		} else {
			activeConnection = conn
			connected = true
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
	go func() {
		for d := range messages {
			logger.LOG.Printf("Received a message: %s\n", d.Body)
		}
	}()
	logger.LOG.Printf("Waiting for messages... To exit press CTRL+C")

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET(
		"/ping",
		controllers.Ping,
	)
	fullAddress := strings.Join([]string{ADDRESS, PORT}, ":")
	logger.LOG.Printf("Starting API server on %v...\n", fullAddress)
	logger.LOG.Printf("API server started on %v!\n", fullAddress)
	logger.FailOnError(router.Run(fullAddress), "Failed to run the server")
	<-forever
}
