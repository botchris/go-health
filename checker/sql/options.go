package sql

import (
	"database/sql"
	"fmt"
	"slices"
)

type Option func(*options) error

type options struct {
	driver string
	dsn    string
	opener func(driver, dsn string) (*sql.DB, error)
}

// WithDriver sets the SQL driver to use (e.g., "mysql" or "postgres").
func WithDriver(driver string) Option {
	return func(o *options) error {
		if !slices.Contains([]string{"mysql", "postgres"}, driver) {
			return fmt.Errorf("unsupported driver: %s", driver)
		}

		o.driver = driver

		return nil
	}
}

func WithOpener(opener func(driver, dsn string) (*sql.DB, error)) Option {
	return func(o *options) error {
		if opener == nil {
			return fmt.Errorf("opener function cannot be nil")
		}

		o.opener = opener

		return nil
	}
}
