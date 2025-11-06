package protohealth

import (
	"context"

	"github.com/botchris/go-health"
	ghealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type proto struct {
	service string
	server  *ghealth.Server
}

// New creates a new reporter that reports health status to the gRPC health server
// under the given service name.
func New(serviceName string, server *ghealth.Server) health.Reporter {
	return &proto{
		service: serviceName,
		server:  server,
	}
}

func (p proto) Report(_ context.Context, status health.Status) error {
	pbStatus := healthpb.HealthCheckResponse_NOT_SERVING
	if err := status.AsError(); err == nil {
		pbStatus = healthpb.HealthCheckResponse_SERVING
	}

	p.server.SetServingStatus(p.service, pbStatus)

	return nil
}
