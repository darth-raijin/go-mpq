package repositories

import (
	"fmt"

	"github.com/darth-raijin/go-mpq/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Message interface {
	SaveMessage(message models.Message) error
	GetByID() error
}

type MessageRepositoryOptions struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(opts MessageRepositoryOptions) Message {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", opts.Username, opts.Password, opts.Host, opts.Port, opts.DBName)
	return &MessageRepository{
		db: sqlx.MustConnect("postgres", dsn),
	}
}

func (m MessageRepository) SaveMessage(message models.Message) error {
	//TODO implement me
	panic("implement me")
}

func (m MessageRepository) GetByID() error {
	//TODO implement me
	panic("implement me")
}
