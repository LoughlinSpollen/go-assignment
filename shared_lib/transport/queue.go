package transport

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQ abstraction required for unit testing
type MQConnector interface {
	Dial(connectionString string) (MQConnection, error)
}

func NewMQConnector() MQConnector {
	return &msgQConnector{}
}

type msgQConnector struct{}

func (r *msgQConnector) Dial(connectionString string) (MQConnection, error) {
	conn, err := amqp091.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return &msgQConnection{conn}, nil
}

type MQConnection interface {
	Channel() (MQChannel, error)
	Close() error
}

type MQChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
	Close() error
}

type msgQConnection struct {
	conn *amqp091.Connection
}

func (r *msgQConnection) Channel() (MQChannel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	return &msgQChannel{ch}, nil
}

func (r *msgQConnection) Close() error {
	return r.conn.Close()
}

type msgQChannel struct {
	ch *amqp091.Channel
}

func (r *msgQChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error) {
	return r.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (r *msgQChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
	return r.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (r *msgQChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	return r.ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

func (r *msgQChannel) Close() error {
	return r.ch.Close()
}
