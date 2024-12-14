package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
)

func TestDatabseIsRetryableError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Retryable connection error",
			err:      &pgconn.PgError{Code: pgerrcode.ConnectionException},
			expected: true,
		},
		{
			name:     "Non-retryable error",
			err:      &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			expected: false,
		},
		{
			name:     "Non-PgError",
			err:      e.ErrTesting,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := database.IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type RetryTest struct {
	name            string
	query           func(attempts *int) error
	expectedErr     error
	expectedRetries int
}

func DatabaseWithRetryTests(t *testing.T) []RetryTest {
	t.Helper()

	tests := []RetryTest{
		{
			"Successful query on first attempt",
			func(attempts *int) error {
				*attempts++

				return nil
			},
			nil,
			0,
		},
		{
			"Retryable error followed by success",
			func(attempts *int) error {
				*attempts++
				if *attempts == 1 {
					return &pgconn.PgError{Code: pgerrcode.ConnectionException}
				}

				return nil
			},
			nil,
			1,
		},
		{
			"Non-retryable error",
			func(attempts *int) error {
				*attempts++

				return e.ErrTesting
			},
			e.ErrTesting,
			0,
		},
		{
			"Unique violation error",
			func(attempts *int) error {
				*attempts++

				return &pgconn.PgError{Code: pgerrcode.UniqueViolation}
			},
			e.ErrTesting,
			0,
		},
	}

	return tests
}

func TestDatabaseWithRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()

	tests := DatabaseWithRetryTests(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts int
			boff := backoff.NewExponentialBackOff(
				backoff.WithMaxElapsedTime(time.Second),
				backoff.WithInitialInterval(time.Millisecond),
			)

			err := database.WithRetry(ctx, boff, logger, e.ErrTesting, func() error {
				return tt.query(&attempts)
			})

			require.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedRetries, attempts-1)
		})
	}
}
