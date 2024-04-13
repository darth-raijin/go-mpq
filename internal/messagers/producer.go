package messagers

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Producer is responsible for sending messages to a RabbitMQ queue.
type Producer interface {
	PublishMessage(queueName string, body []byte) error
	Close() error
}

// rabbitMQProducer implements the Producer interface for RabbitMQ.
type rabbitMQProducer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
}

// NewProducer initializes and returns a new instance of a rabbitMQProducer.
func NewProducer(amqpURL string, logger *zap.Logger) (Producer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to RabbitMQ", zap.String("url", amqpURL))

	return &rabbitMQProducer{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

// PublishMessage sends a message to the specified queue.
func (p *rabbitMQProducer) PublishMessage(queueName string, body []byte) error {
	_, err := p.channel.QueueDeclare(
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

	err = p.channel.Publish(
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	p.logger.Info("Message published", zap.String("queue", queueName), zap.ByteString("body", body))

	return nil
}

// Close closes the channel and connection to RabbitMQ.
func (p *rabbitMQProducer) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	if err := p.conn.Close(); err != nil {
		return err
	}

	p.logger.Info("RabbitMQ connection closed")
	return nil
}
