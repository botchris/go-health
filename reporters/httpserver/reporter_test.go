package httpserver_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/botchris/go-health"
	"github.com/botchris/go-health/reporters/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFreePort(t *testing.T) string {
	t.Helper()

	lc := net.ListenConfig{}
	l, lErr := lc.Listen(context.TODO(), "tcp", "localhost:0")
	require.NoError(t, lErr)

	defer func() {
		if err := l.Close(); err != nil {
			t.Fatalf("failed to close listener: %v", err)
		}
	}()

	return l.Addr().String()
}

func TestHTTPReporter_HealthyStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := getFreePort(t)
	reporter := httpserver.New(ctx, httpserver.WithAddr(addr))

	time.Sleep(100 * time.Millisecond)

	status := (health.Status{}).Append("db", nil).Append("cache", nil)
	require.NoError(t, reporter.Report(ctx, status))

	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr+"/healthz", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), `"db":"ok"`)
	require.Contains(t, string(body), `"cache":"ok"`)
}

func TestHTTPReporter_UnhealthyStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := getFreePort(t)
	reporter := httpserver.New(ctx, httpserver.WithAddr(addr))

	time.Sleep(100 * time.Millisecond)

	status := (health.Status{}).
		Append("db", errors.New("connection refused")).
		Append("cache", errors.New("connection timeout"))

	require.NoError(t, reporter.Report(ctx, status))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/healthz", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), `"db":"connection refused"`)
	require.Contains(t, string(body), `"cache":"connection timeout"`)
}

func TestHTTPReporter_CustomPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := getFreePort(t)
	path := "/custom"
	reporter := httpserver.New(ctx, httpserver.WithAddr(addr), httpserver.WithPath(path))

	time.Sleep(100 * time.Millisecond)

	status := health.Status{}
	require.NoError(t, reporter.Report(ctx, status))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+path, nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}
