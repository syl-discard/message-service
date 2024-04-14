package repository

import (
	"discard/message-service/pkg/models"

	"github.com/gocql/gocql"
)

type MessageRepository interface {
	Save(message models.Message) (*models.Message, error)
	GetById(id string) (*models.Message, error)
}

type messageRepository struct { //_private
	session *gocql.Session
}

func NewMessageRepository(session *gocql.Session) MessageRepository {
	return &messageRepository{session: session}
}

func (repository *messageRepository) Save(message models.Message) (*models.Message, error) {
	var query string = "INSERT INTO messages (ID, UserID, ServerID, Message) VALUES (?, ?, ?, ?)"

	uuid := gocql.TimeUUID() // ignore provided ID if provided
	message.ID = uuid.String()

	if err := repository.session.Query(
		query, uuid, message.UserID, message.ServerID, message.Message).Exec(); err != nil {
		return nil, err
	}

	return &message, nil
}

func (repository *messageRepository) GetById(id string) (*models.Message, error) {
	var message models.Message
	var query string = "SELECT ID, UserID, ServerID, Message FROM messages WHERE ID = ?"

	if err := repository.session.Query(query, id).Scan(
		&message.ID, &message.UserID, &message.ServerID, &message.Message); err != nil {
		return nil, err
	}

	return &message, nil
}