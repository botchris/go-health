package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/botchris/go-health"
)

// HTTPConfig defines the configuration for the HTTP health checker.
type HTTPConfig struct {
	// URL is the endpoint to be checked.
	URL string

	// Method is the HTTP method to use for the request (e.g., GET, POST).
	// Default is GET.
	Method string

	// Payload is an optional request body for methods like POST or PUT.
	Payload any

	// StatusCode is the expected HTTP status code. If the response status code
	// does not match this value, the check is considered failed. Default is 200.
	StatusCode int

	// ExpectContains is an optional string that should be present in the response body.
	// If the string is not found, the check is considered failed.
	ExpectContains string

	// Client is the HTTP client to use for making requests. If not provided,
	// the default http.Client will be used.
	Client *http.Client

	parsedPayload []byte
	uri           *url.URL
}

type httpChecker struct {
	cfg *HTTPConfig
}

// NewHTTPChecker creates a new HTTPChecker based on the provided configuration.
func NewHTTPChecker(cfg *HTTPConfig) (health.Checker, error) {
	if err := cfg.prepare(); err != nil {
		return nil, err
	}

	return &httpChecker{
		cfg: cfg,
	}, nil
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

	if resp.StatusCode != h.cfg.StatusCode {
		return fmt.Errorf("expected status code %d, got %d", h.cfg.StatusCode, resp.StatusCode)
	}

	if h.cfg.ExpectContains != "" {
		data, rErr := io.ReadAll(resp.Body)
		if rErr != nil {
			return fmt.Errorf("%w: failed to read response body", rErr)
		}

		if !strings.Contains(string(data), h.cfg.ExpectContains) {
			return fmt.Errorf("expected string %q not found in response body", h.cfg.ExpectContains)
		}
	}

	return nil
}

func (h *httpChecker) do(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, h.cfg.Method, h.cfg.uri.String(), bytes.NewReader(h.cfg.parsedPayload))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create HTTP request", err)
	}

	resp, err := h.cfg.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: HTTP request failed", err)
	}

	return resp, nil
}

func (h *HTTPConfig) parsePayload(b any) ([]byte, error) {
	if b == nil {
		return nil, nil
	}

	switch val := b.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		jb, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to marshal json body", err)
		}

		return jb, nil
	}
}

func (h *HTTPConfig) prepare() error {
	if h.URL == "" {
		return errors.New("URL cannot be empty")
	}

	uri, err := url.Parse(h.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	h.uri = uri

	if h.StatusCode == 0 {
		h.StatusCode = http.StatusOK
	}

	if h.Method == "" {
		h.Method = "GET"
	}

	if h.Client == nil {
		h.Client = http.DefaultClient
	}

	payload, err := h.parsePayload(h.Payload)
	if err != nil {
		return errors.New("invalid payload")
	}

	h.parsedPayload = payload

	return nil
}
