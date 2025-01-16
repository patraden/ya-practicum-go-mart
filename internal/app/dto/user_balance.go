package dto

//easyjson:json
type UserBalanceResponse struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
