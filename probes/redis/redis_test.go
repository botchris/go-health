package redis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/botchris/go-health/probes/redis"
	"github.com/stretchr/testify/require"
)

func TestRedis_Success(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	checker, err := redis.New(dsn)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.NoError(t, err)
}

func TestRedis_InvalidDSN(t *testing.T) {
	_, err := redis.New("invalid-dsn")
	require.Error(t, err)
}

func TestRedis_InvalidServer(t *testing.T) {
	dsn := "redis://localhost:63999"
	checker, err := redis.New(dsn)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.Error(t, err)
}

func TestRedis_SetCheck_Success(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	setOpt := redis.WithSetChecker(redis.SetCheck{
		Key:        "health:set",
		Value:      "ok",
		Expiration: 0,
	})
	checker, err := redis.New(dsn, setOpt)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.NoError(t, err)

	found, gErr := srv.Get("health:set")
	require.NoError(t, gErr)
	require.Equal(t, "ok", found)
}

func TestRedis_GetCheck_Success(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)

	defer srv.Close()

	require.NoError(t, srv.Set("health:get", "expected"))

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	getOpt := redis.WithGetChecker(redis.GetCheck{
		Key:           "health:get",
		ExpectedValue: "expected",
	})
	checker, err := redis.New(dsn, getOpt)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.NoError(t, err)
}

func TestRedis_GetCheck_MissingKey_Error(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	getOpt := redis.WithGetChecker(redis.GetCheck{
		Key: "missing:key",
	})
	checker, err := redis.New(dsn, getOpt)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestRedis_GetCheck_MissingKey_NoError(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)

	defer srv.Close()

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	getOpt := redis.WithGetChecker(redis.GetCheck{
		Key:                 "missing:key",
		NoErrorOnMissingKey: true,
	})
	checker, err := redis.New(dsn, getOpt)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.NoError(t, err)
}

func TestRedis_GetCheck_UnexpectedValue(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	require.NoError(t, srv.Set("health:get", "actual"))

	dsn := fmt.Sprintf("redis://%s", srv.Addr())
	getOpt := redis.WithGetChecker(redis.GetCheck{
		Key:           "health:get",
		ExpectedValue: "expected",
	})
	checker, err := redis.New(dsn, getOpt)
	require.NoError(t, err)

	err = checker.Check(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected value")
}
