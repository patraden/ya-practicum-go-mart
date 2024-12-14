package database_test

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
)

func TestDatabaseInitAndPingSuccess(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	mockPool.ExpectPing()

	db := database.NewDatabase(dsn, log)
	defer db.Close()

	err = db.Init(ctx)
	require.NoError(t, err)

	db = db.WithPool(mockPool)
	err = db.Ping(ctx)
	require.NoError(t, err)
}

func TestDatabaseInitFailureBadDSN(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	dsn := "bad_dsn"
	ctx := context.Background()

	db := database.NewDatabase(dsn, log)
	defer db.Close()

	err := db.Init(ctx)
	require.Error(t, err)

	err = db.Ping(ctx)
	require.Error(t, err)
}

func TestDatabasePingFailure(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	mockPool, err := pgxmock.NewPool()
	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	ctx := context.Background()

	require.NoError(t, err)

	mockPool.ExpectPing().WillReturnError(e.ErrTesting)

	db := database.NewDatabase(dsn, log)
	defer db.Close()

	err = db.Init(ctx)
	require.NoError(t, err)

	db = db.WithPool(mockPool)
	err = db.Ping(ctx)
	require.Error(t, err)
}

// Test replacing the connection pool using WithPool.
func TestDatabaseWithPool(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	mockPool, err := pgxmock.NewPool()
	dsn := "postgres://fake-dsn"

	require.NoError(t, err)

	db := database.NewDatabase(dsn, log)
	db = db.WithPool(mockPool)

	require.Equal(t, mockPool, db.ConnPool, "connection pool should be set correctly")
	db.Close()
	require.Nil(t, db.ConnPool, "connection pool should be nil after closing")
}

func TestDatabaseCloseWithoutInit(t *testing.T) {
	t.Parallel()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	dsn := "postgres://fake-dsn"
	db := database.NewDatabase(dsn, log)

	require.NotPanics(t, func() {
		db.Close()
	}, "calling Close on an uninitialized database should not panic")
}
