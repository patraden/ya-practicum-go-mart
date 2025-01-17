package dto

import (
	"github.com/shopspring/decimal"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

//easyjson:json
type OrderStatusAccrual struct {
	ID      string          `json:"order"`
	Status  model.Status    `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

//easyjson:json
type OrderStatusResponse struct {
	ID        string       `json:"number"`
	Status    model.Status `json:"status"`
	Accrual   float64      `json:"accrual,omitempty"`
	CreatedAt string       `json:"uploaded_at"`
}

//easyjson:json
type OrderStatusResponseBatch []OrderStatusResponse
