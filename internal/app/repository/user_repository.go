package repository

import (
	"context"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	ValidateUser(ctx context.Context, username, password string) (*model.User, error)
}
