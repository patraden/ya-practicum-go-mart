package sql

import (
	"github.com/google/uuid"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

func CreateUserParamsFromModel(user *model.User) *CreateUserParams {
	return &CreateUserParams{
		ID:        user.ID,
		Username:  user.Username,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToModelUser(sqluser User) *model.User {
	return &model.User{
		ID:        sqluser.ID,
		Username:  sqluser.Username,
		Password:  sqluser.Password,
		CreatedAt: sqluser.CreatedAt,
		UpdatedAt: sqluser.UpdatedAt,
	}
}

func (sqluser User) NoPassword() User {
	return User{
		ID:        sqluser.ID,
		Username:  sqluser.Username,
		Password:  []byte{},
		CreatedAt: sqluser.CreatedAt,
		UpdatedAt: sqluser.UpdatedAt,
	}
}

func ToModelUserBalance(row GetUserBalancesRow) *model.UserBalance {
	return &model.UserBalance{
		UserID:    row.Userid,
		Balance:   row.Balance,
		Withdrawn: row.Withdrawn,
	}
}

func CreateOrderParamsFromModel(order *model.Order) *CreateOrderParams {
	return &CreateOrderParams{
		ID:             order.ID,
		Userid:         order.UserID,
		CreatedAt:      order.CreatedAt,
		Status:         order.Status,
		Accrual:        order.Accrual,
		UpdatedAt:      order.UpdatedAt,
		CreatedAtEpoch: order.CreatedAtEpoch,
	}
}

func ToModelOrder(sqlorder Order) *model.Order {
	return &model.Order{
		ID:             sqlorder.ID,
		UserID:         sqlorder.Userid,
		Status:         sqlorder.Status,
		Accrual:        sqlorder.Accrual,
		CreatedAt:      sqlorder.CreatedAt,
		UpdatedAt:      sqlorder.UpdatedAt,
		CreatedAtEpoch: sqlorder.CreatedAtEpoch,
	}
}

func UpdateOrderStatusParamsFromStatus(orderStatus *model.OrderStatus) UpdateOrderStatusParams {
	return UpdateOrderStatusParams{
		ID:      orderStatus.ID,
		Status:  orderStatus.Status,
		Accrual: orderStatus.Accrual,
	}
}

func GetCreateUserWithdrawalParamsFromTrx(trx *model.OrderTransaction) CreateUserWithdrawalParams {
	return CreateUserWithdrawalParams{
		Orderid:        trx.OrderID,
		Userid:         trx.UserID,
		Amount:         trx.Amount,
		CreatedAt:      trx.CreatedAt,
		CreatedAtEpoch: trx.CreatedAtEpoch,
	}
}

func CreateOrderAccrualParamsFromModel(trx *model.OrderTransaction) CreateOrderAccrualParams {
	return CreateOrderAccrualParams{
		Orderid:        trx.OrderID,
		Userid:         trx.UserID,
		Amount:         trx.Amount,
		CreatedAt:      trx.CreatedAt,
		CreatedAtEpoch: trx.CreatedAtEpoch,
	}
}

func ToModelOrderTransaction(row GetUserWithdrawalsRow, userID uuid.UUID, isDebit bool) model.OrderTransaction {
	return model.OrderTransaction{
		OrderID:        row.Orderid,
		UserID:         userID,
		IsDebit:        isDebit,
		Amount:         row.Amount,
		CreatedAt:      row.CreatedAt,
		CreatedAtEpoch: row.CreatedAtEpoch,
	}
}
