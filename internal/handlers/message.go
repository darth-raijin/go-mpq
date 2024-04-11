package handlers

import (
	gompq "github.com/darth-raijin/go-mpq/internal/protos"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

//go:generate mockery --name Message
type Message interface {
	HandleMessage(data []byte) error
}

type MessageOptions struct {
	Logger *zap.Logger
	DB     *sqlx.DB
}

type message struct {
	logger *zap.Logger
	db     *sqlx.DB
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
