package strwriter_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/botchris/go-health"
	"github.com/botchris/go-health/reporters/strwriter"
)

func TestReport_WritesStatusToFile(t *testing.T) {
	var buf bytes.Buffer

	f := &fakeFile{buf: &buf}
	reporter := strwriter.New(f)

	status := health.Status{
		Errors: map[string]error{
			"db":    errors.New("connection failed"),
			"cache": errors.New("timeout"),
		},
	}

	err := reporter.Report(context.Background(), status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "cache: timeout; db: connection failed"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReport_NoErrors(t *testing.T) {
	var buf bytes.Buffer
	f := &fakeFile{buf: &buf}

	reporter := strwriter.New(f)

	status := health.Status{
		Errors: map[string]error{},
	}

	err := reporter.Report(context.Background(), status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

type fakeFile struct {
	buf *bytes.Buffer
}

func (f *fakeFile) WriteString(s string) (int, error) {
	return f.buf.WriteString(s)
}

var _ interface {
	WriteString(string) (int, error)
} = (*fakeFile)(nil)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
