package main

import (
	"os"

	"github.com/darth-raijin/go-mpq/internal/handlers"
	"github.com/darth-raijin/go-mpq/internal/repositories"
	"go.uber.org/zap"
)

func main() {
	logger := zap.New(nil, zap.AddCaller())

	db := repositories.NewMessageRepository(repositories.MessageRepositoryOptions{
		Username: os.Getenv("DB_USER"),
		Password: "",
		Host:     "",
		Port:     "",
		DBName:   "",
	})

	handlers.NewMessage(handlers.MessageOptions{
		Logger:            logger,
		DB:                nil,
		MessageRepository: nil,
	})

}
