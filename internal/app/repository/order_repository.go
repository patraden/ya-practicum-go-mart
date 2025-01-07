package repository

import (
	"context"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderStatus *dto.OrderStatus) error
}
