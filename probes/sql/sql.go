package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/botchris/go-health"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	_ "github.com/lib/pq"              // import pg driver
)

type sqlProbe struct {
	opts *options
}

// New creates a new SQL checker that uses the provided DSN to connect
// to the database. Supported drivers are "mysql" and "postgres".
func New(dsn string, o ...Option) (health.Probe, error) {
	opts := &options{
		dsn:    dsn,
		driver: "mysql",
		opener: sql.Open,
	}

	for i := range o {
		if err := o[i](opts); err != nil {
			return nil, err
		}
	}

	return &sqlProbe{opts: opts}, nil
}

func (m sqlProbe) Check(ctx context.Context) (result error) {
	db, err := m.opts.opener(m.opts.driver, m.opts.dsn)
	if err != nil {
		result = fmt.Errorf("%w: failed to open MySQL connection", err)

		return
	}

	defer func() {
		if err = db.Close(); err != nil && result == nil {
			result = fmt.Errorf("%w: failed to close MySQL connection", err)
		}
	}()

	if err = db.PingContext(ctx); err != nil {
		result = fmt.Errorf("%w: failed to ping MySQL database", err)

		return
	}

	rows, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		result = fmt.Errorf("%w: failed to execute version query", err)

		return
	}

	//nolint:gocritic
	defer func() {
		if err = rows.Close(); err != nil && result == nil {
			result = fmt.Errorf("%w: failed to close rows", err)
		}
	}()

	return
}
