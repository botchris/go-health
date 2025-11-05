package health

import "context"

// Probe defines a component responsible for performing health checks.
// The Check method should return nil if the check is successful,
// or an error if the check fails.
type Probe interface {
	// Check performs the health check and returns an error if the check fails.
	Check(ctx context.Context) error
}

// ProbeFunc is an adapter to allow the use of ordinary functions as Probe.
type ProbeFunc func(ctx context.Context) error

// Check calls f(ctx).
func (f ProbeFunc) Check(ctx context.Context) error {
	return f(ctx)
}
