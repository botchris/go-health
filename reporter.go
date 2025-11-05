package health

import "context"

// Reporter is an interface for reporting health status.
// TODO: work in progress.
type Reporter interface {
	// Report is used to notify the reporter of the current health status.
	Report(ctx context.Context, status Status) error
}
