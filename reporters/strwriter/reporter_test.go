package strwriter_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/botchris/go-health"
	"github.com/botchris/go-health/reporters/strwriter"
	"github.com/stretchr/testify/assert"
)

func TestReport_WritesStatusToFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var buf bytes.Buffer

	f := &fakeFile{buf: &buf}
	reporter := strwriter.New(f)

	status := health.Status{
		Errors: map[string]error{
			"db":    errors.New("connection failed"),
			"cache": errors.New("timeout"),
		},
	}

	err := reporter.Report(ctx, status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "cache: timeout; db: connection failed"
	assert.Equal(t, want, got)
}

func TestReport_NoErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var buf bytes.Buffer
	f := &fakeFile{buf: &buf}

	reporter := strwriter.New(f)
	status := health.Status{
		Errors: map[string]error{
			"db":    nil,
			"cache": nil,
		},
	}

	err := reporter.Report(ctx, status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	assert.Contains(t, line, "db: ok")
	assert.Contains(t, line, "cache: ok")
}

var _ io.StringWriter = (*fakeFile)(nil)

type fakeFile struct {
	buf *bytes.Buffer
}

func (f *fakeFile) WriteString(s string) (int, error) {
	return f.buf.WriteString(s)
}
