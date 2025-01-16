package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
	q "github.com/patraden/ya-practicum-go-mart/internal/app/postgres/queries"
)

type testCase struct {
	name          string
	queryFn       database.QueryFunc
	mockSetup     func()
	expectedError error
}

func TestWithTransaction(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	defer mockPool.Close()

	tests := []testCase{
		{
			name:          "success",
			queryFn:       func(*q.Queries) error { return nil },
			mockSetup:     func() { mockPool.ExpectBegin(); mockPool.ExpectCommit() },
			expectedError: nil,
		},
		{
			name:          "query function fails",
			queryFn:       func(*q.Queries) error { return e.ErrTesting },
			mockSetup:     func() { mockPool.ExpectBegin(); mockPool.ExpectRollback() },
			expectedError: e.ErrTesting,
		},
		{
			name:          "begin transaction fails",
			queryFn:       func(*q.Queries) error { return nil },
			mockSetup:     func() { mockPool.ExpectBegin().WillReturnError(e.ErrTesting) },
			expectedError: e.ErrTesting,
		},
		{
			name:    "commit transaction fails",
			queryFn: func(*q.Queries) error { return nil },
			mockSetup: func() {
				mockPool.ExpectBegin()
				mockPool.ExpectCommit().WillReturnError(e.ErrTesting)
				mockPool.ExpectRollback()
			}, expectedError: e.ErrTesting,
		},
		{
			name:          "rollback transaction fails",
			queryFn:       func(*q.Queries) error { return e.ErrTesting },
			mockSetup:     func() { mockPool.ExpectBegin(); mockPool.ExpectRollback().WillReturnError(e.ErrTesting) },
			expectedError: e.ErrTesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			trxQueryFn := database.WithinTrx(context.Background(), mockPool, pgx.TxOptions{}, tt.queryFn)
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			queries := q.New(mockPool)
			err = trxQueryFn(queries)

			require.ErrorIs(t, err, tt.expectedError)
			err = mockPool.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
