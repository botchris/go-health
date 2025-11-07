package httpserver

import "strings"

// Option is a functional option for httpReporter.
type Option func(*httpReporter)

// WithAddr sets the bind address for the HTTP server.
func WithAddr(addr string) Option {
	return func(r *httpReporter) {
		if !strings.Contains(addr, ":") {
			addr = ":" + addr
		}

		r.addr = addr
	}
}

// WithPath sets the HTTP path for health checks.
func WithPath(path string) Option {
	return func(r *httpReporter) {
		strings.TrimPrefix(path, "/")

		r.path = "/" + path
	}
}
