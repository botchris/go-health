package checker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/botchris/go-health/checker"
	"github.com/stretchr/testify/require"
)

func TestHTTPChecker_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, wErr := w.Write([]byte(`{"status":"ok"}`)); wErr != nil {
			http.Error(w, wErr.Error(), http.StatusInternalServerError)
		}
	}))

	defer srv.Close()

	cfg := &checker.HTTPConfig{
		URL:            srv.URL,
		Method:         "GET",
		StatusCode:     http.StatusOK,
		ExpectContains: `{"status":"ok"}`,
		Client:         srv.Client(),
	}

	c, err := checker.NewHTTPChecker(cfg)
	require.NoError(t, err)

	err = c.Check(ctx)
	require.NoError(t, err)
}

func TestHTTPChecker_StatusCodeMismatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	defer srv.Close()

	cfg := &checker.HTTPConfig{
		URL:        srv.URL,
		Method:     "GET",
		StatusCode: http.StatusOK,
		Client:     srv.Client(),
	}

	c, err := checker.NewHTTPChecker(cfg)
	require.NoError(t, err)

	err = c.Check(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected status code")
}

func TestHTTPChecker_ExpectStringNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, wErr := w.Write([]byte(`{"status":"not-ok"}`)); wErr != nil {
			http.Error(w, wErr.Error(), http.StatusInternalServerError)
		}
	}))

	defer srv.Close()

	cfg := &checker.HTTPConfig{
		URL:            srv.URL,
		Method:         "GET",
		StatusCode:     http.StatusOK,
		ExpectContains: "missing-string",
		Client:         srv.Client(),
	}

	c, err := checker.NewHTTPChecker(cfg)
	require.NoError(t, err)

	err = c.Check(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected string")
}
