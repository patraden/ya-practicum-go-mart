package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
)

type IUserUseCase interface {
	CreateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error)
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error)
}

type ITransactionsUseCase interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error)
	GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]model.OrderTransaction, error)
	CreateWithdrawal(ctx context.Context, wd *dto.Withdrawal) error
}

type IOrderUseCase interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, orderID string) (*model.Order, error)
	GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
}
