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
	Errors   map[string]error
	Duration time.Duration
	flatten  []error
}

// Append adds a new error for the given probe name to the Status.
func (s Status) Append(probeName string, err error) Status {
	if s.Errors == nil {
		s.Errors = make(map[string]error)
	}

	s.Errors[probeName] = err

	if err != nil {
		s.flatten = append(s.flatten, err)
	}

	return s
}

// AsError aggregates all errors in the Status and returns them
// as a single error using errors.Join. If there are no errors,
// it returns nil.
func (s Status) AsError() error {
	if len(s.flatten) == 0 {
		return nil
	}

	combined := errors.Join(s.flatten...)

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
			Errors: make(map[string]error),
		},
	}
}

func (s *syncStatus) probe(ctx context.Context, pc *probeConfig) {
	probeCtx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()

	err := pc.probe.Check(probeCtx)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.Duration = time.Since(s.started)

	s.status.Append(pc.name, err)
}

func (s *syncStatus) read() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.status
}
