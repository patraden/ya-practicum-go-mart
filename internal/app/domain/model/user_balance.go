package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
)

type UserBalance struct {
	UserID    uuid.UUID       `json:"-"`
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func NewUserBalance(userID uuid.UUID) *UserBalance {
	return &UserBalance{
		UserID:    userID,
		Balance:   decimal.Zero,
		Withdrawn: decimal.Zero,
	}
}

func (balance *UserBalance) Accrual(amount decimal.Decimal) error {
	if amount.LessThan(decimal.Zero) {
		return e.ErrModelUserBalanceInvalid
	}

	balance.Balance = balance.Balance.Add(amount)

	return nil
}

func (balance *UserBalance) Withdraw(amount decimal.Decimal) error {
	if balance.Balance.Sub(balance.Withdrawn.Add(amount)).LessThan(decimal.Zero) {
		return e.ErrModelUserBalanceInvalid
	}

	balance.Withdrawn = balance.Withdrawn.Add(amount)

	return nil
}
