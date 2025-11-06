package health

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Checker manages and performs periodic health checks using registered checkers.
type Checker struct {
	opts      checkerOptions
	checkers  map[string]*probeConfig
	reporters []Reporter
	mu        sync.RWMutex
}

type probeConfig struct {
	name  string
	probe Probe
	opts  probeOptions
}

// NewChecker creates a new Checker instance with the specified checking period.
// The period defines how often health checks are performed. And it must be
// at least 1 second.
func NewChecker(o ...CheckerOption) (*Checker, error) {
	opts := defaultCheckerOptions

	for i := range o {
		if err := o[i](&opts); err != nil {
			return nil, err
		}
	}

	return &Checker{
		opts:      opts,
		checkers:  make(map[string]*probeConfig),
		reporters: make([]Reporter, 0),
	}, nil
}

// Start initiates the health checking process in the background
// until the provided context is canceled. It returns a channel
// that emits Summary objects at each checking interval.
func (h *Checker) Start(ctx context.Context) <-chan Status {
	statusChan := make(chan Status)

	go h.startChecking(ctx, statusChan)
	go h.startReporting(ctx, statusChan)

	return statusChan
}

// AddProbe adds a new Probe with the specified name.
// The Probe will be executed during the health checking process.
func (h *Checker) AddProbe(name string, checker Probe, o ...ProbeOption) *Checker {
	h.mu.Lock()
	defer h.mu.Unlock()

	opts := defaultProbeOptions

	for i := range o {
		o[i](&opts)
	}

	h.checkers[name] = &probeConfig{
		name:  name,
		probe: checker,
		opts:  opts,
	}

	return h
}

// AddReporter adds a new Reporter to the Checker.
func (h *Checker) AddReporter(reporter Reporter) *Checker {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.reporters = append(h.reporters, reporter)

	return h
}

// startChecking performs health checks at regular intervals
// defined by the Checker's period. It sends the results to the provided status channel.
// The function runs until the provided context is canceled.
func (h *Checker) startChecking(ctx context.Context, statusChan chan<- Status) {
	ticker := time.NewTicker(h.opts.period)
	defer ticker.Stop()

	if h.opts.initialDelay > 0 {
		select {
		case <-ctx.Done():
			close(statusChan)

			return
		case <-time.After(h.opts.initialDelay):
		}
	}

	for {
		select {
		case <-ctx.Done():
			close(statusChan)

			return
		case <-ticker.C:
			st := newSyncStatus()
			wg := sync.WaitGroup{}
			checkers := h.getCheckers()

			for i := range checkers {
				wg.Add(1)

				go func(config *probeConfig) {
					defer wg.Done()

					st.probe(ctx, config)
				}(checkers[i])
			}

			wg.Wait()

			statusChan <- st.read()
		}
	}
}

// startReporting listens for status updates and reports them using all registered reporters.
// It runs until the provided context is canceled or the status channel is closed.
func (h *Checker) startReporting(ctx context.Context, statusChan <-chan Status) {
	for {
		select {
		case <-ctx.Done():
			return
		case status, ok := <-statusChan:
			if !ok {
				return
			}

			reporters := h.getReporters()
			wg := sync.WaitGroup{}

			for _, reporter := range reporters {
				wg.Add(1)

				go func(r Reporter) {
					defer wg.Done()

					rErr := backoff.Retry(
						func() error { return r.Report(ctx, status) },
						backoff.WithContext(
							backoff.NewExponentialBackOff(
								backoff.WithRetryStopDuration(30*time.Second),
							),
							ctx,
						),
					)

					if rErr != nil {
						log.Printf("health checker: reporter %T failed to report status: %s", r, rErr)
					}
				}(reporter)
			}

			wg.Wait()
		}
	}
}

func (h *Checker) getCheckers() []*probeConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()

	checkers := make([]*probeConfig, 0, len(h.checkers))
	for _, config := range h.checkers {
		checkers = append(checkers, config)
	}

	return checkers
}

func (h *Checker) getReporters() []Reporter {
	h.mu.RLock()
	defer h.mu.RUnlock()

	reporters := make([]Reporter, len(h.reporters))
	copy(reporters, h.reporters)

	return reporters
}
