package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderTransaction struct {
	OrderID        int64           `json:"order"`
	UserID         uuid.UUID       `json:"-"`
	IsDebit        bool            `json:"is_debit"`
	Amount         decimal.Decimal `json:"sum"`
	CreatedAt      time.Time       `json:"processed_at"`
	CreatedAtEpoch int64           `json:"-"`
}

func NewAccrual(orderID int64, userID uuid.UUID, amount decimal.Decimal) *OrderTransaction {
	t := time.Now().UTC()

	return &OrderTransaction{
		OrderID:        orderID,
		UserID:         userID,
		IsDebit:        true,
		Amount:         amount,
		CreatedAt:      t,
		CreatedAtEpoch: t.UnixNano(),
	}
}

func NewWithdrawal(orderID int64, userID uuid.UUID, amount decimal.Decimal) *OrderTransaction {
	t := time.Now().UTC()

	return &OrderTransaction{
		OrderID:        orderID,
		UserID:         userID,
		IsDebit:        false,
		Amount:         amount,
		CreatedAt:      t,
		CreatedAtEpoch: t.UnixNano(),
	}
}
