package errors

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrConfigEnvParse            = errors.New("env vars parsing error")
	ErrConfigDBURI               = errors.New("db uri unset")
	ErrModelUserPassHash         = errors.New("user password hashing error")
	ErrModelUserBalanceInvalid   = errors.New("invalid operation on user balanced")
	ErrPGEmptyPool               = errors.New("empty database conn pool")
	ErrPGTransaction             = errors.New("db transaction error")
	ErrPGQueryExec               = errors.New("db query execution error")
	ErrRepoUserIDCollision       = errors.New("user id exists")
	ErrRepoOrderIDCollision      = errors.New("order id exists")
	ErrRepoOrderNotFound         = errors.New("order not found")
	ErrRepoOrderExists           = errors.New("order exists for another user")
	ErrRepoOrderUserExists       = errors.New("order exists")
	ErrRepoUserExists            = errors.New("user exists")
	ErrRepoUserNotFound          = errors.New("user not found")
	ErrRepoOrderNoOrders         = errors.New("no user orders")
	ErrRepoOrderNoWithdrawals    = errors.New("no user withdrawals")
	ErrRepoOrderNoFunds          = errors.New("not enough balance for trx")
	ErrRepoUserPassMismatch      = errors.New("user password mismatch")
	ErrUseCaseBadOrder           = errors.New("bad order format")
	ErrUseCaseInternal           = errors.New("internal error")
	ErrUseCasePassword           = errors.New("failed to set password")
	ErrAuthInvalidToken          = errors.New("token is invalid")
	ErrAuthGenerateToken         = errors.New("failed to generate new auth token")
	ErrAuthInvalidKeyType        = errors.New("auth key is of invalid type")
	ErrAuthUnknownClaims         = errors.New("unknown auth claims type")
	ErrAuthNoToken               = errors.New("no auth token found")
	ErrJSONUnmarshal             = errors.New("failed to parse json")
	ErrJSONMarshal               = errors.New("failed to create json")
	ErrAdapterMissedEvent        = errors.New("event missed")
	ErrAdpaterDLQEvent           = errors.New("event sent to dlq")
	ErrAdpaterAccrualNotAlive    = errors.New("accrual system is not alive")
	ErrAccrualOrderNotRegistered = errors.New("order not registered in the accrual system")
	ErrAccrualInternalServer     = errors.New("internal server error from accrual system")
	ErrAccrualUnknownResponse    = errors.New("unknown response code from accrual system")
	ErrTesting                   = errors.New("general testing error")
)

type AccrualTooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *AccrualTooManyRequestsError) Error() string {
	return fmt.Sprintf("too many requests, retry after %v", e.RetryAfter)
}

// Wrap wraps an error with additional context.
func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}
