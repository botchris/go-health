package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/botchris/go-health"
)

type httpReporter struct {
	addr   string
	path   string
	mu     sync.RWMutex
	last   health.Status
	server *http.Server
}

// New creates a new HTTP health reporter. By default, it listens on ":8081" and serves health status
// at the "/healthz" endpoint. These defaults can be overridden using functional options.
//
// The given context is used to manage the lifecycle of the HTTP server. When the context is canceled,
// the server will be gracefully shutdown.
func New(ctx context.Context, opts ...Option) health.Reporter {
	r := &httpReporter{
		addr: ":8081",
		path: "/healthz",
	}

	for _, opt := range opts {
		opt(r)
	}

	r.startServer(ctx)

	return r
}

// Report saves the last status and makes it available via HTTP.
func (r *httpReporter) Report(_ context.Context, status health.Status) error {
	r.mu.Lock()
	r.last = status
	r.mu.Unlock()

	return nil
}

func (r *httpReporter) startServer(ctx context.Context) {
	h := http.NewServeMux()
	h.HandleFunc(r.path, r.handleHealth)

	r.server = &http.Server{
		Addr:    r.addr,
		Handler: h,
	}

	go func() {
		if err := r.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("httpReporter server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("httpReporter shutdown error: %v", err)
		}
	}()
}

func (r *httpReporter) handleHealth(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	status := r.last
	r.mu.RUnlock()

	stCode := http.StatusOK

	if status.AsError() != nil {
		stCode = http.StatusInternalServerError
	}

	w.WriteHeader(stCode)
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(status)
}
