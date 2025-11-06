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
	checker.AddProbe("mysql-db01", health.ProbeFunc(func(context.Context) error {
		time.Sleep(100 * time.Millisecond) // Simulate a database check

		return nil // return an error if the check fails
	}), health.WithProbeTimeout(time.Second))

	// 4. Start the health checker and handle the results.
	for status := range checker.Start(ctx) {
		if err := status.AsError(); err != nil {
			fmt.Printf("Check failed: %s\n", err)
		} else {
			fmt.Printf("Check passed\n")
		}
	}
}
