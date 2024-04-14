package messagers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/darth-raijin/go-mpq/internal/handlers"
	"github.com/darth-raijin/go-mpq/internal/messagers"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewConsumer(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		connectionURL string
		expectError   bool
	}{
		{
			name:          "Valid connection URL",
			connectionURL: strings.ReplaceAll(rabbitMQConnectionString, "guest", "invalid"),
			expectError:   false,
		},
		{
			name:          "Invalid connection URL",
			connectionURL: "amqp://invalid:invalid@localhost:5672/",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call NewConsumer function
			consumer, err := messagers.NewConsumer(messagers.ConsumerOptions{
				RabbitMQURL: tt.connectionURL,
			})

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				assert.Nil(t, consumer, "Consumer should be nil on error")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
				assert.NotNil(t, consumer, "Consumer should not be nil")
			}
		})
	}
}

func TestRabbitMQConsumer_StartListening(t *testing.T) {
	t.Run("can listen and receive messages", func(t *testing.T) {
		mockMessageHandler := handlers.NewMockMessage(t)
		done := make(chan bool)

		mockMessageHandler.On("HandleMessage", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			done <- true
		})

		consumer, err := messagers.NewConsumer(messagers.ConsumerOptions{
			RabbitMQURL:    rabbitMQConnectionString,
			MessageHandler: mockMessageHandler,
			WorkerCount:    1,
			Logger:         zap.NewNop(),
		})
		if err != nil {
			t.Fatalf("Failed to create consumer: %v", err)
		}
		defer consumer.StopListening()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		expectedExchange := uuid.NewString()
		expectedQueue := uuid.NewString()
		expectedRoutingKey := strings.ReplaceAll(uuid.NewString(), "-", ".")
		expectedMessage := uuid.NewString()

		go func() {
			time.Sleep(1 * time.Second) // Allow consumer to start
			publishTestMessage(t, expectedExchange, expectedRoutingKey, expectedMessage, ctx)
		}()

		go func() {
			err = consumer.StartListening(ctx, expectedExchange, expectedQueue, expectedRoutingKey)
			assert.NoError(t, err)
			if err != nil {
				cancel()
				t.Fail()
			}
		}()

		select {
		case <-done:
			err := consumer.StopListening()
			assert.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("Test timed out")
		}

		mockMessageHandler.AssertNumberOfCalls(t, "HandleMessage", 1)
		mockMessageHandler.AssertCalled(t, "HandleMessage", []byte(expectedMessage))
	})
}

func publishTestMessage(t *testing.T, exchangeName, routingKey, message string, ctx context.Context) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQConnectionString)
	assert.NoError(t, err)
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	assert.NoError(t, err)
	defer ch.Close()

	// Declare the exchange as a topic exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name of the exchange
		"topic",      // type of exchange
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	assert.NoError(t, err)

	// Publish the message
	err = ch.PublishWithContext(
		ctx,
		exchangeName, // exchange name
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	assert.NoError(t, err)
}
