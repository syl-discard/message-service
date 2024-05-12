package main

import (
	"context"
	"discard/message-service/pkg/models/logger"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestSimpleRabbitMQStart(t *testing.T) {
	ctx := context.Background()
	rabbitMQ, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.13-management-alpine"),
		rabbitmq.WithAdminUsername("guest"),
		rabbitmq.WithAdminPassword("guest"),
	)

	if err != nil {
		logger.FailOnError(err, "Failed to start RabbitMQ container ")
	}

	defer func() {
		if err := rabbitMQ.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate RabbitMQ container")
		}
	}()
}

func TestMessageServiceStart(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/syl-discard/message-service:main",
		ExposedPorts: []string{"8080/tcp"},
	}
	messageServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		logger.FailOnError(err, "Failed to start message-service container")
	}

	defer func() {
		if err := messageServiceContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate message-service container")
		}
	}()
}

func TestUserServiceStart(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/syl-discard/user-service:main",
		ExposedPorts: []string{"8080/tcp"},
	}
	userServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		logger.FailOnError(err, "Failed to start user-service container")
	}

	defer func() {
		if err := userServiceContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate user-service container")
		}
	}()
}

func TestDeletionMessageFromUserToMessageService(t *testing.T) {
	ctx := context.Background()
	rabbitMQ, err := rabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3.13-management-alpine"),
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("root"),
	)
	logger.FailOnError(err, "Failed to start RabbitMQ container")
	mqport, _ := rabbitMQ.MappedPort(ctx, "5672")
	dsn := fmt.Sprintf("amqp://guest:guest@127.0.0.1:%s/", mqport.Port())
	logger.FailOnError(err, "Failed to get RabbitMQ connection URL")
	logger.LOG.Println("RabbitMQ connection URL: ", dsn)

	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/syl-discard/message-service:main",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"RABBITMQ_SERVER_ADDRESS": dsn,
			"DISCARD_STATE":           "INTEGRATION",
		},
		// WaitingFor: wait.ForLog(".*Successfully connected to RabbitMQ!.*").AsRegexp(),
	}
	messageServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	messageServiceContainer.StartLogProducer(ctx)
	messageServiceContainer.FollowOutput(&TestLogConsumer{})

	logger.FailOnError(err, "Failed to start message-service container")
	logger.LOG.Println("Message service container started")
	host, err := messageServiceContainer.Host(ctx)
	logger.FailOnError(err, "Failed to get message-service container host")
	port, err := messageServiceContainer.MappedPort(ctx, "8080")
	logger.FailOnError(err, "Failed to get message-service container port")
	logger.LOG.Println("Message service container port: ", port)
	logger.LOG.Println("Message service container host: ", host)

	apiReady := false
	for !apiReady {
		request, err := http.NewRequest("GET", "http://"+host+":"+port.Port()+"/ping", nil)
		if err != nil {
			// logger.WARN.Println("API not ready, retrying in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}

		_, err = http.DefaultClient.Do(request)
		if err != nil {
			// logger.WARN.Println("API not ready, retrying in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}

		apiReady = true
	}

	// http request to ping
	response, err := http.Get("http://" + host + ":8080/ping")
	logger.FailOnError(err, "Failed to ping message-service container")
	logger.LOG.Println("Message service container ping response: ", response)

	defer func() {
		if err := rabbitMQ.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate RabbitMQ container")
		}
		if err := messageServiceContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate user-service container")
		}
	}()
}

type TestLogConsumer struct {
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	fmt.Fprintf(os.Stdout, "[CONTAINER LOG] %s %s\n", time.Now().Format("2006/01/02 15:04:05"), l.Content)
}
