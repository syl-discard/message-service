package api

import (
	"discard/message-service/pkg/configuration"
	"discard/message-service/pkg/controllers"
	"discard/message-service/pkg/database"
	"discard/message-service/pkg/models/logger"
	"discard/message-service/pkg/repository"
	"os"

	"github.com/gin-gonic/gin"
)

func InitializeAPI(configuration configuration.Configuration) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	var messageRepository repository.MessageRepository
	var messageHandler controllers.MessageHandler

	if os.Getenv("DISCARD_STATE") != "INTEGRATION" {
		databaseSession := database.ConnectToDatabase(
			configuration,
		)

		defer databaseSession.Close()

		messageRepository = repository.NewMessageRepository(databaseSession)
		messageHandler = controllers.NewMessageHandler(&messageRepository)
	} else {
		messageRepository = repository.NewInMemoryMessageRepository()
		messageHandler = controllers.NewMessageHandler(&messageRepository)
	}

	// endpoints
	router.GET("/api/v1/message/ping", controllers.Ping)
	router.POST("/api/v1/message", messageHandler.SaveMessage)
	router.GET("/api/v1/message/:id", messageHandler.GetMessageById)
	router.GET("/api/v1/message/user/:id", messageHandler.GetMessagesByUserId)
	router.DELETE("/api/v1/message/user/:id", messageHandler.DeleteMessagesByUserId)

	fullAddress :=
		configuration.APISettings.Address + ":" + configuration.APISettings.Port

	logger.LOG.Printf("Starting API server on %v...\n", fullAddress)
	logger.LOG.Printf("API server started on %v!\n", fullAddress)
	logger.FailOnError(router.Run(fullAddress), "Failed to run the server")
}
