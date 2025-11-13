package strwriter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/botchris/go-health"
)

type writer struct {
	f io.StringWriter
}

// New creates a new string writer reporter which writes health
// status to the provided io.StringWriter. For example, os.Stdout can be used
// to print health status to the console.
func New(f io.StringWriter) health.Reporter {
	return &writer{f: f}
}

func (i writer) Report(_ context.Context, status health.Status) error {
	_, err := i.f.WriteString(i.statusToLogLine(status))

	return err
}

func (i writer) statusToLogLine(status health.Status) string {
	out := make(map[string]string)

	for probe, err := range status.Errors() {
		right := "ok"
		if err != nil {
			right = err.Error()
		}

		out[probe] = right
	}

	jsonOut, jErr := json.Marshal(out)
	if jErr == nil {
		return string(jsonOut)
	}

	return fmt.Sprintf("failed to marshal health status to JSON: %v; raw status: %v", jErr, status.Errors())
}
