package handlers

import (
	gompq "github.com/darth-raijin/go-mpq/internal/protos"
	"github.com/darth-raijin/go-mpq/internal/repositories"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

//go:generate mockery --name Message --inpackage
type Message interface {
	HandleMessage(data []byte) error
}

type MessageOptions struct {
	Logger            *zap.Logger
	DB                *sqlx.DB
	MessageRepository repositories.Message
}

type message struct {
	logger            *zap.Logger
	db                *sqlx.DB
	messageRepository repositories.Message
}

func NewMessage(opts MessageOptions) Message {
	return &message{
		logger: opts.Logger,
		db:     opts.DB,
	}
}

func (m message) HandleMessage(data []byte) error {
	var msg gompq.Message
	if err := proto.Unmarshal(data, &msg); err != nil {
		m.logger.Error("Failed to unmarshal message", zap.Error(err))
		return err
	}

	return nil
}
