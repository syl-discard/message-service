package controllers

import (
	"discard/message-service/pkg/models"
	"discard/message-service/pkg/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type MessageHandler interface {
	SaveMessage(*gin.Context)
	GetMessageById(*gin.Context)
	GetMessagesByUserId(*gin.Context)
	DeleteMessagesByUserId(*gin.Context)
}

type messageHandler struct {
	repository repository.MessageRepository
}

func NewMessageHandler(repository *repository.MessageRepository) MessageHandler {
	return &messageHandler{repository: *repository}
}

func (handler *messageHandler) SaveMessage(context *gin.Context) {
	var message models.Message
	if err := context.ShouldBindBodyWith(&message, binding.JSON); err != nil {
		context.AbortWithStatusJSON(
			http.StatusBadRequest, models.Response{
				Message:    "Invalid JSON data: " + err.Error(),
				HttpStatus: http.StatusBadRequest,
				Success:    false,
			})
		return
	}

	response, err := handler.repository.Save(message)
	if err != nil {
		context.AbortWithStatusJSON(
			http.StatusBadRequest, models.Response{
				Message:    "Not able to send message: " + err.Error(),
				HttpStatus: http.StatusBadRequest,
				Success:    false,
			})
		return
	}

	context.IndentedJSON(http.StatusCreated, models.Response{
		Message:    "Successfully sent message: " + response.ID,
		HttpStatus: http.StatusCreated,
		Success:    true,
	})
}

func (handler *messageHandler) GetMessageById(context *gin.Context) {
	id := context.Param("id")

	message, err := handler.repository.GetById(id)
	if err != nil {
		context.AbortWithStatusJSON(
			http.StatusNotFound, models.Response{
				Message:    "No such message found with id " + id + ": " + err.Error(),
				HttpStatus: http.StatusNotFound,
				Success:    false,
			})
		return
	}

	context.IndentedJSON(http.StatusCreated, models.Response{
		Message:    "Successfully retrieved message with id: " + message.ID,
		HttpStatus: http.StatusOK,
		Success:    true,
		Data:       message,
	})
}

func (handler *messageHandler) GetMessagesByUserId(context *gin.Context) {
	id := context.Param("id")

	messages, err := handler.repository.GetAllByUserId(id)
	if err != nil {
		context.AbortWithStatusJSON(
			http.StatusNotFound, models.Response{
				Message:    "No such messages found with user id " + id + ": " + err.Error(),
				HttpStatus: http.StatusNotFound,
				Success:    false,
			})
		return
	}

	if len(messages) <= 0 {
		context.IndentedJSON(http.StatusOK, models.Response{
			Message:    "No messages found with user id: " + id,
			HttpStatus: http.StatusNotFound,
			Success:    false,
		})
		return
	}

	context.IndentedJSON(http.StatusOK, models.Response{
		Message:    "Successfully retrieved messages with user id: " + id,
		HttpStatus: http.StatusOK,
		Success:    true,
		Data:       messages,
	})
}

func (handler *messageHandler) DeleteMessagesByUserId(context *gin.Context) {
	id := context.Param("id")

	err := handler.repository.DeleteAllByUserId(id)
	if err != nil {
		context.AbortWithStatusJSON(
			http.StatusNotFound, models.Response{
				Message:    "No such messages found with user id " + id + ": " + err.Error(),
				HttpStatus: http.StatusNotFound,
				Success:    false,
			})
		return
	}

	context.IndentedJSON(http.StatusOK, models.Response{
		Message:    "Successfully deleted messages with user id: " + id,
		HttpStatus: http.StatusOK,
		Success:    true,
	})
}
