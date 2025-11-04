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

type httpChecker struct {
	opts *options
}

// New creates a new HTTPChecker based on the provided configuration.
func New(url string, o ...Option) (health.Checker, error) {
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

	return &httpChecker{opts: opts}, nil
}

// Check performs the HTTP health check based on the configuration.
func (h *httpChecker) Check(ctx context.Context) error {
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
			return fmt.Errorf("%w: failed to read response body", rErr)
		}

		if !strings.Contains(string(data), h.opts.expectContains) {
			return fmt.Errorf("expected string %q not found in response body", h.opts.expectContains)
		}
	}

	return nil
}

func (h *httpChecker) do(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, h.opts.method, h.opts.url.String(), bytes.NewReader(h.opts.payload))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create HTTP request", err)
	}

	resp, err := h.opts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: HTTP request failed", err)
	}

	return resp, nil
}
