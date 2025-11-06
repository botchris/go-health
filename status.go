package health

import (
	"context"
	"errors"
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

// AsError aggregates all errors in the Status and returns them
// as a single error using errors.Join. If there are no errors,
// it returns nil.
func (s *Status) AsError() error {
	if len(s.flatten) == 0 {
		return nil
	}

	return errors.Join(s.flatten...)
}

type syncStatus struct {
	started time.Time
	status  Status
	mu      sync.RWMutex
}

func newSyncStatus() *syncStatus {
	return &syncStatus{
		started: time.Now(),
		status:  Status{},
	}
}

func (s *syncStatus) probe(ctx context.Context, p *probeConfig) {
	probeCtx, cancel := context.WithTimeout(ctx, p.opts.timeout)
	defer cancel()

	err := p.probe.Check(probeCtx)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status.Errors == nil {
		s.status.Errors = make(map[string]error)
	}

	s.status.Errors[p.name] = err
	s.status.flatten = append(s.status.flatten, err)
	s.status.Duration += time.Since(s.started)
}

func (s *syncStatus) read() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.status
}
