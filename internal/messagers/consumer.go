package messagers

import (
	"github.com/darth-raijin/go-mpq/internal/handlers"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

//go:generate mockery --name Consumer
type Consumer interface {
	StartListening(queueName string) error
	StopListening() error
}

// rabbitMQConsumer implements the Consumer interface for RabbitMQ.
type rabbitMQConsumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	messageHandler handlers.Message
	workerCount    int
	stopChan       chan bool
	logger         *zap.Logger
}

type ConsumerOptions struct {
	RabbitMQURL    string
	MessageHandler handlers.Message
	WorkerCount    int
	StopChan       chan bool
	Logger         *zap.Logger
}

// NewConsumer initializes and returns a Consumer with RabbitMQ as the underlying message broker
func NewConsumer(opts ConsumerOptions) (Consumer, error) {
	conn, err := amqp.Dial(opts.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	opts.Logger.Info("Connected to RabbitMQ", zap.String("url", opts.RabbitMQURL))
	return &rabbitMQConsumer{
		conn:           conn,
		channel:        ch,
		messageHandler: opts.MessageHandler,
		workerCount:    opts.WorkerCount,
		stopChan:       opts.StopChan,
	}, nil
}

// StartListening begins listening for messages on the specified queue and processes them concurrently.
func (c *rabbitMQConsumer) StartListening(queueName string) error {
	messages, err := c.channel.Consume(
		queueName,
		"go-mpq", // consumer tag - unique identifier for the consumer
		false,
		false,
		false, // no-local - false means receive messages sent from this connection too
		false, // no-wait - false means the server will respond to commands
		nil,   // arguments - additional command arguments
	)
	if err != nil {
		return err
	}

	for i := 0; i < c.workerCount; i++ {
		go func() {
			for {
				select {
				case d, ok := <-messages:
					if !ok {
						return
					}
					if err := c.messageHandler.HandleMessage(d.Body); err != nil {
						c.logger.Error("Error processing message", zap.Error(err))
						d.Nack(false, true) // multiple: false, requeue: true
						continue
					}
					d.Ack(false) // multiple: false
				case <-c.stopChan:
					c.logger.Info("Consumer stopped listening")
					return
				}
			}
		}()
	}

	log.Println("Consumer started listening...")

	return nil
}

func (c *rabbitMQConsumer) StopListening() error {
	close(c.stopChan)

	if err := c.channel.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}

	c.logger.Info("Stopped listening for messages :o")
	return nil
}
