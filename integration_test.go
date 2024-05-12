package main

import (
	"bytes"
	"context"
	"discard/message-service/pkg/models/logger"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func getBridgedTestNetwork(ctx context.Context) (*testcontainers.DockerNetwork, error) {
	return network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		network.WithDriver("bridge"),
		network.WithLabels(map[string]string{"name": "test-network"}),
	)
}

func getRabbitMQContainer(ctx context.Context, network *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	reqRabbit := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.13-management-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog(".*Server startup complete.*").AsRegexp().WithStartupTimeout(120 * time.Second),
		Name:         "rabbitmq_test",
		Hostname:     "rabbitmq",
		Networks:     []string{network.Name},
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: reqRabbit,
		Started:          false,
	})
}

func TestSimpleRabbitMQStart(t *testing.T) {
	net, err := getBridgedTestNetwork(context.Background())
	if err != nil {
		logger.FailOnError(err, "Failed to create network")
	}

	rabbitMQ, err := getRabbitMQContainer(context.Background(), net)
	if err != nil {
		logger.FailOnError(err, "Failed to create RabbitMQ container ")
	}

	err = rabbitMQ.Start(context.Background())
	if err != nil {
		logger.FailOnError(err, "Failed to start RabbitMQ container ")
	}

	defer func() {
		if err := rabbitMQ.Terminate(context.Background()); err != nil {
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

	// Create docker network
	net, err := getBridgedTestNetwork(ctx)
	logger.FailOnError(err, "Failed to create network")

	// Create and start RabbitMQ container
	rabbitMQ, err := getRabbitMQContainer(ctx, net)
	logger.FailOnError(err, "Failed to create RabbitMQ container")
	err = rabbitMQ.Start(ctx)
	logger.FailOnError(err, "Failed to start RabbitMQ container")

	amqpAddress := "amqp://guest:guest@rabbitmq:5672/"
	logger.LOG.Println("RabbitMQ URL: ", amqpAddress)

	// Create and start message-service container
	messageContainerRequest := testcontainers.ContainerRequest{
		Image:        "ghcr.io/syl-discard/message-service:main",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"RABBITMQ_SERVER_ADDRESS": amqpAddress,
			"DISCARD_STATE":           "INTEGRATION",
		},
		WaitingFor: wait.ForLog(".*Successfully connected to RabbitMQ!.*").AsRegexp(),
		Networks:   []string{net.Name},
	}
	messageServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: messageContainerRequest,
		Started:          true,
	})
	logger.FailOnError(err, "Failed to start message-service container")

	// messageServiceContainer.StartLogProducer(ctx)            // deprecated, TODO: remove
	// messageServiceContainer.FollowOutput(&TestLogConsumer{}) // deprecated, TODO: remove

	logger.LOG.Println("Message service container started")
	messageHost, err := messageServiceContainer.Host(ctx)
	logger.FailOnError(err, "Failed to get message-service container host")
	messagePort, err := messageServiceContainer.MappedPort(ctx, "8080")
	logger.FailOnError(err, "Failed to get message-service container port")
	logger.LOG.Println("Message service container port: ", messagePort)
	logger.LOG.Println("Message service container host: ", messageHost)

	// http request to message's ping
	messagePingResponse, err := http.Get("http://" + messageHost + ":" + messagePort.Port() + "/ping")
	logger.FailOnError(err, "Failed to ping message-service container")
	logger.LOG.Println("Message service container ping response: ", messagePingResponse)

	assert.Equal(t, 200, messagePingResponse.StatusCode)

	// Create and start user-service container
	userContainerRequest := testcontainers.ContainerRequest{
		Image:        "ghcr.io/syl-discard/user-service:main",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"RABBITMQ_SERVER_ADDRESS": amqpAddress,
		},
		WaitingFor: wait.ForLog(".*Successfully connected to RabbitMQ!.*").AsRegexp(),
		Networks:   []string{net.Name},
	}
	userServiceContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: userContainerRequest,
		Started:          true,
	})
	logger.FailOnError(err, "Failed to start user-service container")

	// userServiceContainer.StartLogProducer(ctx)            // deprecated, TODO: remove
	// userServiceContainer.FollowOutput(&TestLogConsumer{}) // deprecated, TODO: remove

	logger.LOG.Println("User service container started")
	userHost, err := userServiceContainer.Host(ctx)
	logger.FailOnError(err, "Failed to get message-service container host")
	userPort, err := userServiceContainer.MappedPort(ctx, "8080")
	logger.FailOnError(err, "Failed to get message-service container port")
	logger.LOG.Println("Message service container port: ", userPort)
	logger.LOG.Println("Message service container host: ", userHost)

	// http request to user's ping
	userPingResponse, err := http.Get("http://" + userHost + ":" + userPort.Port() + "/ping")
	logger.FailOnError(err, "Failed to ping user-service container")
	logger.LOG.Println("User service container ping response: ", userPingResponse)
	assert.Equal(t, userPingResponse.StatusCode, 200)

	assert.Equal(t, 200, userPingResponse.StatusCode)

	// send post request to user to delete
	var data = []byte(`{"id": "123e4567-e89b-12d3-a456-426614174000"}`)
	userDeleteRequest, err := http.NewRequest("POST", "http://"+userHost+":"+userPort.Port()+"/delete-user", bytes.NewBuffer(data))
	logger.FailOnError(err, "Failed to create delete user request")
	userDeleteRequest.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(userDeleteRequest)
	logger.FailOnError(err, "Failed to send delete user request")
	responseBody, _ := io.ReadAll(response.Body)
	logger.LOG.Println("User delete response: ", string(responseBody))
	wait.ForLog(".*Received a message: Deletion request.*").AsRegexp().WithPollInterval(25 * time.Millisecond)

	logs, err := messageServiceContainer.Logs(ctx)
	logger.FailOnError(err, "Failed to get message-service container logs")
	all, _ := io.ReadAll(logs)
	logger.LOG.Println("Message service container logs: ")
	logger.LOG.Println("====================BEGIN LOGS===============")
	logger.LOG.Println("\n", string(all))
	logger.LOG.Println("====================END LOGS=================")

	// Test for receiving deletion request
	lines := strings.Split(strings.TrimSpace(string(all)), "\n")
	lastLine := lines[len(lines)-1]
	assert.Contains(t, lastLine, "Received a message: Deletion request for user: 123e4567-e89b-12d3-a456-426614174000")

	defer func() {
		if err := rabbitMQ.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate RabbitMQ container")
		}
		if err := messageServiceContainer.Terminate(ctx); err != nil {
			logger.FailOnError(err, "Failed to terminate user-service container")
		}
		if err := net.Remove(ctx); err != nil {
			logger.FailOnError(err, "Failed to remove network")
		}
	}()
}

// Below is deprecated, TODO: remove
// type TestLogConsumer struct {
// }

// func (g *TestLogConsumer) Accept(l testcontainers.Log) {
// 	fmt.Fprintf(os.Stdout, "[CONTAINER LOG] %s %s\n", time.Now().Format("2006/01/02 15:04:05"), l.Content)
// }
