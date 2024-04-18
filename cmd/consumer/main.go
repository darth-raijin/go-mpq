package main

import (
	"context"
	"github.com/darth-raijin/go-mpq/internal/config"
	"github.com/darth-raijin/go-mpq/internal/handlers"
	"github.com/darth-raijin/go-mpq/internal/messagers"
	"github.com/darth-raijin/go-mpq/internal/repositories"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger, _ := zap.NewProduction()

	configVariables, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("error loading configVariables", zap.Error(err))
		os.Exit(1)
	}

	consumer := wireDependencies(configVariables, logger, err)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := consumer.StartListening(ctx, "events", "transactions", "transactions.created"); err != nil {
			logger.Error("error in consumer listening", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully")

	err = consumer.StopListening()
	if err != nil {
		os.Exit(1)
	}
	logger.Sync()

	os.Exit(0)
}

func wireDependencies(config config.Config, logger *zap.Logger, err error) messagers.Consumer {
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
		logger.Error("error creating consumer", zap.Error(err))
		os.Exit(1)
	}
	return consumer
}
