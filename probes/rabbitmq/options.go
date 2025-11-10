package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Option defines a configuration option for the RabbitMQ Probe.
type Option func(*options) error

// DialFunc defines a function type for dialing a RabbitMQ server.
type DialFunc func(dsn string, config amqp.Config) (Connection, error)

type options struct {
	dsn            string
	dialer         DialFunc
	declareTopic   string
	declareQueue   string
	consumeTimeout time.Duration
	dialTimeout    time.Duration
}

var defaultOptions = options{
	dialer:         defaultDialer,
	consumeTimeout: 3 * time.Second,
	dialTimeout:    5 * time.Second,
}

// WithDialer allows providing a custom dialer function for establishing the RabbitMQ connection.
func WithDialer(d DialFunc) Option {
	return func(o *options) error {
		if d == nil {
			return fmt.Errorf("dialer cannot be nil")
		}

		o.dialer = d

		return nil
	}
}

// WithDeclareTopic attempts to declare a topic exchange with the given name.
// If an empty string is provided, an error is returned.
func WithDeclareTopic(topic string) Option {
	return func(o *options) error {
		if topic == "" {
			return fmt.Errorf("declare topic exchange name cannot be empty")
		}

		o.declareTopic = topic

		return nil
	}
}

// WithDeclareQueue attempts to declare a queue with the given name.
// If an empty string is provided, an error is returned.
func WithDeclareQueue(queue string) Option {
	return func(o *options) error {
		if queue == "" {
			return fmt.Errorf("declare queue name cannot be empty")
		}

		o.declareQueue = queue

		return nil
	}
}

// WithConsumeTimeout overrides the timeout used while waiting for a consumed message.
func WithConsumeTimeout(d time.Duration) Option {
	return func(o *options) error {
		if d <= 0 {
			return fmt.Errorf("consume timeout must be > 0")
		}

		o.consumeTimeout = d

		return nil
	}
}

// WithDialTimeout overrides the timeout used when establishing the connection.
func WithDialTimeout(d time.Duration) Option {
	return func(o *options) error {
		if d <= 0 {
			return fmt.Errorf("dial timeout must be > 0")
		}

		o.dialTimeout = d

		return nil
	}
}
