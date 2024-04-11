package repositories

import "github.com/jmoiron/sqlx"

type Message interface {
	SaveMessage() error
	GetByID() error
}

type MessageRepositoryOptions struct {
	DB *sqlx.DB
}

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(opts MessageRepositoryOptions) Message {
	return &MessageRepository{
		db: opts.DB,
	}
}

func (m MessageRepository) SaveMessage() error {
	//TODO implement me
	panic("implement me")
}

func (m MessageRepository) GetByID() error {
	//TODO implement me
	panic("implement me")
}
