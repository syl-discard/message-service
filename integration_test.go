package main

import (
	"context"
	"discard/message-service/pkg/models/logger"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestCassandraServiceStart(t *testing.T) {
	ctx := context.Background()
	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:5.0"),
		cassandra.WithInitScripts("integration_cassandra.cql"),
		cassandra.WithConfigFile("integration_cassandra.yaml"),
	)
	if err != nil {
		logger.FailOnError(err, "Failed to start Cassandra container")
	}
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate Cassandra container")
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
		rabbitmq.WithAdminUsername("guest"),
		rabbitmq.WithAdminPassword("guest"),
	)
	logger.FailOnError(err, "Failed to start RabbitMQ container")

	rabbitConnectionURL, err := rabbitMQ.AmqpURL(ctx)
	logger.FailOnError(err, "Failed to get RabbitMQ connection URL")
	logger.LOG.Println("RabbitMQ connection URL: ", rabbitConnectionURL)

	messageServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/syl-discard/message-service:main",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"RABBITMQ_SERVER_ADDRESS": rabbitConnectionURL,
			},
			WaitingFor: wait.ForHTTP("/ping").WithPort("8080"),
		},
		Started: true,
	})
	logger.FailOnError(err, "Failed to start message-service container")
	logger.LOG.Println("Message service container started")

	defer func() {
		if err := rabbitMQ.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate RabbitMQ container")
		}
		if err := messageServiceContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate user-service container")
		}
	}()
}
