package messagers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/darth-raijin/go-mpq/internal/handlers"
	"golang.org/x/sync/errgroup"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

//go:generate mockery --name Consumer --inpackage
type Consumer interface {
	StartListening(ctx context.Context, exchangeName, queueName, routingKey string) error
	StopListening() error
	isRetryable(err error) bool
	ProcessMessage(ctx context.Context, d amqp.Delivery) error
	initializeExchange(exchangeName string) error
}

type rabbitMQConsumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	messageHandler handlers.Message
	workerCount    int
	logger         *zap.Logger
}

type ConsumerOptions struct {
	Host     string
	VHost    string
	Port     string
	Username string
	Password string

	MessageHandler handlers.Message
	WorkerCount    int
	Logger         *zap.Logger
}

func NewConsumer(opts ConsumerOptions) (Consumer, error) {
	connURL := fmt.Sprintf("amqp://%v:%v@%v:%v/%v", opts.Username, opts.Password, opts.Host, opts.Port, opts.VHost)
	conn, err := amqp.Dial(connURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // proper error handling
		return nil, err
	}

	opts.Logger.Info("Connected to RabbitMQ", zap.String("url", connURL))
	return &rabbitMQConsumer{
		conn:           conn,
		channel:        ch,
		messageHandler: opts.MessageHandler,
		workerCount:    opts.WorkerCount,
		logger:         opts.Logger,
	}, nil
}

func (c *rabbitMQConsumer) StartListening(ctx context.Context, exchangeName, queueName, routingKey string) error {
	if err := c.initializeExchange(exchangeName); err != nil {
		c.logger.Error("Failed to declare exchange", zap.Error(err))
		return err
	}

	err := c.initializeQueue(queueName, routingKey, exchangeName)
	if err != nil {
		return err
	}

	messages, err := c.channel.ConsumeWithContext(
		ctx,
		queueName,
		"go-mpq",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < c.workerCount; i++ {
		g.Go(func() error {
			for {
				select {
				case d, ok := <-messages:
					if !ok {
						return nil // stopping the goroutine if channel is closed
					}
					if err := c.ProcessMessage(ctx, d); err != nil {
						c.logger.Error("Error processing message", zap.Error(err))
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
		c.logger.Info("Started worker", zap.Int("worker_id", i+1))
	}

	c.logger.Info("Started listening for messages")
	return g.Wait()
}

func (c *rabbitMQConsumer) initializeQueue(queueName string, routingKey string, exchangeName string) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *rabbitMQConsumer) ProcessMessage(ctx context.Context, d amqp.Delivery) error {
	const maxRetries = 5
	var retryCount int

	for {
		if retryCount > maxRetries {
			return d.Nack(false, true)
		}

		err := c.messageHandler.HandleMessage(d.Body)
		if err == nil {
			return d.Ack(false)
		}

		// If the error is not retryable we just nack the message and remove it from the queue
		if !c.isRetryable(err) {
			return d.Nack(false, false)
		}

		c.logger.Error("Error processing message, will retry", zap.Error(err))

		// Backing off with jitter
		// Sleep is based on the formula 2^retryCount * 1 second
		backoff := time.Duration(1<<uint(retryCount)) * time.Second
		select {
		case <-time.After(backoff):
			// Retry the message
		case <-ctx.Done():
			return ctx.Err()
		}

		retryCount++
	}
}

func (c *rabbitMQConsumer) StopListening() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}

	c.logger.Info("Stopped listening and closed connection")
	return nil
}

func (c *rabbitMQConsumer) isRetryable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// We could implement domain specific errors here and expand our retryability criteria
	// For some errors like "queue not found" we might not want to retry
	// If errors are caused by invalid data it is not retryable

	return false
}

func (c *rabbitMQConsumer) initializeExchange(exchangeName string) error {
	return c.channel.ExchangeDeclare(
		exchangeName, // name of the exchange
		"topic",      // type of exchange
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
}
