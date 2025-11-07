package protohealth

import (
	"context"

	"github.com/botchris/go-health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// HealthServer abstract the part of gRPC health server we need.
type HealthServer interface {
	SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus)
}

type proto struct {
	service string
	server  HealthServer
}

// New creates a new reporter that reports health status to the gRPC health server
// under the given service name.
func New(serviceName string, server HealthServer) health.Reporter {
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
