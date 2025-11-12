package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/botchris/go-health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type grpcProbe struct {
	opts *options
}

// New creates a new Probe that performs a gRPC health check
// against the specified address. This probe attempts to establish
// a gRPC connection and optionally performs a health check using the
// gRPC Health Checking Protocol.
func New(addr string, o ...Option) (health.Probe, error) {
	if err := validateAddr(addr); err != nil {
		return nil, err
	}

	opts := &options{
		addr: addr,
	}

	for i := range o {
		if oErr := o[i](opts); oErr != nil {
			return nil, oErr
		}
	}

	return grpcProbe{opts: opts}, nil
}

func (g grpcProbe) Check(ctx context.Context) (checkErr error) {
	conn, err := grpc.NewClient(g.opts.addr, g.opts.dialOptions...)
	if err != nil {
		return fmt.Errorf("failed to establish gRPC connection: %w", err)
	}

	defer func() {
		if cErr := conn.Close(); cErr != nil && checkErr == nil {
			checkErr = fmt.Errorf("failed to close gRPC connection: %w", cErr)
		}
	}()

	conn.Connect()

	if !conn.WaitForStateChange(ctx, connectivity.Ready) {
		checkErr = fmt.Errorf("failed to connect to gRPC server: timeout exceeded")

		return
	}

	if g.opts.serviceName == "" {
		g.opts.serviceName = ""
	}

	healthClient := grpc_health_v1.NewHealthClient(conn)

	res, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: g.opts.serviceName})
	if err != nil {
		checkErr = fmt.Errorf("gRPC health check failed: %w", err)

		return
	}

	if res.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		checkErr = fmt.Errorf("gRPC service is not healthy: %s", res.GetStatus().String())

		return
	}

	return nil
}

func validateAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Validate host:port format
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format, expected host:port: %w", err)
	}

	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if strings.TrimSpace(port) == "" {
		return fmt.Errorf("port cannot be empty")
	}

	return nil
}
