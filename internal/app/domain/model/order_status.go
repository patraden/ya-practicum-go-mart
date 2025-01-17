package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus struct {
	ID      int64           `json:"order_id"`
	UserID  uuid.UUID       `json:"-"`
	Status  Status          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func NewOrderStatus(id int64, userID uuid.UUID, accrual decimal.Decimal) *OrderStatus {
	return &OrderStatus{
		ID:      id,
		UserID:  userID,
		Status:  StatusNew,
		Accrual: accrual,
	}
}

func (os *OrderStatus) ChangeStatus(status Status) *OrderStatus {
	return &OrderStatus{
		ID:      os.ID,
		UserID:  os.UserID,
		Status:  status,
		Accrual: os.Accrual,
	}
}
