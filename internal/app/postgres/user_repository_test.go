package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres"
)

const (
	username            = "user1"
	userPassStr         = "user1pass"
	existingUserPassStr = "user2pass"
)

func SetupUserRepoCreate(
	t *testing.T,
	mockPool pgxmock.PgxPoolIface,
) (context.Context, *model.User, *model.User, *postgres.UserRepository) {
	t.Helper()

	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	user := model.NewUser(username)
	existingUser := model.NewUser(username)

	repo := postgres.NewUserRepository(mockPool, log)
	ctx := context.Background()

	err := user.SetPassword(userPassStr)
	require.NoError(t, err)

	err = existingUser.SetPassword(existingUserPassStr)
	require.NoError(t, err)

	assert.True(t, user.CheckPassword(userPassStr))
	assert.True(t, existingUser.CheckPassword(existingUserPassStr))

	return ctx, user, existingUser, repo
}

func TestUserRepoCreateSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))

	result, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	require.Equal(t, user.ID, result.ID)
	require.Equal(t, user.Username, result.Username)
	require.Empty(t, result.Password)
	require.Equal(t, user.CreatedAt, result.CreatedAt)
	require.Equal(t, user.UpdatedAt, result.UpdatedAt)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoCreateFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, exUser, repo := SetupUserRepoCreate(t, mockPool)

	// unique vialation for duplicate user
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "UniqueViolation"})

	res, err := repo.CreateUser(ctx, user)
	require.ErrorIs(t, err, e.ErrRepoUserIDCollision)
	assert.Nil(t, res)

	// User exists
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(exUser.ID, exUser.Username, exUser.Password, exUser.CreatedAt, exUser.UpdatedAt))

	res, err = repo.CreateUser(ctx, user)
	require.ErrorIs(t, err, e.ErrRepoUserExists)
	assert.Nil(t, res)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoCreateRetriable(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	// First retry
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionFailure})

	// Second retry
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection})

	// Success on third try
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))

	result, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	require.Equal(t, user.ID, result.ID)
	require.Equal(t, user.Username, result.Username)
	require.Empty(t, result.Password)
	require.Equal(t, user.CreatedAt, result.CreatedAt)
	require.Equal(t, user.UpdatedAt, result.UpdatedAt)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoGetSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))

	result, err := repo.GetUser(ctx, user.Username)
	require.NoError(t, err)

	require.Equal(t, user.ID, result.ID)
	require.Equal(t, user.Username, result.Username)
	require.Equal(t, user.Password, result.Password)
	require.Equal(t, user.CreatedAt, result.CreatedAt)
	require.Equal(t, user.UpdatedAt, result.UpdatedAt)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoGetFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	// User not found
	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnError(sql.ErrNoRows)

	res, err := repo.GetUser(ctx, user.Username)
	require.ErrorIs(t, err, e.ErrRepoUserNotFound)
	assert.Nil(t, res)

	// arbitrary error
	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})

	res, err = repo.GetUser(ctx, user.Username)
	require.Error(t, err)
	assert.Nil(t, res)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoValidateSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))

	_, err = repo.ValidateUser(ctx, user.Username, userPassStr)
	require.NoError(t, err)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoValidateFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, exUser, repo := SetupUserRepoCreate(t, mockPool)

	// User not found
	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.ValidateUser(ctx, user.Username, userPassStr)
	require.ErrorIs(t, err, e.ErrRepoUserNotFound)

	// arbitrary error
	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})

	_, err = repo.ValidateUser(ctx, user.Username, userPassStr)
	require.Error(t, err)

	// password mismatch
	mockPool.
		ExpectQuery(`SELECT id, username, password, created_at, updated_at`).
		WithArgs(user.Username).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(exUser.ID, exUser.Username, exUser.Password, exUser.CreatedAt, exUser.UpdatedAt))

	_, err = repo.ValidateUser(ctx, user.Username, userPassStr)
	require.ErrorIs(t, err, e.ErrRepoUserPassMismatch)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}
