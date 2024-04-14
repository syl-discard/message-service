package api

import (
	"discard/message-service/pkg/configuration"
	"discard/message-service/pkg/controllers"
	"discard/message-service/pkg/database"
	"discard/message-service/pkg/models/logger"
	"discard/message-service/pkg/repository"

	"github.com/gin-gonic/gin"
)

func InitializeAPI(configuration configuration.Configuration) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	databaseSession := database.ConnectToDatabase(
		configuration.DatabaseSettings.Url,
		configuration.DatabaseSettings.Keyspace,
	)

	defer databaseSession.Close()

	messageRepository := repository.NewMessageRepository(databaseSession)
	messageHandler := controllers.NewMessageHandler(&messageRepository)

	// endpoints
	router.GET("/ping", controllers.Ping)
	router.POST("/v1/message", messageHandler.SaveMessage)
	router.GET("/v1/message/:id", messageHandler.GetMessageById)

	fullAddress :=
		configuration.APISettings.Address + ":" + configuration.APISettings.Port

	logger.LOG.Printf("Starting API server on %v...\n", fullAddress)
	logger.LOG.Printf("API server started on %v!\n", fullAddress)
	logger.FailOnError(router.Run(fullAddress), "Failed to run the server")
}
