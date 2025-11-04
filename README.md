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
	"time"

	"github.com/botchris/go-health"
)

func main() {
	// Create a new health checker
	hc := health.New(time.Second)

	// Register a simple health check
	hc.Register("database", time.Second, health.CheckFunc(func() error {
		// Simulate a database check
		time.Sleep(100 * time.Millisecond)

		return nil // return an error if the check fails
	}))

	// Perform health checks
	for err := range hc.Start(context.TODO()) {
		if err != nil {
			fmt.Printf("Check failed: %s\n",  err)
		} else {
			fmt.Printf("Check passed\n")
		}
	}
}
```
