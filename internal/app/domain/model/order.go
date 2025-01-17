package model

import (
	"strconv"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusRegistered Status = "REGISTERED"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func (s Status) Valid() bool {
	return s == StatusNew ||
		s == StatusInvalid ||
		s == StatusRegistered ||
		s == StatusProcessing ||
		s == StatusProcessed
}

type Order struct {
	ID             int64           `json:"order_id"`
	UserID         uuid.UUID       `json:"user_id"`
	Status         Status          `json:"status"`
	Accrual        decimal.Decimal `json:"accrual"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"-"`
	CreatedAtEpoch int64           `json:"-"`
}

func NewOrder(id int64, userID uuid.UUID) *Order {
	t := time.Now().UTC()

	return &Order{
		ID:             id,
		UserID:         userID,
		Status:         StatusNew,
		Accrual:        decimal.Zero,
		CreatedAt:      t,
		UpdatedAt:      t,
		CreatedAtEpoch: t.UnixNano(),
	}
}

func (ord *Order) CreatedAtUnix() int64 {
	return ord.CreatedAt.UnixNano()
}

func (ord *Order) CheckLuhn() bool {
	err := goluhn.Validate(strconv.FormatInt(ord.ID, 10))

	return err == nil
}
