package rabbitmq

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// mockChannel implements the Channel interface for testing.
type mockChannel struct {
	failExchange bool
	failQueue    bool
	failBind     bool
	failConsume  bool
	failPublish  bool

	consumeCh     chan amqp.Delivery
	publishCalled bool
}

func (m *mockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	if m.failExchange {
		return errors.New("exchange declare error")
	}

	return nil
}

func (m *mockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if m.failQueue {
		return amqp.Queue{}, errors.New("queue declare error")
	}

	return amqp.Queue{Name: name}, nil
}

func (m *mockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	if m.failBind {
		return errors.New("queue bind error")
	}

	return nil
}

func (m *mockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.publishCalled = true
	if m.failPublish {
		return errors.New("publish error")
	}

	if m.consumeCh != nil {
		m.consumeCh <- amqp.Delivery{Body: msg.Body}
	}

	return nil
}

func (m *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if m.failConsume {
		return nil, errors.New("consume start error")
	}

	if m.consumeCh == nil {
		m.consumeCh = make(chan amqp.Delivery, 1)
	}

	return m.consumeCh, nil
}

func (m *mockChannel) Close() error {
	return nil
}

// mockConn implements the Connection interface for testing.
type mockConn struct {
	ch         *mockChannel
	channelErr error
	closeErr   error
	closed     bool
}

func (c *mockConn) Channel() (Channel, error) {
	if c.channelErr != nil {
		return nil, c.channelErr
	}

	return c.ch, nil
}

func (c *mockConn) Close() error {
	c.closed = true

	return c.closeErr
}

// helper to create a custom dialer returning a mockConn.
func mockDialer(conn *mockConn, dialErr error) DialFunc {
	return func(dsn string, cfg amqp.Config) (Connection, error) {
		if dialErr != nil {
			return nil, dialErr
		}

		return conn, nil
	}
}

func TestCheckSuccess(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("test.topic"),
		WithDeclareQueue("test.queue"),
		WithConsumeTimeout(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	if err := p.Check(context.Background()); err != nil {
		t.Fatalf("Check unexpected error: %v", err)
	}

	if !ch.publishCalled {
		t.Fatalf("expected publish to be called")
	}
}

func TestCheckDialFailure(t *testing.T) {
	p, err := New("mock://",
		WithDialer(mockDialer(nil, errors.New("dial boom"))),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "rabbitmq dial failed") {
		t.Fatalf("expected dial failed error, got %v", err)
	}
}

func TestCheckChannelOpenFailure(t *testing.T) {
	conn := &mockConn{channelErr: errors.New("channel boom")}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "open channel failed") {
		t.Fatalf("expected open channel failed error, got %v", err)
	}
}

func TestCheckExchangeDeclareFailure(t *testing.T) {
	ch := &mockChannel{failExchange: true}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("x"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "declare exchange") {
		t.Fatalf("expected exchange declare error, got %v", err)
	}
}

func TestCheckQueueDeclareFailure(t *testing.T) {
	ch := &mockChannel{failQueue: true}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareQueue("q"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "declare queue") {
		t.Fatalf("expected queue declare error, got %v", err)
	}
}

func TestCheckQueueBindFailure(t *testing.T) {
	ch := &mockChannel{failBind: true}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("t"),
		WithDeclareQueue("q"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "bind queue") {
		t.Fatalf("expected bind queue error, got %v", err)
	}
}

func TestCheckConsumeStartFailure(t *testing.T) {
	ch := &mockChannel{failConsume: true}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("t"),
		WithDeclareQueue("q"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "consume start failed") {
		t.Fatalf("expected consume start failed error, got %v", err)
	}
}

func TestCheckPublishFailure(t *testing.T) {
	ch := &mockChannel{failPublish: true}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("t"),
		WithDeclareQueue("q"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "publish failed") {
		t.Fatalf("expected publish failed error, got %v", err)
	}
}

func TestCheckContextCanceled(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("t"),
		WithDeclareQueue("q"),
		WithConsumeTimeout(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = p.Check(ctx)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context canceled error, got %v", err)
	}
}

func TestCheckNoPublishWhenOnlyTopicConfigured(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("t"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	if err := p.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ch.publishCalled {
		t.Fatalf("publish should not be called when queue is not configured")
	}
}

func TestCheckNoPublishWhenOnlyQueueConfigured(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConn{ch: ch}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareQueue("q"),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	if err := p.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ch.publishCalled {
		t.Fatalf("publish should not be called when topic is not configured")
	}
}

func TestCheckConnectionCloseError(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConn{
		ch:       ch,
		closeErr: errors.New("close boom"),
	}

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
	)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "rabbitmq connection close") {
		t.Fatalf("expected connection close error, got %v", err)
	}
}
