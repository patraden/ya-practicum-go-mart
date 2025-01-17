package dto

import (
	"github.com/google/uuid"
)

//easyjson:json
type Withdrawal struct {
	OrderID string    `json:"order"`
	UserID  uuid.UUID `json:"-"`
	Amount  float64   `json:"sum"`
}

func (w *Withdrawal) SetUser(uuid uuid.UUID) {
	w.UserID = uuid
}

//easyjson:json
type WithdrawalResponse struct {
	OrderID   string  `json:"order"`
	Amount    float64 `json:"sum"`
	CreatedAt string  `json:"processed_at"`
}

//easyjson:json
type WithdrawalResponseBatch []WithdrawalResponse
