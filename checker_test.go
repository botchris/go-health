package health_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/botchris/go-health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_RegisterAndStart_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	successChecker := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("success", 50*time.Millisecond, successChecker)

	for st := range checker.Start(ctx) {
		require.NoError(t, st.AsError())
		require.NotZero(t, st.Duration)
	}
}

func TestHealth_RegisterAndStart_Failure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	failChecker := health.ProbeFunc(func(ctx context.Context) error { return errors.New("fail") })
	checker.AddProbe("fail", 50*time.Millisecond, failChecker)

	for st := range checker.Start(ctx) {
		require.Error(t, st.AsError())
	}
}

func TestHealth_MultipleCheckers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var sentinel = errors.New("sentinel")

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	successChecker := health.ProbeFunc(func(ctx context.Context) error { return nil })
	failChecker := health.ProbeFunc(func(ctx context.Context) error { return sentinel })

	checker.AddProbe("success", 50*time.Millisecond, successChecker)
	checker.AddProbe("fail", 50*time.Millisecond, failChecker)

	for st := range checker.Start(ctx) {
		require.ErrorIs(t, st.AsError(), sentinel)
	}
}

func TestHealth_WithInitialDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	delay := 500 * time.Millisecond
	checker, err := health.NewChecker(
		health.WithInitialDelay(delay),
		health.WithPeriod(1*time.Second),
	)
	require.NoError(t, err)

	probe := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("delayed", 1*time.Second, probe)

	start := time.Now()
	statusCh := checker.Start(ctx)

	st, ok := <-statusCh
	require.True(t, ok)

	elapsed := time.Since(start)
	require.GreaterOrEqual(t, elapsed, delay)
	require.NoError(t, st.AsError())
}

func TestHealth_WithSuccessThreshold(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checker, err := health.NewChecker(
		health.WithPeriod(10*time.Millisecond),
		health.WithSuccessThreshold(3),
	)
	require.NoError(t, err)

	probeCalls := atomic.Int64{}
	successChecker := health.ProbeFunc(func(ctx context.Context) error {
		probeCalls.Add(1)

		return nil
	})

	checker.AddProbe("success", 50*time.Millisecond, successChecker)

	statusCh := checker.Start(ctx)
	seenStatuses := 0

	for st := range statusCh {
		require.NoError(t, st.AsError())

		seenStatuses++

		break
	}

	assert.EqualValues(t, 3, probeCalls.Load())
	assert.EqualValues(t, 1, seenStatuses)
}

func TestHealth_WithFailureThreshold(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checker, err := health.NewChecker(
		health.WithPeriod(50*time.Millisecond),
		health.WithFailureThreshold(2),
	)
	require.NoError(t, err)

	probeCalls := atomic.Int64{}
	failChecker := health.ProbeFunc(func(ctx context.Context) error {
		probeCalls.Add(1)

		return errors.New("fail")
	})

	checker.AddProbe("fail", 50*time.Millisecond, failChecker)

	statusCh := checker.Start(ctx)
	seenStatuses := 0

	for st := range statusCh {
		require.Error(t, st.AsError())

		seenStatuses++

		break
	}

	assert.EqualValues(t, 2, probeCalls.Load())
	assert.Equal(t, 1, seenStatuses)
}

func TestHealth_Watch_EmitsStatusChanges(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	probe := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("watch", 50*time.Millisecond, probe)

	watchCh := checker.Watch()

	checker.Start(ctx)

	select {
	case st, ok := <-watchCh:
		require.True(t, ok)
		require.NoError(t, st.AsError())
	case <-ctx.Done():
		t.Fatal("did not receive status update from Watch channel")
	}
}

func TestHealth_Watch_ClosedOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	probe := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("watch-cancel", 50*time.Millisecond, probe)

	watchCh := checker.Watch()
	checker.Start(ctx)

	<-ctx.Done()

	_, ok := <-watchCh
	assert.False(t, ok, "Watch channel should be closed after context cancel")
}

func TestHealth_MultipleWatchers_ReceiveSameStatuses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	probe := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("multi-watch", 50*time.Millisecond, probe)

	watchCh1 := checker.Watch()
	watchCh2 := checker.Watch()

	checker.Start(ctx)

	var (
		st1, st2 health.Status
		ok1, ok2 bool
	)

	select {
	case st1, ok1 = <-watchCh1:
	case <-ctx.Done():
		t.Fatal("watchCh1 did not receive status")
	}

	select {
	case st2, ok2 = <-watchCh2:
	case <-ctx.Done():
		t.Fatal("watchCh2 did not receive status")
	}

	require.True(t, ok1)
	require.True(t, ok2)
	assert.Equal(t, st1, st2)
	assert.NoError(t, st1.AsError())
}

func TestHealth_Reporter_ReceivesStatusUpdates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	probe := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("reporter-probe", 50*time.Millisecond, probe)

	mock := &mockReporter{}
	checker.AddReporter(mock)
	checker.Start(ctx)

	require.Eventually(t, func() bool {
		return mock.calls.Load() > 0
	}, 10*time.Second, 50*time.Millisecond, "reporter did not receive status update")

	assert.NoError(t, mock.getLast().AsError())
}

type mockReporter struct {
	calls atomic.Int64

	last health.Status
	lMu  sync.RWMutex
}

func (m *mockReporter) Report(_ context.Context, st health.Status) error {
	m.calls.Add(1)

	m.lMu.Lock()
	m.last = st
	m.lMu.Unlock()

	return nil
}

func (m *mockReporter) getLast() health.Status {
	m.lMu.RLock()
	defer m.lMu.RUnlock()

	return m.last
}
