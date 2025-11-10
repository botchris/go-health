package rabbitmq

import (
	"fmt"
	"time"

	"github.com/botchris/go-health"
)

// Option defines a configuration option for the RabbitMQ Probe.
type Option func(*options) error

type options struct {
	dsn            string
	declareTopic   string
	declareQueue   string
	consumeTimeout time.Duration
	dialTimeout    time.Duration
}

var defaultOptions = options{
	declareTopic:   "",
	declareQueue:   "",
	consumeTimeout: 3 * time.Second,
	dialTimeout:    5 * time.Second,
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
