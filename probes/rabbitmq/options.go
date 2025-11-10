package rabbitmq

import (
	"fmt"
	"time"

	"github.com/botchris/go-health"
)

// Option defines a configuration option for the RabbitMQ Probe.
type Option func(*options) error

type options struct {
	dsn string
}
