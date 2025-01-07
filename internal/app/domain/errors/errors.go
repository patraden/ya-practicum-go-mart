package errors

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrConfigEnvParse          = errors.New("env vars parsing error")
	ErrConfigDBURI             = errors.New("db uri unset")
	ErrModelUserPassHash       = errors.New("user password hashing error")
	ErrModelUserBalanceInvalid = errors.New("invalid operation on user balanced")
	ErrPGEmptyPool             = errors.New("empty database conn pool")
	ErrPGTransaction           = errors.New("db transaction error")
	ErrPGQueryExec             = errors.New("db query execution error")
	ErrRepoUserIDCollision     = errors.New("user id exists")
	ErrRepoUserNotFound        = errors.New("user not found")
	ErrRepoUserBalanceNotFound = errors.New("user balance not found")
	ErrRepoUserExists          = errors.New("user exists")
	ErrRepoUserPassMismatch    = errors.New("user password mismatch")
	ErrUseCaseInternal         = errors.New("internal error")
	ErrAuthInvalidToken        = errors.New("token is invalid")
	ErrAuthGenerateToken       = errors.New("failed to generate new auth token")
	ErrAuthInvalidKeyType      = errors.New("auth key is of invalid type")
	ErrAuthUnknownClaims       = errors.New("unknown auth claims type")
	ErrAuthNoToken             = errors.New("no auth token found")
	ErrJSONUnmarshal           = errors.New("failed to parse json")
	ErrJSONMarshal             = errors.New("failed to create json")
	ErrAdapterMissedEvent      = errors.New("event missed")
	ErrAdpaterDLQEvent         = errors.New("event sent to dlq")
	ErrAdpaterAccrualNotAlive  = errors.New("accrual system is not alive")
	ErrTesting                 = errors.New("general testing error")
)

func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

type AccrualClientError struct {
	StatusCode int
	RetryAfter time.Duration
	Err        error
}

func (e *AccrualClientError) Error() string {
	return fmt.Sprintf("%d: %v", e.StatusCode, e.Err)
}

func (e *AccrualClientError) Unwrap() error {
	return e.Err
}
