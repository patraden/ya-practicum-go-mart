package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cenkalti/backoff/v4"
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

func (repo *UserRepository) withRetry(ctx context.Context, query func() error) error {
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
		quser, err := repo.queries.CreateUser(ctx, q.CreateUserParams{
			ID:        user.ID,
			Username:  user.Username,
			Password:  user.Password,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
		if err != nil {
			repo.log.
				Error().Err(err).
				Str("userID", user.ID.String()).
				Str("userName", user.Username).
				Msg("failed to execute user create query")

			return e.Wrap("failed to execute query", err)
		}

		repoUser = &model.User{
			ID:        quser.ID,
			Username:  quser.Username,
			Password:  []byte{},
			CreatedAt: quser.CreatedAt,
			UpdatedAt: quser.UpdatedAt,
		}

		if repoUser.ID != user.ID {
			return e.ErrRepoUserExists
		}

		return nil
	}

	err := repo.withRetry(ctx, query)
	if err != nil {
		return nil, e.Wrap("failed to create user", err)
	}

	return repoUser, nil
}

func (repo *UserRepository) GetUser(ctx context.Context, username string) (*model.User, error) {
	var repoUser *model.User

	query := func() error {
		quser, err := repo.queries.GetUser(ctx, username)

		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoUserNotFound
		}

		if err != nil {
			repo.log.
				Error().Err(err).
				Str("user_name", username).
				Msg("failed to execute user get query")

			return e.Wrap("failed to execute query", err)
		}

		repoUser = &model.User{
			ID:        quser.ID,
			Username:  quser.Username,
			Password:  quser.Password,
			CreatedAt: quser.CreatedAt,
			UpdatedAt: quser.UpdatedAt,
		}

		return nil
	}

	err := repo.withRetry(ctx, query)
	if err != nil {
		return repoUser, e.Wrap("failed to get user", err)
	}

	return repoUser, nil
}

func (repo *UserRepository) ValidateUser(ctx context.Context, username, password string) error {
	user, err := repo.GetUser(ctx, username)
	if err != nil {
		return err
	}

	if !user.CheckPassword(password) {
		return e.ErrRepoUserPassMismatch
	}

	return nil
}
