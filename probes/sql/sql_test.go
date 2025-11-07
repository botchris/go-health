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

	probe, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	require.NoError(t, err)
	require.NoError(t, probe.Check(ctx))
	require.NoError(t, mock.ExpectationsWereMet())
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

	probe, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	require.NoError(t, err)
	require.Error(t, probe.Check(ctx))
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

	probe, err := sql.New("mock-dsn",
		sql.WithOpener(func(driverName, dataSourceName string) (*sdksql.DB, error) {
			return db, nil
		}),
	)

	require.NoError(t, err)
	require.Error(t, probe.Check(ctx))
}
