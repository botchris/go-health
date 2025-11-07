package strwriter

import (
	"context"
	"io"
	"slices"
	"strings"

	"github.com/botchris/go-health"
)

type writer struct {
	f io.StringWriter
}

// New creates a new string writer reporter which writes health status to the provided io.StringWriter.
func New(f io.StringWriter) health.Reporter {
	return &writer{f: f}
}

func (i writer) Report(_ context.Context, status health.Status) error {
	_, err := i.f.WriteString(i.statusToLogLine(status) + "\n")

	return err
}

func (i writer) statusToLogLine(status health.Status) string {
	out := make([]string, 0)

	for check, err := range status.Errors {
		out = append(out, check+": "+err.Error())
	}

	slices.Sort(out)

	return strings.Join(out, "; ")
}
