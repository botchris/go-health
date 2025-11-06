package health

import "time"

// CheckerOption is a function that configures health check behavior.
type CheckerOption func(*checkerOptions) error

type checkerOptions struct {
	initialDelay     time.Duration
	period           time.Duration
	successThreshold int
	failureThreshold int
	reporterTimeout  time.Duration
}

var defaultCheckerOptions = checkerOptions{
	initialDelay:     0,
	period:           10 * time.Second,
	successThreshold: 1,
	failureThreshold: 3,
	reporterTimeout:  30 * time.Second,
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

// WithSuccessThreshold sets the number of consecutive successful
// checks required to consider the system healthy. The threshold
// must be at least 1. If a value less than 1 is provided, it
// defaults to 1.
func WithSuccessThreshold(threshold int) CheckerOption {
	return func(o *checkerOptions) error {
		if threshold < 1 {
			threshold = 1
		}

		o.successThreshold = threshold

		return nil
	}
}

// WithFailureThreshold sets the number of consecutive failed
// checks required to consider the system unhealthy. The threshold
// must be at least 1. If a value less than 1 is provided, it
// defaults to 1.
func WithFailureThreshold(threshold int) CheckerOption {
	return func(o *checkerOptions) error {
		if threshold < 1 {
			threshold = 1
		}

		o.failureThreshold = threshold

		return nil
	}
}

// WithReporterTimeout sets the timeout duration for reporter operations.
// The timeout must be at least 1 second. If a duration less than
// 1 second is provided, it is rounded up to 1 second.
func WithReporterTimeout(d time.Duration) CheckerOption {
	return func(o *checkerOptions) error {
		if d < time.Second {
			d = time.Second
		}

		o.reporterTimeout = d

		return nil
	}
}
