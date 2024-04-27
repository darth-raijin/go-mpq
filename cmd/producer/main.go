package main

import (
	"context"
	"github.com/darth-raijin/go-mpq/internal/config"
	"github.com/darth-raijin/go-mpq/internal/messagers"
	gompq "github.com/darth-raijin/go-mpq/internal/protos"
	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := zap.New(nil, zap.AddCaller())

	configVariables, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("error loading configVariables", zap.Error(err))
		os.Exit(1)
	}

	producer, err := wireDependencies(configVariables, logger, err)
	if err != nil {
		logger.Error("error wiring dependencies", zap.Error(err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fake := faker.New()

	for i := 0; i < configVariables.ProducerWorkers; i++ {
		go func() {
			for {
				inputMessage := gompq.Message{
					Id:        uuid.NewString(),
					Timestamp: timestamppb.Now(),
					Body:      fake.Person().FirstName(),
				}

				if err := producer.PublishMessage(ctx, "transactions.created", &inputMessage); err != nil {
					logger.Error("error publishing message", zap.Error(err))
					stop()
				}
			}
		}()
	}

	<-ctx.Done()
	logger.Info("shutting down gracefully")

}

func wireDependencies(config config.Config, logger *zap.Logger, err error) (messagers.ProducerInterface, error) {
	return messagers.NewProducer(messagers.ProducerOptions{
		Host:         config.RabbitMQHost,
		VHost:        config.RabbitMQVHost,
		Port:         config.RabbitMQPort,
		Username:     config.RabbitMQUser,
		Password:     config.RabbitMQPass,
		Logger:       logger,
		ExchangeName: "events",
		ExchangeType: "topic",
	})
}
