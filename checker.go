package health

import (
	"context"
	"sync"
	"time"
)

// Checker manages and performs periodic health checks using registered checkers.
type Checker struct {
	mu       sync.Mutex
	checkers map[string]probeConfig
	period   time.Duration
}

type probeConfig struct {
	name    string
	timeout time.Duration
	checker Probe
}

// NewChecker creates a new Checker instance with the specified checking period.
// The period defines how often health checks are performed. And it must be
// at least 1 second.
func NewChecker(period time.Duration) *Checker {
	if period < time.Second {
		period = time.Second
	}

	return &Checker{
		checkers: make(map[string]probeConfig),
		period:   period,
	}
}

// Start initiates the health checking process in the background
// until the provided context is canceled. It returns a channel
// that emits Status objects at each checking interval.
func (h *Checker) Start(ctx context.Context) <-chan Status {
	ticker := time.NewTicker(h.period)
	statusChan := make(chan Status)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				close(statusChan)

				return
			case <-ticker.C:
				st := &synStatus{status: Status{}}
				wg := sync.WaitGroup{}

				{
					h.mu.Lock()

					for i := range h.checkers {
						wg.Add(1)

						go func(config probeConfig) {
							defer wg.Done()

							checkCtx, cancel := context.WithTimeout(ctx, config.timeout)
							defer cancel()

							st.addError(config.name, config.checker.Check(checkCtx))
						}(h.checkers[i])
					}

					h.mu.Unlock()
				}

				wg.Wait()

				statusChan <- st.status
			}
		}
	}()

	return statusChan
}

// Register adds a new Probe with the specified name and timeout.
// The Probe will be executed during the health checking process.
func (h *Checker) Register(name string, timeout time.Duration, checker Probe) *Checker {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checkers[name] = probeConfig{
		name:    name,
		timeout: timeout,
		checker: checker,
	}

	return h
}
