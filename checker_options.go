package health

import "time"

// CheckerOption is a function that configures health check behavior.
type CheckerOption func(*checkerOptions) error

type checkerOptions struct {
	initialDelay        time.Duration
	period              time.Duration
	successThreshold    int
	failureThreshold    int
	probeDefaultTimeout time.Duration
	reporterTimeout     time.Duration
	bufferSize          int
}

var defaultCheckerOptions = checkerOptions{
	initialDelay:        0,
	period:              10 * time.Second,
	successThreshold:    1,
	failureThreshold:    3,
	probeDefaultTimeout: 5 * time.Second,
	reporterTimeout:     30 * time.Second,
	bufferSize:          10,
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

// WithProbeDefaultTimeout sets the default timeout duration for probes.
// The timeout must be at least 1 second. If a duration less than
// 1 second is provided, it is rounded up to 1 second.
// If not set, the default timeout is 5 seconds.
//
// This value is used if a probe does not specify its own timeout. See
// Checker.AddProbe for more details.
func WithProbeDefaultTimeout(d time.Duration) CheckerOption {
	return func(o *checkerOptions) error {
		if d < time.Second {
			d = time.Second
		}

		o.probeDefaultTimeout = d

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

// WithBufferSize sets the buffer size for watcher channels.
// The buffer size must be at least 1. If a value less than 1
// is provided, it defaults to 1.
//
// This value determines how many status updates can be queued
// for watchers before starting to drop updates. A slow watcher
// may miss some updates if it cannot keep up with the health
// checking frequency.
func WithBufferSize(size int) CheckerOption {
	return func(o *checkerOptions) error {
		if size < 1 {
			size = 1
		}

		o.bufferSize = size

		return nil
	}
}
