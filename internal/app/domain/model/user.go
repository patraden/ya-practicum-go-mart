package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(username string) *User {
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Password:  []byte{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return e.ErrUserPassHash
	}

	u.Password = hashedPassword

	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(u.Password, []byte(password)) == nil
}