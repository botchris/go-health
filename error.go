package health

import (
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

type synStatus struct {
	status Status
	mu     sync.Mutex
}

func (s *synStatus) addError(checkerName string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status.Errors == nil {
		s.status.Errors = make(map[string]error)
	}

	s.status.Errors[checkerName] = err
	s.status.flatten = append(s.status.flatten, err)
}
