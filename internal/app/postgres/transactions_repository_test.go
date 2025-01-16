package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres"
)

func setupOrderTransactionRepo(t *testing.T, mockPool pgxmock.PgxPoolIface) (
	context.Context,
	*model.User,
	*model.OrderTransaction,
	*postgres.OrderTransactionsRepository,
) {
	t.Helper()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	user := model.NewUser(username)
	repo := postgres.NewOrderTransactionsRepository(mockPool, log)
	ctx := context.Background()
	trx := model.NewWithdrawal(1, user.ID, decimal.NewFromFloat32(10.0))

	return ctx, user, trx, repo
}

func TestOrderTrxRepoGetBalanceSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := setupOrderTransactionRepo(t, mockPool)

	mockPool.
		ExpectQuery(`SELECT userID AS userid`).
		WithArgs(user.ID).
		WillReturnRows(pgxmock.NewRows([]string{"userID", "balance", "withdrawn"}).
			AddRow(user.ID, decimal.Zero, decimal.Zero))

	result, err := repo.GetUserBalance(ctx, user.ID)
	require.NoError(t, err)

	assert.Equal(t, user.ID, result.UserID)
	assert.Equal(t, result.Balance, decimal.Zero)
	assert.Equal(t, result.Withdrawn, decimal.Zero)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderTrxRepoGetBalanceFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := setupOrderTransactionRepo(t, mockPool)

	// User balance not found
	mockPool.
		ExpectQuery(`SELECT userID AS userid,\s*SUM\(CASE WHEN is_debit THEN amount ELSE -amount END\)`).
		WithArgs(user.ID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetUserBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.UserID)
	assert.Equal(t, result.Balance, decimal.Zero)
	assert.Equal(t, result.Withdrawn, decimal.Zero)

	// arbitrary error
	mockPool.
		ExpectQuery(`SELECT userID AS userid,\s*SUM\(CASE WHEN is_debit THEN amount ELSE -amount END\)`).
		WithArgs(user.ID).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})

	_, err = repo.GetUserBalance(ctx, user.ID)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, pgerrcode.AdminShutdown, pgErr.Code)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderTrxRepoCreateWithdrawalSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, _, trx, repo := setupOrderTransactionRepo(t, mockPool)

	mockPool.ExpectBegin()

	mockPool.
		ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(model.LockID(trx.UserID)).
		WillReturnResult(pgxmock.NewResult("SELECT", 1))

	mockPool.
		ExpectQuery(`WITH order_balance AS.*INSERT INTO order_transactions`).
		WithArgs(trx.OrderID, trx.UserID, trx.Amount, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"orderId"}).AddRow(trx.OrderID))

	mockPool.ExpectCommit()

	err = repo.CreateWithdrawal(ctx, trx)
	require.NoError(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderTrxRepoCreateWithdrawalNoFunds(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, _, trx, repo := setupOrderTransactionRepo(t, mockPool)

	mockPool.ExpectBegin()

	mockPool.
		ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(model.LockID(trx.UserID)).
		WillReturnResult(pgxmock.NewResult("SELECT", 1))

	mockPool.
		ExpectQuery(`WITH order_balance AS.*INSERT INTO order_transactions`).
		WithArgs(trx.OrderID, trx.UserID, trx.Amount, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"orderId"}))

	mockPool.ExpectRollback()

	err = repo.CreateWithdrawal(ctx, trx)
	require.ErrorIs(t, err, e.ErrRepoOrderNoFunds)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderTrxRepoCreateWithdrawalFailures(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, _, trx, repo := setupOrderTransactionRepo(t, mockPool)

	// failed to lock
	mockPool.ExpectBegin()
	mockPool.
		ExpectExec(`SELECT pg_advisory_xact_lock`).
		WithArgs(model.LockID(trx.UserID)).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})
	mockPool.ExpectRollback()

	err = repo.CreateWithdrawal(ctx, trx)
	require.Error(t, err)

	// failed to create trx
	mockPool.ExpectBegin()
	mockPool.
		ExpectExec(`SELECT pg_advisory_xact_lock`).
		WithArgs(model.LockID(trx.UserID)).
		WillReturnResult(pgxmock.NewResult("SELECT", 1))

	mockPool.
		ExpectQuery(`WITH order_balance AS.*INSERT INTO order_transactions`).
		WithArgs(trx.OrderID, trx.UserID, trx.Amount, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})
	mockPool.ExpectRollback()

	err = repo.CreateWithdrawal(ctx, trx)
	require.Error(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}
