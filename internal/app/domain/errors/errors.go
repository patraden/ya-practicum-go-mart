package errors

import (
	"errors"
	"fmt"
)

var (
	ErrConfigEnvParse       = errors.New("env vars parsing error")
	ErrConfigDBURI          = errors.New("db uri unset")
	ErrUserPassHash         = errors.New("user password hashing error")
	ErrPGEmptyPool          = errors.New("empty database conn pool")
	ErrRepoUserIDCollision  = errors.New("user ID exists")
	ErrRepoUserNotFound     = errors.New("user not found")
	ErrRepoUserExists       = errors.New("user exists")
	ErrRepoUserPassMismatch = errors.New("user password mismatch")
	ErrUseCaseInternal      = errors.New("internal error")
	ErrAuthInvalidToken     = errors.New("token is invalid")
	ErrAuthGenerateToken    = errors.New("failed to generate new auth token")
	ErrAuthInvalidKeyType   = errors.New("auth key is of invalid type")
	ErrAuthUnknownClaims    = errors.New("unknown auth claims type")
	ErrAuthNoToken          = errors.New("no auth token found")
	ErrJSONUnmarshal        = errors.New("failed to parse json")
	ErrJSONMarshal          = errors.New("failed to create json")
	ErrTesting              = errors.New("general testing error")
)

func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}
