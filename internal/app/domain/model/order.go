package model

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	LuhnMod       = 2
	LuhnThreshold = 9
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
	ID        int64     `json:"order_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewOrder(id int64, userID uuid.UUID) *Order {
	return &Order{
		ID:        id,
		UserID:    userID,
		Status:    StatusNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (ord *Order) CreatedAtUnix() int64 {
	return ord.CreatedAt.UnixNano()
}

func (ord *Order) checkLuhn() bool {
	parity := int(ord.ID % LuhnMod)
	idStr := strconv.FormatInt(ord.ID, 10)
	sum := 0

	for i, digitRune := range idStr {
		mod := LuhnMod
		digit, _ := strconv.Atoi(string(digitRune))

		if i%mod == parity {
			digit *= mod
			if digit > LuhnThreshold {
				digit -= LuhnThreshold
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
