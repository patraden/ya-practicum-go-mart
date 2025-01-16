package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

func (repo *UserRepository) withRetry(ctx context.Context, dbOp func() error) error {
	return database.WithRetry(
		ctx,
		backoff.NewExponentialBackOff(),
		repo.log,
		dbOp,
	)
}

func (repo *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	var repoUser *model.User

	queryFn := func(queries *q.Queries) error {
		sqluser, err := queries.CreateUser(ctx, *q.CreateUserParamsFromModel(user))

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return e.ErrRepoUserIDCollision
		}

		if err != nil {
			return e.Wrap("CreateUser", err)
		}

		repoUser = q.ToModelUser(sqluser.NoPassword())

		if repoUser.ID != user.ID {
			return e.ErrRepoUserExists
		}

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_id", user.ID.String()).
			Str("user_name", user.Username).
			Msg("Repo: failed to create user")

		return nil, err
	}

	return repoUser, nil
}

func (repo *UserRepository) GetUser(ctx context.Context, username string) (*model.User, error) {
	var repoUser *model.User

	queryFn := func(queries *q.Queries) error {
		sqluser, err := queries.GetUser(ctx, username)
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrRepoUserNotFound
		}

		if err != nil {
			return e.Wrap("GetUser", err)
		}

		repoUser = q.ToModelUser(sqluser)

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }
	if err := repo.withRetry(ctx, dbOp); err != nil {
		repo.log.
			Error().Err(err).
			Str("user_name", username).
			Msg("Repo: failed to get user")

		return nil, err
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
