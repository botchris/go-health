package health

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Checker manages and performs periodic health checks using registered Probes.
type Checker struct {
	opts                 checkerOptions
	consecutiveSuccesses int
	consecutiveFailures  int

	probes    map[string]*probeConfig
	reporters []Reporter
	chMu      sync.RWMutex

	watchers   []chan Status
	watchersMu sync.Mutex
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
		probes:    make(map[string]*probeConfig),
		reporters: make([]Reporter, 0),
	}, nil
}

// Start initiates the health checking process in the background
// until the provided context is canceled. It returns a channel
// that emits StatusStruct objects at each checking interval.
func (ch *Checker) Start(ctx context.Context) <-chan Status {
	go ch.startChecking(ctx)
	go ch.startReporting(ctx)

	return ch.Watch()
}

// AddProbe adds a new Probe with the specified name.
// The Probe will be executed during the health checking process.
func (ch *Checker) AddProbe(name string, timeout time.Duration, probe Probe) *Checker {
	ch.chMu.Lock()
	defer ch.chMu.Unlock()

	ch.probes[name] = &probeConfig{
		name:    name,
		probe:   probe,
		timeout: timeout,
	}

	return ch
}

// AddReporter adds a new Reporter to the Checker.
func (ch *Checker) AddReporter(reporter Reporter) *Checker {
	ch.chMu.Lock()
	defer ch.chMu.Unlock()

	ch.reporters = append(ch.reporters, reporter)

	return ch
}

// Watch returns a channel that emits Status changes.
// The channel will receive updates whenever the health status changes
// based on the configured success and failure thresholds.
// Make sure to start the Checker before calling Watch.
func (ch *Checker) Watch() <-chan Status {
	watcher := make(chan Status, ch.opts.bufferSize)

	ch.watchersMu.Lock()
	ch.watchers = append(ch.watchers, watcher)
	ch.watchersMu.Unlock()

	return watcher
}

// startChecking performs health checks at regular intervals
// defined by the Checker's period. It sends the results to the provided status channel.
// The function runs until the provided context is canceled.
func (ch *Checker) startChecking(ctx context.Context) {
	ticker := time.NewTicker(ch.opts.period)
	defer ticker.Stop()

	if ch.opts.initialDelay > 0 {
		select {
		case <-ctx.Done():
			ch.closeAllWatchers()

			return
		case <-time.After(ch.opts.initialDelay):
		}
	}

	for {
		select {
		case <-ctx.Done():
			ch.closeAllWatchers()

			return
		case <-ticker.C:
			st := NewStatus()
			wg := sync.WaitGroup{}
			probes := ch.getProbes()

			for i := range probes {
				wg.Add(1)

				go func(pc *probeConfig) {
					defer wg.Done()

					probeCtx, cancel := context.WithTimeout(ctx, pc.timeout)
					defer cancel()

					st.Append(pc.name, pc.probe.Check(probeCtx))
				}(probes[i])
			}

			wg.Wait()
			ch.notifyStatus(st)
		}
	}
}

func (ch *Checker) notifyStatus(st Status) {
	shouldNotify := false
	hasFailed := st.AsError() != nil

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
		ch.watchersMu.Lock()

		for _, watcher := range ch.watchers {
			select {
			case watcher <- st:
			default:
			}
		}

		ch.watchersMu.Unlock()
	}
}

func (ch *Checker) closeAllWatchers() {
	ch.watchersMu.Lock()
	defer ch.watchersMu.Unlock()

	for _, watcher := range ch.watchers {
		close(watcher)
	}

	ch.watchers = nil
}

// startReporting listens for status updates and reports them using all registered reporters.
// It runs until the provided context is canceled or the status channel is closed.
func (ch *Checker) startReporting(ctx context.Context) {
	stream := ch.Watch()

	for {
		select {
		case <-ctx.Done():
			return
		case st, ok := <-stream:
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
						func() error { return r.Report(ctx, st) },
						backoff.WithContext(
							backoff.NewExponentialBackOff(
								backoff.WithMaxElapsedTime(ch.opts.reporterTimeout),
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

func (ch *Checker) getProbes() []*probeConfig {
	ch.chMu.RLock()
	defer ch.chMu.RUnlock()

	pbs := make([]*probeConfig, 0, len(ch.probes))
	for _, config := range ch.probes {
		pbs = append(pbs, config)
	}

	return pbs
}

func (ch *Checker) getReporters() []Reporter {
	ch.chMu.RLock()
	defer ch.chMu.RUnlock()

	reporters := make([]Reporter, len(ch.reporters))
	copy(reporters, ch.reporters)

	return reporters
}
