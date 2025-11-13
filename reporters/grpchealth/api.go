package grpchealth

import (
	ghealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var _ HealthServer = (*ghealth.Server)(nil)

// HealthServer abstract the part of gRPC health server we need.
type HealthServer interface {
	SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus)
}
