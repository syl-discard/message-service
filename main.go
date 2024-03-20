package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ADDRESS                 string = "0.0.0.0"
	PORT                    string = "8081"
	RABBITMQ_SERVER_ADDRESS string = "amqp://guest:guest@rabbitmq/"
)

var (
	WARN  = log.New(os.Stderr, "[WARNING]\t", log.LstdFlags|log.Lmsgprefix)
	ERROR = log.New(os.Stderr, "[ERROR]\t", log.LstdFlags|log.Lmsgprefix)
	LOG   = log.New(os.Stdout, "[INFO]\t", log.LstdFlags|log.Lmsgprefix)
)

func failOnError(err error, msg string) {
	if err != nil {
		ERROR.Printf("%s: %s\n", msg, err)
		panic(err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	LOG.Println("Successfully connected to RabbitMQ!")

	ch, err := conn.Channel()
	failOnError(err, "Failed to create a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	messages, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to start consuming messages")

	forever := make(chan bool)
	go func() {
		for d := range messages {
			LOG.Printf("Received a message: %s\n", d.Body)
		}
	}()
	LOG.Printf("Waiting for messages... To exit press CTRL+C")
	<-forever

	// fullAddress := strings.Join([]string{ADDRESS, PORT}, ":")
	// fmt.Print("Starting server on")
	// fmt.Println(fullAddress)

	// gin.SetMode(gin.ReleaseMode)
	// router := gin.Default()

	// if err := router.Run(fullAddress); err != nil {
	// 	log.Printf("failed to run the server: %v", err)
	// }
}
