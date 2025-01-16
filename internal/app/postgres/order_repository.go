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

type OrderRepository struct {
	connPool database.ConnenctionPool
	queries  *q.Queries
	log      *zerolog.Logger
}

func NewOrderRepository(pool database.ConnenctionPool, log *zerolog.Logger) *OrderRepository {
	return &OrderRepository{
		connPool: pool,
		queries:  q.New(pool),
		log:      log,
	}
}

func (repo *OrderRepository) withRetry(ctx context.Context, query func() error) error {
	return database.WithRetry(
		ctx,
		backoff.NewExponentialBackOff(),
		repo.log,
		query,
	)
}

func (repo *OrderRepository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	var repoOrder *model.Order

	queryFn := func(queries *q.Queries) error {
		sqlOrder, err := queries.CreateOrder(ctx, *q.CreateOrderParamsFromModel(order))
		if err != nil {
			return e.Wrap("CreateOrder", err)
		}

		repoOrder = q.ToModelOrder(sqlOrder)

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Int64("order_id", order.ID).
			Str("user_id", order.UserID.String()).
			Msg("Repo: failed to create order")

		return nil, err
	}

	return repoOrder, nil
}

func (repo *OrderRepository) GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order

	queryFn := func(queries *q.Queries) error {
		slqOrders, err := queries.GetOrders(ctx, userID)

		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoOrderNoOrders
		}

		if err != nil {
			return e.Wrap("GetOrders", err)
		}

		orders = make([]model.Order, len(slqOrders))
		for i, sqlOrder := range slqOrders {
			orders[i] = *q.ToModelOrder(sqlOrder)
		}

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", userID.String()).
			Msg("Repo: failed to get orders")

		return []model.Order{}, err
	}

	return orders, nil
}

func (repo *OrderRepository) UpdateStatus(ctx context.Context, orderStatus *model.OrderStatus) error {
	switch orderStatus.Status {
	case model.StatusProcessed:
		return repo.updateStatusProcessed(ctx, orderStatus)
	case model.StatusNew, model.StatusRegistered, model.StatusProcessing, model.StatusInvalid:
		return repo.updateStatus(ctx, orderStatus)
	}

	return nil
}

func (repo *OrderRepository) updateStatusProcessed(ctx context.Context, orderStatus *model.OrderStatus) error {
	queryFn := database.WithinTrx(ctx, repo.connPool, pgx.TxOptions{}, func(queries *q.Queries) error {
		err := queries.UpdateOrderStatus(ctx, q.UpdateOrderStatusParamsFromStatus(orderStatus))
		if err != nil {
			return e.Wrap("UpdateOrderStatus", err)
		}

		debitTrx := model.NewAccrual(orderStatus.ID, orderStatus.UserID, orderStatus.Accrual)

		err = queries.CreateOrderAccrual(ctx, q.CreateOrderAccrualParamsFromModel(debitTrx))
		if err != nil {
			return e.Wrap("CreateOrderTransaction", err)
		}

		return nil
	})

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Int64("order_id", orderStatus.ID).
			Str("order_status", string(orderStatus.Status)).
			Msg("Repo: failed to update order status")

		return err
	}

	return nil
}

func (repo *OrderRepository) updateStatus(ctx context.Context, orderStatus *model.OrderStatus) error {
	queryFn := func(queries *q.Queries) error {
		err := queries.UpdateOrderStatus(ctx, q.UpdateOrderStatusParamsFromStatus(orderStatus))
		if err != nil {
			return e.Wrap("UpdateOrderStatus", err)
		}

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Int64("order_id", orderStatus.ID).
			Str("order_status", string(orderStatus.Status)).
			Msg("Repo: failed to update order status")

		return err
	}

	return nil
}
