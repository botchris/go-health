package grpchealth

import (
	"context"

	"github.com/botchris/go-health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type proto struct {
	opts *options
}

// New creates a new reporter that reports health status to the
// gRPC health server. Use the option WithServiceName to specify
// the service name to report the status for. If no service name
// is provided, the status will be reported for the overall server.
func New(server HealthServer, o ...Option) health.Reporter {
	opts := &options{server: server}

	for i := range o {
		o[i](opts)
	}

	return &proto{opts: opts}
}

func (p proto) Report(_ context.Context, status health.Status) error {
	pbStatus := healthpb.HealthCheckResponse_NOT_SERVING
	if err := status.AsError(); err == nil {
		pbStatus = healthpb.HealthCheckResponse_SERVING
	}

	if len(p.opts.serviceNames) == 0 {
		p.opts.server.SetServingStatus("", pbStatus)

		return nil
	}

	for i := range p.opts.serviceNames {
		p.opts.server.SetServingStatus(p.opts.serviceNames[i], pbStatus)
	}

	return nil
}
