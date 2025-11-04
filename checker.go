package health

import "context"

// Checker defines a component responsible for performing health checks.
// For example, it can be used to verify the health of various subsystems
// such as databases, external services, or internal components.
type Checker interface {
	// Check performs the health check and returns an error if the check fails.
	Check(ctx context.Context) error
}

// CheckFunc is an adapter to allow the use of ordinary functions as Checker.
type CheckFunc func(ctx context.Context) error

// Check calls f(ctx).
func (f CheckFunc) Check(ctx context.Context) error {
	return f(ctx)
}
