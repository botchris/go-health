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
	opts                 checkerOptions
	consecutiveSuccesses int
	consecutiveFailures  int

	checkers  map[string]*probeConfig
	reporters []Reporter
	mu        sync.RWMutex
}

type probeConfig struct {
	name    string
	probe   Probe
	timeout time.Duration
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
// that emits Status objects at each checking interval.
func (ch *Checker) Start(ctx context.Context) <-chan Status {
	statusChan := make(chan Status)

	go ch.startChecking(ctx, statusChan)
	go ch.startReporting(ctx, statusChan)

	return statusChan
}

// AddProbe adds a new Probe with the specified name.
// The Probe will be executed during the health checking process.
func (ch *Checker) AddProbe(name string, timeout time.Duration, checker Probe) *Checker {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.checkers[name] = &probeConfig{
		name:    name,
		probe:   checker,
		timeout: timeout,
	}

	return ch
}

// AddReporter adds a new Reporter to the Checker.
func (ch *Checker) AddReporter(reporter Reporter) *Checker {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.reporters = append(ch.reporters, reporter)

	return ch
}

// startChecking performs health checks at regular intervals
// defined by the Checker's period. It sends the results to the provided status channel.
// The function runs until the provided context is canceled.
func (ch *Checker) startChecking(ctx context.Context, statusChan chan<- Status) {
	ticker := time.NewTicker(ch.opts.period)
	defer ticker.Stop()

	if ch.opts.initialDelay > 0 {
		select {
		case <-ctx.Done():
			close(statusChan)

			return
		case <-time.After(ch.opts.initialDelay):
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
			checkers := ch.getCheckers()

			for i := range checkers {
				wg.Add(1)

				go func(config *probeConfig) {
					defer wg.Done()

					st.probe(ctx, config)
				}(checkers[i])
			}

			wg.Wait()
			ch.notifyStatus(st, statusChan)
		}
	}
}

func (ch *Checker) notifyStatus(st *syncStatus, statusChan chan<- Status) {
	result := st.read()
	shouldNotify := false
	hasFailed := result.AsError() != nil

	if hasFailed {
		ch.consecutiveFailures++
		ch.consecutiveSuccesses = 0

		if ch.consecutiveFailures >= ch.opts.failureThreshold {
			shouldNotify = true
		}
	}

	if !hasFailed {
		ch.consecutiveSuccesses++
		ch.consecutiveFailures = 0

		if ch.consecutiveSuccesses >= ch.opts.successThreshold {
			shouldNotify = true
		}
	}

	if shouldNotify {
		statusChan <- result
	}
}

// startReporting listens for status updates and reports them using all registered reporters.
// It runs until the provided context is canceled or the status channel is closed.
func (ch *Checker) startReporting(ctx context.Context, statusChan <-chan Status) {
	for {
		select {
		case <-ctx.Done():
			return
		case status, ok := <-statusChan:
			if !ok {
				return
			}

			reporters := ch.getReporters()
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

func (ch *Checker) getCheckers() []*probeConfig {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	checkers := make([]*probeConfig, 0, len(ch.checkers))
	for _, config := range ch.checkers {
		checkers = append(checkers, config)
	}

	return checkers
}

func (ch *Checker) getReporters() []Reporter {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	reporters := make([]Reporter, len(ch.reporters))
	copy(reporters, ch.reporters)

	return reporters
}
