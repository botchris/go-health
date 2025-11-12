# Go Health

[![go test](https://github.com/botchris/go-health/actions/workflows/go-test.yml/badge.svg)](https://github.com/botchris/go-health/actions/workflows/go-test.yml)
[![golangci-lint](https://github.com/botchris/go-health/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/botchris/go-health/actions/workflows/golangci-lint.yml)

A simple Golang package for performing health checks within your Go applications.

---

## Installation

```sh
go get github.com/botchris/go-health
```

## Example Usage

```go
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/botchris/go-health"
)

func main() {
	// 1. Create a context that is canceled on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 2. Create a new health checker.
	checker, err := health.NewChecker(health.WithPeriod(time.Second))
	if err != nil {
		panic(err)
	}

	// 3. AddProbe a simple probe.
	checker.AddProbe("mysql-db01", time.Second, health.ProbeFunc(func(context.Context) error {
		time.Sleep(100 * time.Millisecond) // Simulate a database check

		return nil // return an error if the check fails
	}))

	// 4. Start the health checker and handle the results.
	for status := range checker.Start(ctx) {
		if err := status.AsError(); err != nil {
			fmt.Printf("Check failed: %s\n", err)
		} else {
			fmt.Printf("Check passed\n")
		}
	}
}

```

## Concepts

This section explains the main concepts used in the `go-health` package.

- [Probe](#probe)
- [Checker](#checker)
- [Reporter](#reporter)

### Probe

A Probe is anything that implements the `Probe` interface. The simplest way to create a probe is to use
the `ProbeFunc` type, which allows you to define a probe using a function.

Probes are expected to return an error if the check fails, or `nil` if the check passes.

#### Built-in Probes

- **DynamoDB**: A probe that checks the health of an AWS DynamoDB table.
- **gRPC**: A probe that performs a gRPC health check on a specified gRPC server.
- **HTTP**: A probe that performs an HTTP request to a specified URL and checks the response status code.
- **RabbitMQ**: A probe that checks the health of a RabbitMQ server by connecting and optionally checking a queue.
- **Redis**: A probe that pings a Redis server to check its availability, and optionally checks for a specific keys.
- **SQL**: A probe that pings a SQL database to check its availability.

#### Building Custom Probes

You can create custom probes by implementing the `Probe` interface. Here's an example of a simple custom probe:

```go
package custom

import "context"

type CustomProbe struct {
    // Add any fields you need for your probe.
}   

func (p *CustomProbe) Check(context.Context) error {
    // Implement your custom health check logic here.
    return nil // return an error if the check fails.
}

```

### Checker

A Checker is responsible for managing and executing probes at specified intervals. You can register multiple probes
with a Checker, all of which are executed concurrently for generating health status updates. If a probe fails, the
Checker will report the failure in the health status.

#### Configuration Options

The Checker can be configured using various options, such as the check period, timeout duration, and more.
These options can be set when creating a new Checker using the `NewChecker` function.

You can configure the Checker using the following functional options:

- **InitialDelay**: Sets an initial delay before the first health check is performed.  
  If the duration is between 0 and 1 second, it is rounded up to 1 second. Defaults to 0 (no delay).

- **Period**: Sets the period between consecutive health checks.  
  The minimum allowed value is 1 second; smaller values are rounded up.  
  Defaults to 10 seconds.

- **SuccessThreshold**: Sets the number of consecutive successful checks required to consider the system healthy.  
  The minimum allowed value is 1. Defaults to 1.

- **FailureThreshold**: Sets the number of consecutive failed checks required to consider the system unhealthy.  
  The minimum allowed value is 1. Defaults to 3.

- **ReporterTimeout**: Sets the timeout duration for reporter operations.  
  The minimum allowed value is 1 second; smaller values are rounded up.  
  Defaults to 30 seconds.

You can combine these options when creating a new Checker:

```go
checker, err := health.NewChecker(
    health.WithPeriod(5 * time.Second),
    health.WithInitialDelay(2 * time.Second),
    health.WithSuccessThreshold(2),
    health.WithFailureThreshold(3),
    health.WithReporterTimeout(10 * time.Second),
)
```

In the example above the Checker is configured to perform health checks every 5 seconds,
with an initial delay of 2 seconds before the first check.

It requires 2 consecutive successful checks to consider the system healthy and 3 consecutive
failed checks to consider it unhealthy. The status won't be reported until any of these thresholds are met.

### Reporter

A reporter is anything capable of reporting the status changes reported by the `Checker`. For example,
logging the status changes to the console, sending alerts, updating a dashboard, an HTTP endpoint, etc.

NOTE: reporters may not receive a status update on startup until threshold conditions
are met. So you may want to initialize your reporter with an initial "unknown" status.

#### Built-in Reporters

- **HTTP**: An HTTP reporter that exposes an endpoint for health status checks.
- **Proto Buffer**: A reporter that exposes health status service using the Health Checking Protocol defined in gRPC.
- **String Writer**: A reporter that writes health status updates to an `io.StringWriter`, such as `os.Stdout` or a log file.
