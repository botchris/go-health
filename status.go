package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Status represents the result of health checks,
// containing any errors encountered indexed by checker name.
type Status struct {
	// Errors maps checker names to their respective errors.
	// If a checker passed, its error will be nil.
	Errors map[string]error

	// Flatten is a slice of all non-nil errors from the health checks.
	Flatten []error

	// Duration indicates the total time taken to perform the health checks.
	Duration time.Duration
}

// AsError aggregates all errors in the Status and returns them
// as a single error using errors.Join. If there are no errors,
// it returns nil.
func (s Status) AsError() error {
	if len(s.Flatten) == 0 {
		return nil
	}

	combined := errors.Join(s.Flatten...)

	return fmt.Errorf("health check failed with %d errors: %w", len(s.Errors), combined)
}

type syncStatus struct {
	started time.Time
	status  Status
	mu      sync.RWMutex
}

func newSyncStatus() *syncStatus {
	return &syncStatus{
		started: time.Now(),
		status: Status{
			Errors:  make(map[string]error),
			Flatten: make([]error, 0),
		},
	}
}

func (s *syncStatus) probe(ctx context.Context, pc *probeConfig) {
	probeCtx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()

	err := pc.probe.Check(probeCtx)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Errors[pc.name] = err
	s.status.Duration = time.Since(s.started)

	if err != nil {
		s.status.Flatten = append(s.status.Flatten, err)
	}
}

func (s *syncStatus) read() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.status
}
