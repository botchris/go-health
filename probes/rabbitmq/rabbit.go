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
// - Dialing the RabbitMQ server
// - Declaring an exchange (if configured)
// - Declaring a queue (if configured)
// - Binding the queue to the exchange (if both are configured)
// - Publishing a message to the exchange and consuming it from the queue (if both are configured)
//
// The DSN is required and must be a valid RabbitMQ connection string.
// Additional options can be provided to customize the probe behavior.
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

	return &rabbitProbe{opts: opts}, nil
}

// Check executes the probe flow.
func (r *rabbitProbe) Check(ctx context.Context) (checkErr error) {
	conn, err := r.dial()
	if err != nil {
		return fmt.Errorf("rabbitmq dial failed: %w", err)
	}

	defer func() {
		if cErr := conn.Close(); cErr != nil && checkErr == nil {
			checkErr = fmt.Errorf("rabbitmq connection close: %w", cErr)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		checkErr = fmt.Errorf("open channel failed: %w", err)

		return
	}

	defer func() {
		if cErr := ch.Close(); cErr != nil && checkErr == nil {
			checkErr = fmt.Errorf("channel close: %w", cErr)
		}
	}()

	if r.opts.declareTopic != "" {
		if err := r.declareExchange(ch); err != nil {
			checkErr = err

			return
		}
	}

	if r.opts.declareQueue != "" {
		if err := r.declareQueue(ch); err != nil {
			checkErr = err

			return
		}
	}

	if r.opts.declareTopic != "" && r.opts.declareQueue != "" {
		if err := r.bindQueue(ch); err != nil {
			checkErr = err

			return
		}

		checkErr = r.publishAndConsume(ctx, ch)

		return
	}

	return nil
}

func (r *rabbitProbe) dial() (*amqp.Connection, error) {
	return amqp.DialConfig(r.opts.dsn, amqp.Config{
		Dial: amqp.DefaultDial(r.opts.dialTimeout),
	})
}

func (r *rabbitProbe) declareExchange(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(r.opts.declareTopic, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", r.opts.declareTopic, err)
	}

	return nil
}

func (r *rabbitProbe) declareQueue(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(r.opts.declareQueue, false, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %q: %w", r.opts.declareQueue, err)
	}

	return nil
}

func (r *rabbitProbe) bindQueue(ch *amqp.Channel) error {
	if err := ch.QueueBind(r.opts.declareQueue, routingKey, r.opts.declareTopic, false, nil); err != nil {
		return fmt.Errorf("bind queue %q to exchange %q: %w", r.opts.declareQueue, r.opts.declareTopic, err)
	}

	return nil
}

func (r *rabbitProbe) publishAndConsume(ctx context.Context, ch *amqp.Channel) error {
	msgs, err := ch.Consume(r.opts.declareQueue, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume start failed: %w", err)
	}

	payload := []byte(time.Now().Format(time.RFC3339Nano))
	if err := ch.Publish(r.opts.declareTopic, routingKey, false, false, amqp.Publishing{Body: payload}); err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	timer := time.NewTimer(r.opts.consumeTimeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled while waiting message: %w", ctx.Err())
	case <-timer.C:
		return fmt.Errorf("message not consumed within %s", r.opts.consumeTimeout)
	case _, ok := <-msgs:
		if !ok {
			return fmt.Errorf("consumer channel closed unexpectedly")
		}
	}

	return nil
}
