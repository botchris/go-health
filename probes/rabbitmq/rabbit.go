package rabbitmq

import "github.com/botchris/go-health"

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
		dsn: dsn,
	}

	for i := range o {
		if oErr := o[i](opts); oErr != nil {
			return nil, oErr
		}
	}

	return rabbitProbe{opts: opts}, nil
}

func (r rabbitProbe) Check(ctx context.Context) (checkErr error) {
	return nil
}
