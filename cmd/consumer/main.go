package main

import (
	"context"
	"github.com/darth-raijin/go-mpq/internal/config"
	"github.com/darth-raijin/go-mpq/internal/handlers"
	"github.com/darth-raijin/go-mpq/internal/messagers"
	"github.com/darth-raijin/go-mpq/internal/repositories"
	"go.uber.org/zap"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	logger := zap.New(nil)

	messageRepository := repositories.NewMessageRepository(repositories.MessageRepositoryOptions{
		Username: config.DBUsername,
		Password: config.DBPassword,
		Host:     config.DBHost,
		Port:     config.DBPort,
		DBName:   config.DBName,
	})

	messageHandler := handlers.NewMessage(handlers.MessageOptions{
		Logger:            logger,
		MessageRepository: messageRepository,
	})

	consumer, err := messagers.NewConsumer(messagers.ConsumerOptions{
		Logger:         logger,
		MessageHandler: messageHandler,
		Host:           config.RabbitMQHost,
		Port:           config.RabbitMQPort,
		Username:       config.RabbitMQUser,
		Password:       config.RabbitMQPass,
		VHost:          config.RabbitMQVHost,
		WorkerCount:    1,
	})
	if err != nil {
		panic(err)
	}

	err = consumer.StartListening(context.Background(),
		"events",
		"transactions",
		"transactions.created")
	if err != nil {
		return
	}

}

func wireDependencies() {

}
