package sql_test

import (
	"context"
	sdksql "database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/botchris/go-health/probes/sql"
	"github.com/stretchr/testify/require"
)

func TestCheck_MySQLMockSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer func() {
		require.NoError(t, db.Close())
	}()

	mock.ExpectPing()
	mock.ExpectQuery("SELECT 1.*").
		WillReturnRows(
			sqlmock.
				NewRows([]string{"1"}).
				AddRow("8.0.0"),
		)
	mock.ExpectClose()

	checker, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = checker.Check(ctx)
	if err != nil {
		t.Errorf("expected successful check, got error: %v", err)
	}

	if eErr := mock.ExpectationsWereMet(); eErr != nil {
		t.Errorf("unmet sqlmock expectations: %v", eErr)
	}
}

func TestCheck_MySQLMockPingError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer func() {
		require.NoError(t, db.Close())
	}()

	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	checker, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cErr := checker.Check(ctx); cErr == nil {
		t.Error("expected error for ping failure, got nil")
	}
}

func TestCheck_MySQLMockQueryError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer func() {
		require.NoError(t, db.Close())
	}()

	mock.ExpectPing()
	mock.ExpectQuery("SELECT 1.*").WillReturnError(sqlmock.ErrCancelled)

	checker, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = checker.Check(ctx)
	if err == nil {
		t.Error("expected error for query failure, got nil")
	}
}
