package health

import "time"

// CheckerOption is a function that configures health check behavior.
type CheckerOption func(*checkerOptions) error

type checkerOptions struct {
	initialDelay time.Duration
	period       time.Duration
}

var defaultCheckerOptions = checkerOptions{
	initialDelay: 0,
	period:       10 * time.Second,
}

// WithInitialDelay sets an initial delay before the first health check is performed
// Valid values are 0, or >= 1 second. If the provided duration is in the range (0, 1s),
// it is rounded up to 1 second.
func WithInitialDelay(d time.Duration) CheckerOption {
	return func(o *checkerOptions) error {
		if d < 0 {
			d = 0
		}

		if d > 0 && d < time.Second {
			d = time.Second
		}

		o.initialDelay = d

		return nil
	}
}

// WithPeriod sets the period between consecutive health checks.
// The period must be at least 1 second. If a duration less than
// 1 second is provided, it is rounded up to 1 second.
func WithPeriod(d time.Duration) CheckerOption {
	return func(o *checkerOptions) error {
		if d < time.Second {
			d = time.Second
		}

		o.period = d

		return nil
	}
}
