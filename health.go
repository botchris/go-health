package health

import (
	"context"
	"sync"
	"time"
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
func (h *Health) Start(ctx context.Context) <-chan *Error {
	ticker := time.NewTicker(h.period)
	statusChan := make(chan *Error)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				close(statusChan)

				return
			case <-ticker.C:
				errMu := sync.Mutex{}
				errs := make(map[string]error)
				wg := sync.WaitGroup{}

				{
					h.mu.Lock()

					for i := range h.checkers {
						wg.Add(1)

						go func(config config) {
							defer wg.Done()

							checkCtx, cancel := context.WithTimeout(ctx, config.timeout)
							defer cancel()

							if err := config.checker.Check(checkCtx); err != nil {
								errMu.Lock()
								errs[config.name] = err
								errMu.Unlock()
							}
						}(h.checkers[i])
					}

					h.mu.Unlock()
				}

				wg.Wait()

				if len(errs) == 0 {
					continue
				}

				flattenedErrs := make([]error, 0)
				for _, err := range errs {
					flattenedErrs = append(flattenedErrs, err)
				}

				statusChan <- &Error{
					Errors:  errs,
					flatten: flattenedErrs,
				}
			}
		}
	}()

	return statusChan
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
