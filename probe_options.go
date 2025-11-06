package health

import "time"

// ProbeOption is a function that configures probe behavior.
type ProbeOption func(*probeOptions)

type probeOptions struct {
	timeout          time.Duration
	successThreshold int
	failureThreshold int
}

var defaultProbeOptions = probeOptions{
	timeout:          5 * time.Second,
	successThreshold: 1,
	failureThreshold: 3,
}

// WithProbeTimeout sets the timeout for the Probe execution.
func WithProbeTimeout(d time.Duration) ProbeOption {
	return func(o *probeOptions) {
		if d < time.Second {
			d = time.Second
		}

		o.timeout = d
	}
}

// WithSuccessThreshold sets the number of consecutive successful
// checks required to consider the probe healthy.
// Minimum value is 1.
func WithSuccessThreshold(t int) ProbeOption {
	return func(o *probeOptions) {
		if t < 1 {
			t = 1
		}

		o.successThreshold = t
	}
}

// WithFailureThreshold sets the number of consecutive failed
// checks required to consider the probe unhealthy.
// Minimum value is 1.
func WithFailureThreshold(t int) ProbeOption {
	return func(o *probeOptions) {
		if t < 1 {
			t = 1
		}

		o.failureThreshold = t
	}
}
