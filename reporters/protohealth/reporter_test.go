package protohealth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/botchris/go-health"
	"github.com/botchris/go-health/reporters/protohealth"
	"github.com/stretchr/testify/require"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type mockHealthServer struct {
	service string
	status  healthpb.HealthCheckResponse_ServingStatus
	calls   int
}

func (m *mockHealthServer) SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	m.service = service
	m.status = status
	m.calls++
}

func TestProtoHealthReporter_ServingStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mockServer := &mockHealthServer{}
	serviceName := "my-service"
	reporter := protohealth.New(serviceName, mockServer)

	healthyStatus := health.Status{
		Errors: map[string]error{
			"db":    nil,
			"cache": nil,
		},
	}

	require.NoError(t, reporter.Report(ctx, healthyStatus))
	require.Equal(t, serviceName, mockServer.service)
	require.Equal(t, healthpb.HealthCheckResponse_SERVING, mockServer.status)
	require.Equal(t, 1, mockServer.calls)

	mockServer.calls = 0
	mockServer.status = 0
	unhealthy := health.Status{
		Errors: map[string]error{
			"db":    nil,
			"cache": errors.New("fail"),
		},
		Flatten: []error{
			errors.New("fail"),
		},
	}

	require.NoError(t, reporter.Report(context.Background(), unhealthy))
	require.Equal(t, healthpb.HealthCheckResponse_NOT_SERVING, mockServer.status)
	require.Equal(t, 1, mockServer.calls)
}
