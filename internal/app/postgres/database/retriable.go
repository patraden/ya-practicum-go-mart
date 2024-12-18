package database

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
)

func IsRetryableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case
			// Retryable errors
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.CannotConnectNow,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.TransactionResolutionUnknown:
			return true
		}
	}

	return false
}

func WithRetry(
	ctx context.Context,
	boff backoff.BackOff,
	log *zerolog.Logger,
	uniqueViolationError error,
	query QueryFunc,
) error {
	operation := func() error {
		err := query()
		if err == nil {
			return nil
		}

		if IsRetryableError(err) {
			log.
				Info().
				Err(err).
				Msg("retrying after error")

			return err
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return backoff.Permanent(uniqueViolationError)
		}

		return backoff.Permanent(err)
	}

	err := backoff.Retry(operation, backoff.WithContext(boff, ctx))
	if err != nil {
		return e.Wrap("retry error", err)
	}

	return nil
}
