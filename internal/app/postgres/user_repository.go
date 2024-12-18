package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
	q "github.com/patraden/ya-practicum-go-mart/internal/app/postgres/queries"
)

type UserRepository struct {
	connPool database.ConnenctionPool
	queries  *q.Queries
	log      *zerolog.Logger
}

func NewUserRepository(pool database.ConnenctionPool, log *zerolog.Logger) *UserRepository {
	return &UserRepository{
		connPool: pool,
		queries:  q.New(pool),
		log:      log,
	}
}

func (repo *UserRepository) withRetry(ctx context.Context, query database.QueryFunc) error {
	return database.WithRetry(
		ctx,
		backoff.NewExponentialBackOff(),
		repo.log,
		e.ErrRepoUserIDCollision,
		query,
	)
}

func (repo *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	var repoUser *model.User

	query := func() error {
		sqluser, err := repo.queries.CreateUser(ctx, *q.CreateUserParamsFromModel(user))
		if err != nil {
			repo.log.
				Error().Err(err).
				Str("user_id", user.ID.String()).
				Str("user_name", user.Username).
				Msg("user create query failure")

			return e.Wrap("repo create user", err)
		}

		repoUser = q.ToModelUser(sqluser.NoPassword())

		if repoUser.ID != user.ID {
			return e.ErrRepoUserExists
		}

		balance := model.NewUserBalance(repoUser.ID)
		if err = repo.queries.CreateUserBalances(ctx, *q.CreateUserBalancesParamsFromModel(balance)); err != nil {
			repo.log.
				Error().Err(err).
				Str("user_id", user.ID.String()).
				Str("user_name", user.Username).
				Msg("user balance create query failure")

			return e.Wrap("repo create user", err)
		}

		return nil
	}

	queryInTrx := database.WithTransaction(ctx, query, repo.connPool, pgx.TxOptions{})

	if err := repo.withRetry(ctx, queryInTrx); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", user.ID.String()).
			Str("user_name", user.Username).
			Msg("failed to create user")

		return nil, e.Wrap("repo create user", err)
	}

	return repoUser, nil
}

func (repo *UserRepository) GetUser(ctx context.Context, username string) (*model.User, error) {
	var repoUser *model.User

	query := func() error {
		sqluser, err := repo.queries.GetUser(ctx, username)

		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoUserNotFound
		}

		if err != nil {
			repo.log.
				Error().Err(err).
				Str("user_name", username).
				Msg("failed to execute user get query")

			return e.Wrap("repo get user", err)
		}

		repoUser = q.ToModelUser(sqluser)

		return nil
	}

	if err := repo.withRetry(ctx, query); err != nil {
		return nil, e.Wrap("repo get user", err)
	}

	return repoUser, nil
}

func (repo *UserRepository) ValidateUser(ctx context.Context, username, password string) (*model.User, error) {
	user, err := repo.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if !user.CheckPassword(password) {
		return nil, e.ErrRepoUserPassMismatch
	}

	return user.NoPassword(), nil
}

func (repo *UserRepository) GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error) {
	var userBalance *model.UserBalance

	query := func() error {
		sqlUserBalance, err := repo.queries.GetUserBalances(ctx, userID)

		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoUserBalanceNotFound
		}

		if err != nil {
			repo.log.
				Error().Err(err).
				Str("user_id", userID.String()).
				Msg("failed to execute user balance get query")

			return e.Wrap("repo get user balance", err)
		}

		userBalance = q.ToModelUserBalance(sqlUserBalance)

		return nil
	}

	if err := repo.withRetry(ctx, query); err != nil {
		return nil, e.Wrap("repo get user balance", err)
	}

	return userBalance, nil
}
