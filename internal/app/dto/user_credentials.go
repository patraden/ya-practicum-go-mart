package dto

//easyjson:json
type UserCredentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}
