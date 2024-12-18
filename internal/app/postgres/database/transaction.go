package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// query function wrapper into database transaction.
func WithTransaction(
	ctx context.Context,
	queryfn QueryFunc,
	connPool ConnenctionPool,
	trxOptions pgx.TxOptions,
) QueryFunc {
	return func() (err error) {
		trx, beginErr := connPool.BeginTx(ctx, trxOptions)
		if beginErr != nil {
			err = beginErr

			return
		}

		defer func() {
			if err != nil {
				if rollbackErr := trx.Rollback(ctx); rollbackErr != nil {
					err = rollbackErr

					return
				}
			}
		}()

		if fnErr := queryfn(); fnErr != nil {
			return fnErr
		}

		err = trx.Commit(ctx)

		return
	}
}
