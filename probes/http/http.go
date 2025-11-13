package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/botchris/go-health"
)

type httpProbe struct {
	opts *options
}

// New creates a new HTTP Probe based on the provided configuration.
func New(url string, o ...Option) (health.Probe, error) {
	opts := &options{
		method:     http.MethodGet,
		statusCode: http.StatusOK,
		client:     http.DefaultClient,
	}

	o = append(o, withURL(url))

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, err
		}
	}

	return &httpProbe{opts: opts}, nil
}

// Check performs the HTTP health check based on the configuration.
func (h *httpProbe) Check(ctx context.Context) error {
	resp, err := h.do(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != h.opts.statusCode {
		return fmt.Errorf("expected status code %d, got %d", h.opts.statusCode, resp.StatusCode)
	}

	if h.opts.expectContains != "" {
		data, rErr := io.ReadAll(resp.Body)
		if rErr != nil {
			return fmt.Errorf("failed to read response body: %w", rErr)
		}

		if !strings.Contains(string(data), h.opts.expectContains) {
			return fmt.Errorf("expected string %q not found in response body", h.opts.expectContains)
		}
	}

	return nil
}

func (h *httpProbe) do(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, h.opts.method, h.opts.url.String(), bytes.NewReader(h.opts.payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := h.opts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}
