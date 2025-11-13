package health

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Status represents the result of a health check as reported
// by a Checker, containing any errors encountered indexed by Probe name.
type Status interface {
	// Append adds a new probe result to the Status.
	//
	// - probeName: is the name of the Probe as registered in the
	//   Checker instance using Checker.AddProbe method.
	// - result: is the error returned by the Probe.Check method,
	//   which may be nil on success.
	Append(probeName string, result error) Status

	// Errors returns a map of Probe names to their respective errors.
	// A nil error indicates a successful probe check.
	Errors() map[string]error

	// Flatten returns a slice of all non-nil errors from the Probe checks.
	Flatten() []error

	// AsError aggregates all errors and returns them as a single error.
	// If there are no errors, it returns nil.
	AsError() error

	// Duration returns the total time taken to perform the Probe checks
	// and calculate this Status.
	Duration() time.Duration
}

type status struct {
	errors   map[string]error
	flatten  []error
	duration time.Duration
	started  time.Time
	mu       sync.RWMutex
}

// NewStatus creates and returns a new Status instance.
// The returned object is thread-safe and can be used concurrently.
//
// Optionally, a specific start time can be provided for duration calculation.
// If no time is provided, the current time is used.
func NewStatus(now ...time.Time) Status {
	n := time.Now()
	if len(now) > 0 {
		n = now[0]
	}

	return &status{
		errors:  make(map[string]error),
		flatten: make([]error, 0),
		started: n,
	}
}

func (s *status) Append(probeName string, result error) Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errors[probeName] = result
	s.duration = time.Since(s.started)

	if result != nil {
		s.flatten = append(s.flatten, result)
	}

	return s
}

func (s *status) Errors() map[string]error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.errors
}

func (s *status) Flatten() []error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.flatten
}

func (s *status) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.duration
}

// AsError aggregates all errors in the StatusStruct and returns them
// as a single error using errors.Join. If there are no errors,
// it returns nil.
func (s *status) AsError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.flatten) == 0 {
		return nil
	}

	return fmt.Errorf("health check failed with %d errors: %w", len(s.flatten), errors.Join(s.flatten...))
}
