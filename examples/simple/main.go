package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/botchris/go-health"
	"github.com/botchris/go-health/reporters/strwriter"
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

	// 3. Add a simple probe.
	checker.AddProbe("mysql-db01", time.Second, health.ProbeFunc(func(context.Context) error {
		time.Sleep(100 * time.Millisecond) // Simulate a database check

		return nil // return an error if the check fails
	}))

	// 4. Add string writer reporter to output health status to console.
	checker.AddReporter(strwriter.New(os.Stdout))

	// 5. Start the health checker and listen for status updates.
	for status := range checker.Start(ctx) {
		if statusErr := status.AsError(); statusErr != nil {
			fmt.Printf("Check failed: %s\n", statusErr)
		} else {
			fmt.Printf("Check passed\n")
		}
	}
}
