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
	// 1. Create a context that is cancelled on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 2. Create a new health checker.
	hc := health.NewChecker(time.Second)

	// 3. Register a simple probe.
	hc.Register("database", time.Second, health.ProbeFunc(func(context.Context) error {
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
