package grpchealth

// Option is a functional option for the gRPC health reporter.
type Option func(*options)

type options struct {
	serviceNames []string
	server       HealthServer
}

// WithServiceNames sets the service names to be set in the gRPC health reporter.
// One or more service names can be provided, including the empty string to set
// the overall server status.
//
// If no service name is provided, the status will be reported for the overall server,
// that is an empty service name.
func WithServiceNames(names ...string) Option {
	return func(o *options) {
		o.serviceNames = names
	}
}
