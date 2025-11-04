package health

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Health manages and performs periodic health checks using registered checkers.
type Health struct {
	mu       sync.Mutex
	checkers map[string]config
	period   time.Duration
}

type config struct {
	name    string
	timeout time.Duration
	checker Checker
}

// New creates a new Health instance with the specified checking period.
// The period defines how often health checks are performed. And it must be
// at least 1 second.
func New(period time.Duration) *Health {
	if period < time.Second {
		period = time.Second
	}

	return &Health{
		checkers: make(map[string]config),
		period:   period,
	}
}

// Start initiates the health checking process until the provided
// context is canceled.
func (h *Health) Start(ctx context.Context) <-chan error {
	ticker := time.NewTicker(h.period)
	errCh := make(chan error)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				close(errCh)

				return
			case <-ticker.C:
				h.mu.Lock()

				errGroup := multierror.Group{}

				for i := range h.checkers {
					c := h.checkers[i]

					errGroup.Go(func() error {
						checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
						defer cancel()

						return c.checker.Check(checkCtx)
					})
				}

				h.mu.Unlock()

				e := errGroup.Wait()

				errCh <- e.ErrorOrNil()
			}
		}
	}()

	return errCh
}

// Register adds a new health checker with the specified name and timeout.
// The checker will be executed during the health checking process.
func (h *Health) Register(name string, timeout time.Duration, checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checkers[name] = config{
		name:    name,
		timeout: timeout,
		checker: checker,
	}
}
