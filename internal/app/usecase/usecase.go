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
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error)
}
