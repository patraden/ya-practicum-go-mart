package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type OrderTransactionsRepository interface {
	CreateWithdrawal(ctx context.Context, trx *model.OrderTransaction) error
	GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.OrderTransaction, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error)
}
