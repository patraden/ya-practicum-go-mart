package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
	q "github.com/patraden/ya-practicum-go-mart/internal/app/postgres/queries"
)

type OrderTransactionsRepository struct {
	connPool database.ConnenctionPool
	queries  *q.Queries
	log      *zerolog.Logger
}

func NewOrderTransactionsRepository(pool database.ConnenctionPool, log *zerolog.Logger) *OrderTransactionsRepository {
	return &OrderTransactionsRepository{
		connPool: pool,
		queries:  q.New(pool),
		log:      log,
	}
}

func (repo *OrderTransactionsRepository) withRetry(ctx context.Context, dbOp func() error) error {
	return database.WithRetry(
		ctx,
		backoff.NewExponentialBackOff(),
		repo.log,
		dbOp,
	)
}

func (repo *OrderTransactionsRepository) GetUserBalance(
	ctx context.Context,
	userID uuid.UUID,
) (*model.UserBalance, error) {
	var userBalance *model.UserBalance

	queryFn := func(queries *q.Queries) error {
		sqlUserBalance, err := queries.GetUserBalances(ctx, userID)

		if errors.Is(err, sql.ErrNoRows) {
			userBalance = model.NewUserBalance(userID)

			return nil
		}

		if err != nil {
			return e.Wrap("GetUserBalances", err)
		}

		userBalance = q.ToModelUserBalance(sqlUserBalance)

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", userID.String()).
			Msg("failed to get user balance")

		return nil, err
	}

	return userBalance, nil
}

func (repo *OrderTransactionsRepository) CreateWithdrawal(
	ctx context.Context,
	trx *model.OrderTransaction,
) error {
	queryFn := database.WithinTrx(ctx, repo.connPool, pgx.TxOptions{}, func(queries *q.Queries) error {
		// leveraging a lightweight pg app lock to ensure serialization for withdrawals.
		// lock is scoped by userID and thus is not blocking other user transactions
		if err := queries.LockUserTransactions(ctx, model.LockID(trx.UserID)); err != nil {
			return e.Wrap("LockUserOrder", err)
		}

		_, err := queries.CreateUserWithdrawal(ctx, q.GetCreateUserWithdrawalParamsFromTrx(trx))
		if errors.Is(err, pgx.ErrNoRows) {
			return e.ErrRepoOrderNoFunds
		}

		if err != nil {
			return e.Wrap("GetUserOrderBalance", err)
		}

		return nil
	})

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", trx.UserID.String()).
			Int64("order_id", trx.OrderID).
			Float64("amount", trx.Amount.InexactFloat64()).
			Msg("failed to make withdrawal")

		return err
	}

	return nil
}

func (repo *OrderTransactionsRepository) GetUserWithdrawals(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.OrderTransaction, error) {
	var orderTrxs []model.OrderTransaction

	queryFn := func(queries *q.Queries) error {
		rows, err := queries.GetUserWithdrawals(ctx, userID)
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoOrderNoWithdrawals
		}

		orderTrxs = make([]model.OrderTransaction, len(rows))
		for i, row := range rows {
			orderTrxs[i] = q.ToModelOrderTransaction(row, userID, false)
		}

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", userID.String()).
			Msg("failed to get withdrawals")

		return []model.OrderTransaction{}, err
	}

	return orderTrxs, nil
}
