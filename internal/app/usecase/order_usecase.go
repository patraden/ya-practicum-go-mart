package usecase

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/integration/accrual"
	"github.com/patraden/ya-practicum-go-mart/internal/app/repository"
)

type OrderUseCase struct {
	repo    repository.OrderRepository
	adapter *accrual.Adapter
	log     *zerolog.Logger
}

func NewOrderUseCase(repo repository.OrderRepository, adapter *accrual.Adapter, log *zerolog.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo:    repo,
		adapter: adapter,
		log:     log,
	}
}

func (u *OrderUseCase) CreateOrder(ctx context.Context, userID uuid.UUID, orderID string) (*model.Order, error) {
	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		u.log.Error().
			Str("user_id", userID.String()).
			Str("order_id", orderID).
			Msg("Could not parse order to int64")

		return nil, e.ErrUseCaseBadOrder
	}

	order := model.NewOrder(id, userID)
	if !order.CheckLuhn() {
		u.log.Error().
			Str("user_id", userID.String()).
			Str("order_id", orderID).
			Msg("Order is not complient with Luhn algo")

		return nil, e.ErrUseCaseBadOrder
	}

	repoOrder, err := u.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, e.ErrUseCaseInternal
	}

	if repoOrder.UserID != order.UserID {
		u.log.Error().
			Str("user_id", userID.String()).
			Str("another_user_id", repoOrder.UserID.String()).
			Str("order_id", orderID).
			Msg("Order registered by another user")

		return nil, e.ErrRepoOrderExists
	}

	if repoOrder.CreatedAtEpoch != order.CreatedAtEpoch {
		u.log.Error().
			Str("user_id", repoOrder.UserID.String()).
			Int64("order_id", repoOrder.ID).
			Int64("epoch", repoOrder.CreatedAtEpoch).
			Msg("Order already registered")

		return nil, e.ErrRepoOrderUserExists
	}

	orderStatus := model.NewOrderStatus(repoOrder.ID, repoOrder.UserID, repoOrder.Accrual)
	if ok := u.adapter.SubmitOrder(orderStatus); !ok {
		u.log.Error().
			Int64("order_id", repoOrder.ID).
			Str("user_id", repoOrder.UserID.String()).
			Msg("Order created but integration event is not submitted.")
	}

	return repoOrder, nil
}

func (u *OrderUseCase) GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	orders, err := u.repo.GetOrders(ctx, userID)
	if err != nil {
		return []model.Order{}, e.ErrUseCaseInternal
	}

	return orders, nil
}
