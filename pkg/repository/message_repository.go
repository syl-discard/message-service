package repository

import (
	"discard/message-service/pkg/models"

	"github.com/gocql/gocql"
)

type MessageRepository interface {
	Save(message models.Message) (*models.Message, error)
	GetById(id string) (*models.Message, error)
	GetAllByUserId(userID string) ([]*models.Message, error)
	DeleteAllByUserId(userID string) error
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

func (repository *messageRepository) GetAllByUserId(userID string) ([]*models.Message, error) {
	var messages []*models.Message
	var query string = "SELECT * FROM messages WHERE UserID = ? ALLOW FILTERING"

	iter := repository.session.Query(query, userID).Iter()
	for {
		var message models.Message
		if !iter.Scan(&message.ID, &message.UserID, &message.ServerID, &message.Message) {
			break
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

func (repository *messageRepository) DeleteAllByUserId(userID string) error {
	var ids []string
	var query string = "SELECT ID FROM messages WHERE UserID = ? ALLOW FILTERING"
	iter := repository.session.Query(query, userID).Iter()
	for {
		var id string
		if !iter.Scan(&id) {
			break
		}
		ids = append(ids, id)
	}

	query = "DELETE FROM messages WHERE ID = ?"
	for _, id := range ids {
		if err := repository.session.Query(query, id).Exec(); err != nil {
			return err
		}
	}

	return nil
}

// === Integration Test ===
type inMemoryMessageRepository struct {
	messages []*models.Message
}

func NewInMemoryMessageRepository() MessageRepository {
	return &inMemoryMessageRepository{
		messages: make([]*models.Message, 0)}
}

func (repository *inMemoryMessageRepository) Save(message models.Message) (*models.Message, error) {
	repository.messages = append(repository.messages, &message)
	return &message, nil
}

func (repository *inMemoryMessageRepository) GetById(id string) (*models.Message, error) {
	for _, message := range repository.messages {
		if message.ID == id {
			return message, nil
		}
	}
	return nil, nil
}

func (repository *inMemoryMessageRepository) GetAllByUserId(userID string) ([]*models.Message, error) {
	var messages []*models.Message
	for _, message := range repository.messages {
		if message.UserID == userID {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

func (repository *inMemoryMessageRepository) DeleteAllByUserId(userID string) error {
	var messages []*models.Message
	for _, message := range repository.messages {
		if message.UserID != userID {
			messages = append(messages, message)
		}
	}
	repository.messages = messages
	return nil
}
