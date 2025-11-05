package health

// Reporter is an interface for reporting health status.
// TODO: work in progress.
type Reporter interface {
	Report(status Status) error
}
