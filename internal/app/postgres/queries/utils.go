package sql

import "github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"

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

func CreateUserBalancesParamsFromModel(balance *model.UserBalance) *CreateUserBalancesParams {
	return &CreateUserBalancesParams{
		Userid:    balance.UserID,
		Balance:   balance.Balance,
		Withdrawn: balance.Withdrawn,
		UpdatedAt: balance.UpdatedAt,
	}
}

func ToModelUserBalance(balance UserBalance) *model.UserBalance {
	return &model.UserBalance{
		UserID:    balance.Userid,
		Balance:   balance.Balance,
		Withdrawn: balance.Withdrawn,
		UpdatedAt: balance.UpdatedAt,
	}
}
