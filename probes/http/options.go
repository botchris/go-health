package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
)

// Option defines a configuration option for the HTTP Probe.
type Option func(*options) error

type options struct {
	// url is the parsed URL for the HTTP request.
	url *url.URL

	// method is the HTTP method to use for the request (e.g., GET, POST).
	// Default is GET.
	method string

	// statusCode is the expected HTTP status code. If the response status code
	// does not match this value, the check is considered failed. Default is 200.
	statusCode int

	// payload is the request body to send with the HTTP request.
	payload []byte

	// expectContains is an optional string that should be present in the response body.
	// If the string is not found, the check is considered failed.
	expectContains string

	// client is the HTTP client to use for making requests. If not provided,
	// the default http.client will be used.
	client *http.Client
}

// withURL sets the URL for the HTTP request.
func withURL(u string) Option {
	return func(c *options) error {
		if u == "" {
			return errors.New("URL cannot be empty")
		}

		uri, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("%w: invalid URL", err)
		}

		c.url = uri

		return nil
	}
}

// WithMethod sets the HTTP method for the request.
func WithMethod(method string) Option {
	return func(c *options) error {
		if method == "" {
			return errors.New("method cannot be empty")
		}

		if !slices.Contains([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodPatch}, method) {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}

		c.method = method

		return nil
	}
}

// WithPayload sets the request body for the HTTP request.
func WithPayload(payload any) Option {
	return func(c *options) error {
		b, err := parsePayload(payload)
		if err != nil {
			return fmt.Errorf("%w: invalid payload", err)
		}

		c.payload = b

		return nil
	}
}

// WithStatusCode sets the expected HTTP status code for the response.
func WithStatusCode(code int) Option {
	return func(c *options) error {
		if code < 100 || code > 599 {
			return errors.New("invalid HTTP status code")
		}

		c.statusCode = code

		return nil
	}
}

// WithExpectContains sets the expected string that should be present in the response body.
func WithExpectContains(s string) Option {
	return func(c *options) error {
		if s == "" {
			return errors.New("expect contains string cannot be empty")
		}

		c.expectContains = s

		return nil
	}
}

// WithClient sets a custom HTTP client for making requests.
func WithClient(client *http.Client) Option {
	return func(c *options) error {
		c.client = client

		return nil
	}
}

func parsePayload(b any) ([]byte, error) {
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
