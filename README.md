# Go Health

[![go test](https://github.com/botchris/go-health/actions/workflows/go-test.yml/badge.svg)](https://github.com/botchris/go-health/actions/workflows/go-test.yml)
[![golangci-lint](https://github.com/botchris/go-health/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/botchris/go-health/actions/workflows/golangci-lint.yml)

A simple Golang package for performing health checks within your Go applications.

---

## Installation

```sh
go get github.com/botchris/go-health
```

## Concepts

### Probe

A Probe is anything that implements the `Probe` interface. The simplest way to create a probe is to use
the `ProbeFunc` type, which allows you to define a probe using a function.

Probes are expected to return an error if the check fails, or `nil` if the check passes.

### Checker

A Checker is responsible for managing and executing probes at specified intervals. You can register multiple probes
with a Checker, all of which are executed concurrently for generating health status updates. If a probe fails, the
Checker will report the failure in the health status.

### Reporter

A reporter is anything capable of reporting the status changes reported by the `Checker`. For example,
logging the status changes to the console, sending alerts, updating a dashboard, an HTTP endpoint, etc.

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
	hc := health.NewChecker(time.Second)

	// 3. AddProbe a simple probe.
	hc.AddProbe("mysql-db01", time.Second, health.ProbeFunc(func(context.Context) error {
		time.Sleep(100 * time.Millisecond) // Simulate a database check

		return nil // return an error if the check fails
	}))

	// 4. Start the health checker and handle the results.
	for status := range hc.Start(ctx) {
		if err := status.AsError(); err != nil {
			fmt.Printf("Check failed: %s\n", err)
		} else {
			fmt.Printf("Check passed\n")
		}
	}
}

```
