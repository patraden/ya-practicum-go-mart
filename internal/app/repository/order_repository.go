package repository

import (
	"context"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
}
