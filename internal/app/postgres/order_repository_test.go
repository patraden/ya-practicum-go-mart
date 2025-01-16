package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres"
)

func setupOrderRepo(t *testing.T, mockPool pgxmock.PgxPoolIface) (
	context.Context,
	*model.Order,
	*postgres.OrderRepository,
) {
	t.Helper()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	order := model.NewOrder(1, uuid.New())
	repo := postgres.NewOrderRepository(mockPool, log)
	ctx := context.Background()

	return ctx, order, repo
}

func TestOrderRepoCreateOrderSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, order, repo := setupOrderRepo(t, mockPool)

	mockPool.
		ExpectQuery(`INSERT INTO orders \(id, userid, created_at, status, accrual, updated_at, created_at_epoch\)`).
		WithArgs(order.ID, order.UserID, order.CreatedAt, order.Status, order.Accrual, order.UpdatedAt, order.CreatedAtEpoch).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "userid", "created_at", "status", "accrual", "updated_at", "created_at_epoch",
		}).
			AddRow(order.ID, order.UserID, order.CreatedAt, order.Status, order.Accrual, order.UpdatedAt, order.CreatedAtEpoch))

	_, err = repo.CreateOrder(ctx, order)
	require.NoError(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderRepoCreateOrderFailures(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, order, repo := setupOrderRepo(t, mockPool)

	// random error
	mockPool.
		ExpectQuery(`INSERT INTO orders \(id, userid, created_at, status, accrual, updated_at, created_at_epoch\)`).
		WithArgs(order.ID, order.UserID, order.CreatedAt, order.Status, order.Accrual, order.UpdatedAt, order.CreatedAtEpoch).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.DuplicateColumn})

	_, err = repo.CreateOrder(ctx, order)
	require.Error(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderRepoUpdateStatusProcessedSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, order, repo := setupOrderRepo(t, mockPool)
	orderStatus := model.
		NewOrderStatus(order.ID, order.UserID, decimal.NewFromFloat(10.0)).
		ChangeStatus(model.StatusProcessed)

	mockPool.ExpectBegin()
	mockPool.
		ExpectExec(`UPDATE orders`).
		WithArgs(orderStatus.Status, orderStatus.Accrual, order.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mockPool.
		ExpectExec(`INSERT INTO order_transactions`).
		WithArgs(orderStatus.ID, orderStatus.UserID, orderStatus.Accrual, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockPool.ExpectCommit()

	err = repo.UpdateStatus(ctx, orderStatus)
	require.NoError(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderRepoUpdateStatusProcessedFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, order, repo := setupOrderRepo(t, mockPool)
	orderStatus := model.
		NewOrderStatus(order.ID, order.UserID, decimal.NewFromFloat(10.0)).
		ChangeStatus(model.StatusProcessed)

		// first statement failure
	mockPool.ExpectBegin()
	mockPool.
		ExpectExec(`UPDATE orders`).
		WithArgs(orderStatus.Status, orderStatus.Accrual, order.ID).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})
	mockPool.ExpectRollback()

	err = repo.UpdateStatus(ctx, orderStatus)
	require.Error(t, err)

	// second statement failure
	mockPool.ExpectBegin()
	mockPool.
		ExpectExec(`UPDATE orders`).
		WithArgs(orderStatus.Status, orderStatus.Accrual, order.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mockPool.
		ExpectExec(`INSERT INTO order_transactions`).
		WithArgs(orderStatus.ID, orderStatus.UserID, orderStatus.Accrual, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
	mockPool.ExpectRollback()

	err = repo.UpdateStatus(ctx, orderStatus)
	require.Error(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestOrderRepoUpdateStatusAny(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, order, repo := setupOrderRepo(t, mockPool)
	orderStatus := model.
		NewOrderStatus(order.ID, order.UserID, decimal.NewFromFloat(10.0)).
		ChangeStatus(model.StatusProcessing)

	// success
	mockPool.
		ExpectExec(`UPDATE orders`).
		WithArgs(orderStatus.Status, orderStatus.Accrual, order.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateStatus(ctx, orderStatus)
	require.NoError(t, err)

	// failure
	mockPool.
		ExpectExec(`UPDATE orders`).
		WithArgs(orderStatus.Status, orderStatus.Accrual, order.ID).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})

	err = repo.UpdateStatus(ctx, orderStatus)
	require.Error(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}
