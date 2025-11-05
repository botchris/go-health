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

	h := health.NewChecker(50 * time.Millisecond)
	successChecker := health.ProbeFunc(func(ctx context.Context) error { return nil })

	h.Register("success", 50*time.Millisecond, successChecker)

	for st := range h.Start(ctx) {
		require.NoError(t, st.AsError())
	}
}

func TestHealth_RegisterAndStart_Failure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h := health.NewChecker(50 * time.Millisecond)
	failChecker := health.ProbeFunc(func(ctx context.Context) error { return errors.New("fail") })

	h.Register("fail", 50*time.Millisecond, failChecker)

	for st := range h.Start(ctx) {
		require.Error(t, st.AsError())
	}
}

func TestHealth_MultipleCheckers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var sentinel = errors.New("sentinel")

	h := health.NewChecker(50 * time.Millisecond)
	successChecker := health.ProbeFunc(func(ctx context.Context) error { return nil })
	failChecker := health.ProbeFunc(func(ctx context.Context) error { return sentinel })

	h.Register("success", 50*time.Millisecond, successChecker)
	h.Register("fail", 50*time.Millisecond, failChecker)

	for st := range h.Start(ctx) {
		require.ErrorIs(t, st.AsError(), sentinel)
	}
}
