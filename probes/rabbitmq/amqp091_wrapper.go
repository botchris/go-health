package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection defines an interface for AMQP connections.
type connectionWrapper struct {
	conn *amqp.Connection
}

func (c *connectionWrapper) Channel() (Channel, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &channelWrapper{ch: ch}, nil
}

func (c *connectionWrapper) Close() error {
	return c.conn.Close()
}

// Channel defines an interface for AMQP channels.
type channelWrapper struct {
	ch *amqp.Channel
}

func (ch *channelWrapper) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return ch.ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (ch *channelWrapper) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return ch.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (ch *channelWrapper) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return ch.ch.QueueBind(name, key, exchange, noWait, args)
}

func (ch *channelWrapper) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return ch.ch.Publish(exchange, key, mandatory, immediate, msg)
}

func (ch *channelWrapper) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return ch.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (ch *channelWrapper) Close() error {
	return ch.ch.Close()
}
