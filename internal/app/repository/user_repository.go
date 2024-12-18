package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	ValidateUser(ctx context.Context, username, password string) (*model.User, error)
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error)
}
