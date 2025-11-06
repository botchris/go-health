package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/botchris/go-health"
	"github.com/stretchr/testify/require"
)

func TestHealth_RegisterAndStart_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	successChecker := health.ProbeFunc(func(ctx context.Context) error { return nil })
	checker.AddProbe("success", successChecker, health.WithProbeTimeout(50*time.Millisecond))

	for st := range checker.Start(ctx) {
		require.NoError(t, st.AsError())
		require.NotZero(t, st.Duration)
	}
}

func TestHealth_RegisterAndStart_Failure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checker, err := health.NewChecker(health.WithPeriod(50 * time.Millisecond))
	require.NoError(t, err)

	failChecker := health.ProbeFunc(func(ctx context.Context) error { return errors.New("fail") })
	checker.AddProbe("fail", failChecker, health.WithProbeTimeout(50*time.Millisecond))

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

	checker.AddProbe("success", successChecker, health.WithProbeTimeout(50*time.Millisecond))
	checker.AddProbe("fail", failChecker, health.WithProbeTimeout(50*time.Millisecond))

	for st := range checker.Start(ctx) {
		require.ErrorIs(t, st.AsError(), sentinel)
	}
}
