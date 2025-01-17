package usecase

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/repository"
)

type TransactionsUseCase struct {
	repo repository.OrderTransactionsRepository
	log  *zerolog.Logger
}

func NewTransactionsUseCase(repo repository.OrderTransactionsRepository, log *zerolog.Logger) *TransactionsUseCase {
	return &TransactionsUseCase{
		repo: repo,
		log:  log,
	}
}

func (u *TransactionsUseCase) GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error) {
	balance, err := u.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, e.ErrUseCaseInternal
	}

	return balance, nil
}

func (u *TransactionsUseCase) CreateWithdrawal(ctx context.Context, wdl *dto.Withdrawal) error {
	orderID, err := strconv.ParseInt(wdl.OrderID, 10, 64)
	if err != nil {
		u.log.Error().
			Str("user_id", wdl.UserID.String()).
			Str("order_id", wdl.OrderID).
			Msg("Could not parse order to int64")

		return e.ErrUseCaseBadOrder
	}

	order := model.NewOrder(orderID, wdl.UserID)
	if !order.CheckLuhn() {
		u.log.Error().
			Str("user_id", wdl.UserID.String()).
			Str("order_id", wdl.OrderID).
			Msg("Order is not Luhn complient")

		return e.ErrUseCaseBadOrder
	}

	trx := model.NewWithdrawal(orderID, wdl.UserID, decimal.NewFromFloat(wdl.Amount))

	err = u.repo.CreateWithdrawal(ctx, trx)

	if errors.Is(err, e.ErrRepoOrderNoFunds) {
		return e.ErrRepoOrderNoFunds
	}

	if err != nil {
		return e.ErrUseCaseInternal
	}

	return nil
}

func (u *TransactionsUseCase) GetUserWithdrawals(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.OrderTransaction, error) {
	transactions, err := u.repo.GetUserWithdrawals(ctx, userID)
	if errors.Is(err, e.ErrRepoOrderNoWithdrawals) {
		u.log.Error().Err(err).
			Str("user_id", userID.String()).
			Msg("No withdrawals")

		return []model.OrderTransaction{}, e.ErrRepoOrderNoWithdrawals
	}

	if err != nil {
		return []model.OrderTransaction{}, e.ErrUseCaseInternal
	}

	return transactions, nil
}
