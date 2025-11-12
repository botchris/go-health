package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckSuccess_PublishAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "test.topic", "topic", true, false, false, false, mock.Anything).Return(nil)
	ch.On("QueueDeclare", "test.queue", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "test.queue"}, nil)
	ch.On("QueueBind", "test.queue", routingKey, "test.topic", false, mock.Anything).Return(nil)

	msgCh := make(chan amqp.Delivery, 1)
	ch.On("Consume", "test.queue", "", true, false, false, false, mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)
	ch.On("Publish", "test.topic", routingKey, false, false, mock.AnythingOfType("amqp091.Publishing")).
		Run(func(args mock.Arguments) {
			msg, ok := args.Get(4).(amqp.Publishing)
			if !ok {
				return
			}

			msgCh <- amqp.Delivery{Body: msg.Body}
		}).
		Return(nil)

	p, err := New("mock://",
		WithDialer(mockDialer(conn, nil)),
		WithDeclareTopic("test.topic"),
		WithDeclareQueue("test.queue"),
		WithConsumeTimeout(200*time.Millisecond),
	)

	assert.NoError(t, err)
	assert.NoError(t, p.Check(ctx))

	conn.AssertExpectations(t)
	ch.AssertExpectations(t)
}

func TestCheckDialFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p, err := New("mock://", WithDialer(mockDialer(nil, errors.New("dial boom"))))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "rabbitmq dial failed")
}

func TestCheckChannelOpenFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := new(mockConn)
	conn.On("Channel").Return(nil, errors.New("channel boom"))
	conn.On("Close").Return(nil)
	p, err := New("mock://", WithDialer(mockDialer(conn, nil)))
	assert.NoError(t, err)
	err = p.Check(ctx)
	assert.ErrorContains(t, err, "open channel failed")
}

func TestCheckExchangeDeclareFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "x", "topic", true, false, false, false, mock.Anything).Return(errors.New("exchange declare error"))

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("x"))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "declare exchange")
}

func TestCheckQueueDeclareFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{}, errors.New("queue declare error"))

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareQueue("q"))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "declare queue")
}

func TestCheckQueueBindFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "t", "topic", true, false, false, false, mock.Anything).Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "q"}, nil)
	ch.On("QueueBind", "q", routingKey, "t", false, mock.Anything).Return(errors.New("queue bind error"))

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("t"), WithDeclareQueue("q"))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "bind queue")
}

func TestCheckConsumeStartFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "t", "topic", true, false, false, false, mock.Anything).Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "q"}, nil)
	ch.On("QueueBind", "q", routingKey, "t", false, mock.Anything).Return(nil)
	ch.On("Consume", "q", "", true, false, false, false, mock.Anything).Return(nil, errors.New("consume start error"))

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("t"), WithDeclareQueue("q"))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "consume start failed")
}

func TestCheckPublishFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "t", "topic", true, false, false, false, mock.Anything).Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "q"}, nil)
	ch.On("QueueBind", "q", routingKey, "t", false, mock.Anything).Return(nil)

	msgCh := make(chan amqp.Delivery, 1)
	ch.On("Consume", "q", "", true, false, false, false, mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)
	ch.On("Publish", "t", routingKey, false, false, mock.AnythingOfType("amqp091.Publishing")).Return(errors.New("publish error"))

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("t"), WithDeclareQueue("q"))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "publish failed")
}

func TestCheckContextCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "t", "topic", true, false, false, false, mock.Anything).Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "q"}, nil)
	ch.On("QueueBind", "q", routingKey, "t", false, mock.Anything).Return(nil)

	msgCh := make(chan amqp.Delivery, 1)
	ch.On("Consume", "q", "", true, false, false, false, mock.Anything).Return((<-chan amqp.Delivery)(msgCh), nil)
	ch.On("Publish", "t", routingKey, false, false, mock.AnythingOfType("amqp091.Publishing")).Return(nil)

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("t"), WithDeclareQueue("q"), WithConsumeTimeout(200*time.Millisecond))
	assert.NoError(t, err)

	cancel()
	assert.ErrorContains(t, p.Check(ctx), "context canceled")
}

func TestCheckNoPublishWhenOnlyTopicConfigured(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("ExchangeDeclare", "t", "topic", true, false, false, false, mock.Anything).Return(nil)

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareTopic("t"))
	assert.NoError(t, err)
	assert.NoError(t, p.Check(ctx))
}

func TestCheckNoPublishWhenOnlyQueueConfigured(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(nil)
	ch.On("Close").Return(nil)
	ch.On("QueueDeclare", "q", false, false, false, false, mock.Anything).Return(amqp.Queue{Name: "q"}, nil)

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)), WithDeclareQueue("q"))
	assert.NoError(t, err)
	assert.NoError(t, p.Check(ctx))
}

func TestCheckConnectionCloseError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch := new(mockChannel)
	conn := new(mockConn)
	conn.ch = ch

	conn.On("Channel").Return(ch, nil)
	conn.On("Close").Return(errors.New("close boom"))
	ch.On("Close").Return(nil)

	p, err := New("mock://", WithDialer(mockDialer(conn, nil)))
	assert.NoError(t, err)

	err = p.Check(ctx)
	assert.ErrorContains(t, err, "rabbitmq connection close")
}

type mockChannel struct {
	mock.Mock
}

func (m *mockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	argsM := m.Called(name, kind, durable, autoDelete, internal, noWait, args)

	return argsM.Error(0)
}

func (m *mockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	argsM := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	q, ok := argsM.Get(0).(amqp.Queue)

	if !ok {
		return amqp.Queue{}, argsM.Error(1)
	}

	return q, argsM.Error(1)
}

func (m *mockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	argsM := m.Called(name, key, exchange, noWait, args)

	return argsM.Error(0)
}

func (m *mockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	argsM := m.Called(exchange, key, mandatory, immediate, msg)

	return argsM.Error(0)
}

func (m *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	argsM := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	c, ok := argsM.Get(0).(<-chan amqp.Delivery)

	if !ok {
		return nil, argsM.Error(1)
	}

	return c, argsM.Error(1)
}

func (m *mockChannel) Close() error {
	argsM := m.Called()

	return argsM.Error(0)
}

type mockConn struct {
	mock.Mock
	ch Channel
}

func (m *mockConn) Channel() (Channel, error) {
	argsM := m.Called()
	c, ok := argsM.Get(0).(Channel)

	if !ok {
		return nil, argsM.Error(1)
	}

	return c, argsM.Error(1)
}

func (m *mockConn) Close() error {
	argsM := m.Called()

	return argsM.Error(0)
}

func mockDialer(conn *mockConn, dialErr error) DialFunc {
	return func(dsn string, cfg amqp.Config) (Connection, error) {
		if dialErr != nil {
			return nil, dialErr
		}

		return conn, nil
	}
}
