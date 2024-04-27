package messagers

import (
	"context"
	"fmt"
	gompq "github.com/darth-raijin/go-mpq/internal/protos"
	"google.golang.org/protobuf/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ProducerInterface defines the methods a producer should have.
type ProducerInterface interface {
	PublishMessage(ctx context.Context, routingKey string, msg *gompq.Message) error
	Close() error
}

// Producer represents a message producer for RabbitMQ.
type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger

	exchangeName string
}

// ProducerOptions holds the configuration options for setting up the Producer.
type ProducerOptions struct {
	Host         string
	VHost        string
	Port         string
	Username     string
	Password     string
	Logger       *zap.Logger
	ExchangeName string
	ExchangeType string
}

// NewProducer creates and returns a new Producer based on the provided options.
func NewProducer(opts ProducerOptions) (ProducerInterface, error) {
	if opts.ExchangeName == "" || opts.ExchangeType == "" {
		return nil, fmt.Errorf("exchange name and type must not be empty")
	}

	connURL := fmt.Sprintf("amqp://%v:%v@%v:%v/%v", opts.Username, opts.Password, opts.Host, opts.Port, opts.VHost)
	conn, err := amqp.Dial(connURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := declareExchange(channel, opts.ExchangeName, opts.ExchangeType); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	opts.Logger.Info("Producer connected and exchange declared",
		zap.String("exchange", opts.ExchangeName))

	return &Producer{
		conn:         conn,
		channel:      channel,
		logger:       opts.Logger,
		exchangeName: opts.ExchangeName,
	}, nil
}

// declareExchange declares a new exchange on the channel.
func declareExchange(ch *amqp.Channel, name, kind string) error {
	return ch.ExchangeDeclare(
		name,  // exchange name
		kind,  // exchange type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// PublishMessage sends a message to the specified exchange with the given routing key.
func (p *Producer) PublishMessage(ctx context.Context, routingKey string, msg *gompq.Message) error {
	// Serialize protobuf message to byte array
	body, err := proto.Marshal(msg)
	if err != nil {
		p.logger.Error("Failed to marshal protobuf message", zap.Error(err))
		return err
	}

	// Publish message
	err = p.channel.PublishWithContext(
		ctx,            // context
		p.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish message", zap.Error(err))
		return err
	}

	p.logger.Info("Protobuf message published", zap.String("routing_key", routingKey))
	return nil
}

// Close closes the channel and connection to RabbitMQ.
func (p *Producer) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Error("Failed to close channel", zap.Error(err))
			if p.conn != nil {
				p.conn.Close() // Attempt to close connection anyway
			}
			return err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			p.logger.Error("Failed to close connection", zap.Error(err))
			return err
		}
	}
	p.logger.Info("Producer connection closed")
	return nil
}
