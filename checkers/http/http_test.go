package http_test

import (
	"context"
	sdkhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/botchris/go-health/checkers/http"
	"github.com/stretchr/testify/require"
)

func TestHTTPChecker_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(sdkhttp.HandlerFunc(func(w sdkhttp.ResponseWriter, r *sdkhttp.Request) {
		w.WriteHeader(sdkhttp.StatusOK)

		if _, wErr := w.Write([]byte(`{"status":"ok"}`)); wErr != nil {
			sdkhttp.Error(w, wErr.Error(), sdkhttp.StatusInternalServerError)
		}
	}))

	defer srv.Close()

	c, err := http.New(
		srv.URL,
		http.WithExpectContains(`{"status":"ok"}`),
		http.WithClient(srv.Client()),
	)
	require.NoError(t, err)

	err = c.Check(ctx)
	require.NoError(t, err)
}

func TestHTTPChecker_StatusCodeMismatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(sdkhttp.HandlerFunc(func(w sdkhttp.ResponseWriter, r *sdkhttp.Request) {
		w.WriteHeader(sdkhttp.StatusNotFound)
	}))

	defer srv.Close()

	c, err := http.New(srv.URL, http.WithClient(srv.Client()))
	require.NoError(t, err)

	err = c.Check(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected status code")
}

func TestHTTPChecker_ExpectStringNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv := httptest.NewServer(sdkhttp.HandlerFunc(func(w sdkhttp.ResponseWriter, r *sdkhttp.Request) {
		w.WriteHeader(sdkhttp.StatusOK)

		if _, wErr := w.Write([]byte(`{"status":"not-ok"}`)); wErr != nil {
			sdkhttp.Error(w, wErr.Error(), sdkhttp.StatusInternalServerError)
		}
	}))

	defer srv.Close()

	c, err := http.New(
		srv.URL,
		http.WithExpectContains("missing-string"),
		http.WithClient(srv.Client()),
	)
	require.NoError(t, err)

	err = c.Check(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected string")
}
