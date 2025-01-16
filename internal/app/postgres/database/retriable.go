package database

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
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
	dbOP func() error,
) error {
	operation := func() error {
		err := dbOP()
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

		return backoff.Permanent(err)
	}

	return backoff.Retry(operation, backoff.WithContext(boff, ctx))
}
