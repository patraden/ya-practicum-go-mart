package dto

import (
	"github.com/shopspring/decimal"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

type OrderStatus struct {
	ID      int64           `json:"order"`
	Status  model.Status    `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func (os *OrderStatus) ChangeStatus(status model.Status) *OrderStatus {
	return &OrderStatus{
		ID:      os.ID,
		Status:  status,
		Accrual: os.Accrual,
	}
}
