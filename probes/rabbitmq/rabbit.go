package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/botchris/go-health"
	amqp "github.com/rabbitmq/amqp091-go"
)

const routingKey = "#"

type rabbitProbe struct {
	opts *options
}

// New builds a new Probe capable of verifying the following RabbitMQ aspects:
//
// - Connection establishing
// - Getting a "Channel" from the connection
// - Declaring topic exchange
// - Declaring queue
// - Binding a queue to the exchange with the defined routing key
// - Publishing a message to the exchange with the defined routing key
// - Consuming published message
func New(dsn string, o ...Option) (health.Probe, error) {
	opts := &options{
		dsn:            dsn,
		consumeTimeout: defaultOptions.consumeTimeout,
		dialTimeout:    defaultOptions.dialTimeout,
	}

	for i := range o {
		if oErr := o[i](opts); oErr != nil {
			return nil, oErr
		}
	}

	return rabbitProbe{opts: opts}, nil
}

func (r rabbitProbe) Check(ctx context.Context) (checkErr error) {
	conn, dErr := amqp.DialConfig(r.opts.dsn, amqp.Config{Dial: amqp.DefaultDial(r.opts.dialTimeout)})
	if dErr != nil {
		checkErr = fmt.Errorf("failed to connect to RabbitMQ: %w", dErr)

		return
	}

	defer func() {
		if err := conn.Close(); err != nil && checkErr == nil {
			checkErr = fmt.Errorf("failed to close RabbitMQ connection: %w", err)
		}
	}()

	ch, cErr := conn.Channel()
	if cErr != nil {
		checkErr = fmt.Errorf("failed to open a channel: %w", cErr)

		return
	}

	defer func() {
		if err := ch.Close(); err != nil && checkErr == nil {
			checkErr = fmt.Errorf("failed to close RabbitMQ channel: %w", err)
		}
	}()

	if r.opts.declareTopic != "" {
		if err := ch.ExchangeDeclare(r.opts.declareTopic, "topic", true, false, false, false, nil); err != nil {
			checkErr = fmt.Errorf("failed to declare exchange: %w", err)

			return
		}
	}

	if r.opts.declareQueue != "" {
		if _, err := ch.QueueDeclare(r.opts.declareQueue, false, false, false, false, nil); err != nil {
			checkErr = fmt.Errorf("failed to declare queue: %w", err)

			return
		}
	}

	if r.opts.declareTopic != "" && r.opts.declareQueue != "" {
		if err := ch.QueueBind(r.opts.declareQueue, routingKey, r.opts.declareTopic, false, nil); err != nil {
			checkErr = fmt.Errorf("failed to bind queue to exchange: %w", err)

			return
		}

		checkErr = r.consumeCheck(ctx, ch)
	}
	
	return
}

func (r rabbitProbe) consumeCheck(ctx context.Context, ch *amqp.Channel) error {
	messages, err := ch.Consume(r.opts.declareQueue, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	done := make(chan struct{})

	go func() {
		<-messages

		close(done)
	}()

	p := amqp.Publishing{Body: []byte(time.Now().Format(time.RFC3339Nano))}
	if pErr := ch.Publish(r.opts.declareTopic, routingKey, false, false, p); pErr != nil {
		return fmt.Errorf("failed to publish message: %w", pErr)
	}

	for {
		select {
		case <-time.After(r.opts.consumeTimeout):
			return fmt.Errorf("failed to consume message within %s", r.opts.consumeTimeout)
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting to consume message: %w", ctx.Err())
		case <-done:
			return nil
		}
	}
}
