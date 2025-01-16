package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	UpdateStatus(ctx context.Context, orderStatus *model.OrderStatus) error
}
