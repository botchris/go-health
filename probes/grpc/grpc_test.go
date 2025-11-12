package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	grpcProbe "github.com/botchris/go-health/probes/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestNew(t *testing.T) {
	t.Run("error with empty address", func(t *testing.T) {
		_, err := grpcProbe.New("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("error with invalid address format", func(t *testing.T) {
		_, err := grpcProbe.New("localhost")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid address format")
	})

	t.Run("error with empty host", func(t *testing.T) {
		_, err := grpcProbe.New(":50051")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "host cannot be empty")
	})

	t.Run("error with empty port", func(t *testing.T) {
		_, err := grpcProbe.New("localhost:")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port cannot be empty")
	})

	t.Run("success with default options", func(t *testing.T) {
		probe, err := grpcProbe.New("localhost:50051")
		require.NoError(t, err)
		assert.NotNil(t, probe)
	})

	t.Run("success with dial options", func(t *testing.T) {
		probe, err := grpcProbe.New(
			"localhost:50051",
			grpcProbe.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
		require.NoError(t, err)
		assert.NotNil(t, probe)
	})

	t.Run("success with service name", func(t *testing.T) {
		probe, err := grpcProbe.New(
			"localhost:50051",
			grpcProbe.WithServiceName("my-service"),
		)
		require.NoError(t, err)
		assert.NotNil(t, probe)
	})

	t.Run("error with empty dial options", func(t *testing.T) {
		_, err := grpcProbe.New("localhost:50051", grpcProbe.WithDialOptions())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one dial option must be provided")
	})

	t.Run("error with empty service name", func(t *testing.T) {
		_, err := grpcProbe.New("localhost:50051", grpcProbe.WithServiceName(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})
}

func TestGrpcProbe_Check(t *testing.T) {
	t.Run("health check returns NOT_SERVING", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "localhost:0")
		require.NoError(t, err)

		defer func() {
			_ = lis.Close()
		}()

		serviceName := "fake-service"
		grpcServer := grpc.NewServer()
		healthServer := health.NewServer()

		healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		go func() {
			t.Helper()

			require.NoError(t, grpcServer.Serve(lis))
		}()

		defer grpcServer.Stop()

		probe, err := grpcProbe.New(
			lis.Addr().String(),
			grpcProbe.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
			grpcProbe.WithServiceName(serviceName),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service is not healthy")
	})

	t.Run("health check returns SERVING", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "localhost:0")
		require.NoError(t, err)

		defer func() {
			_ = lis.Close()
		}()

		serviceName := "test-service"
		grpcServer := grpc.NewServer()
		healthServer := health.NewServer()

		healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		go func() {
			t.Helper()

			require.NoError(t, grpcServer.Serve(lis))
		}()

		defer grpcServer.Stop()

		probe, err := grpcProbe.New(
			lis.Addr().String(),
			grpcProbe.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
			grpcProbe.WithServiceName(serviceName),
		)

		require.NoError(t, err)
		require.NoError(t, probe.Check(ctx))
	})

	t.Run("successful check without service name", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "localhost:0")
		require.NoError(t, err)

		defer func() {
			_ = lis.Close()
		}()

		grpcServer := grpc.NewServer()
		healthServer := health.NewServer()

		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		go func() {
			t.Helper()

			require.NoError(t, grpcServer.Serve(lis))
		}()

		defer grpcServer.Stop()

		probe, err := grpcProbe.New(
			lis.Addr().String(),
			grpcProbe.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
		require.NoError(t, err)

		err = probe.Check(ctx)
		require.NoError(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		probe, err := grpcProbe.New(
			"localhost:99999",
			grpcProbe.WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = probe.Check(ctx)
		require.Error(t, err)
	})
}
