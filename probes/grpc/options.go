package grpc

import (
	"errors"

	"google.golang.org/grpc"
)

// Option defines a configuration option for the gRPC Probe.
type Option func(*options) error

type options struct {
	addr        string
	dialOptions []grpc.DialOption
	serviceName *string
}

// WithDialOptions sets custom gRPC dial options for the gRPC health check.
func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(o *options) error {
		if len(opts) == 0 {
			return errors.New("at least one dial option must be provided")
		}

		o.dialOptions = opts

		return nil
	}
}

// WithServiceName sets the service name to be checked in the gRPC health check.
func WithServiceName(name string) Option {
	return func(o *options) error {
		o.serviceName = &name

		return nil
	}
}
