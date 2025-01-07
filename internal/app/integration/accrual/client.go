package accrual

import (
	"context"

	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
)

type Client interface {
	IsAlive() bool
	GetOrderStatus(ctx context.Context, orderID int64) (*dto.OrderStatus, error)
}
