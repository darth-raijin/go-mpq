package messagers

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

//go:generate mockery --name ProducerInterface --inpackage
type ProducerInterface interface {
	PublishMessage(ctx context.Context, contentType, routingKey, queueName string, body []byte) error
	Close() error
}

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger

	exchangeName string
	exchangeType string
	queueName    string
}

type ProducerOptions struct {
	DialURL string
	Channel *amqp.Channel
	Logger  *zap.Logger

	ExchangeName string
	ExchangeType string
	QueueName    string
	RoutingKey   string
}

func NewProducer(opts ProducerOptions) (ProducerInterface, error) {
	conn, err := amqp.Dial(opts.DialURL)
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

	if err := setupQueue(channel, opts.QueueName, opts.ExchangeName, opts.RoutingKey); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	opts.Logger.Info("Producer connected and exchange declared",
		zap.String("exchange", opts.ExchangeName),
		zap.String("queue", opts.QueueName))

	return &Producer{
		conn:         conn,
		channel:      channel,
		logger:       opts.Logger,
		exchangeName: opts.ExchangeName,
		queueName:    opts.QueueName,
	}, nil
}

func setupQueue(ch *amqp.Channel, queueName, exchangeName, routingKey string) error {
	_, err := ch.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	// If no routing key is provided, we don't need to bind the queue to the exchange.
	// We use the default exchange in this case.
	if routingKey != "" {
		err = ch.QueueBind(
			queueName,    // queue name
			routingKey,   // routing key
			exchangeName, // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return err
		}
	}

	return nil
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

// PublishMessage sends a message to the specified queue.
func (p *Producer) PublishMessage(ctx context.Context, contentType, routingKey, queueName string, body []byte) error {
	err := p.channel.PublishWithContext(
		ctx,            // context
		p.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish message", zap.Error(err))
		return err
	}

	p.logger.Info("Message published", zap.String("queue", queueName), zap.String("routing_key", routingKey))
	return nil
}

func (p *Producer) Close() error {
	var err error
	if p.channel != nil {
		err = p.channel.Close()
	}
	if p.conn != nil {
		err = p.conn.Close()
	}
	if err != nil {
		p.logger.Error("Failed to close producer connection", zap.Error(err))
		return err
	}

	p.logger.Info("Producer connection closed")
	return nil
}
