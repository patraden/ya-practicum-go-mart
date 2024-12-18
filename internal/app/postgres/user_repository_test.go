package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
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

	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))
	mockPool.
		ExpectExec(`INSERT INTO user_balances \(userID, balance, withdrawn, updated_at\)`).
		WithArgs(user.ID, decimal.Zero, decimal.Zero, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockPool.ExpectCommit()

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

func TestUserRepoCreateFailureDuplicateID(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "UniqueViolation"})
	mockPool.ExpectRollback()

	res, err := repo.CreateUser(ctx, user)
	require.ErrorIs(t, err, e.ErrRepoUserIDCollision)
	assert.Nil(t, res)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoCreateFailureDuplicateUsername(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, exUser, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(exUser.ID, exUser.Username, exUser.Password, exUser.CreatedAt, exUser.UpdatedAt))
	mockPool.ExpectRollback()

	res, err := repo.CreateUser(ctx, user)
	require.ErrorIs(t, err, e.ErrRepoUserExists)
	assert.Nil(t, res)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoCreateFailureDuplicateBalance(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))
	mockPool.
		ExpectExec(`INSERT INTO user_balances \(userID, balance, withdrawn, updated_at\)`).
		WithArgs(user.ID, decimal.Zero, decimal.Zero, pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "UniqueViolation"})
	mockPool.ExpectRollback()

	res, err := repo.CreateUser(ctx, user)
	require.ErrorIs(t, err, e.ErrRepoUserIDCollision)
	assert.Nil(t, res)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoCreateRetriable(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	// First try
	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionFailure})
	mockPool.ExpectRollback()

	// Second try
	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection})
	mockPool.ExpectRollback()

	// Third try
	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))
	mockPool.
		ExpectExec(`INSERT INTO user_balances \(userID, balance, withdrawn, updated_at\)`).
		WithArgs(user.ID, decimal.Zero, decimal.Zero, pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionFailure})
	mockPool.ExpectRollback()

	// Success on thouht try
	mockPool.ExpectBegin()
	mockPool.
		ExpectQuery(`INSERT INTO users \(id, username, password, created_at, updated_at\)`).
		WithArgs(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
			AddRow(user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt))
	mockPool.
		ExpectExec(`INSERT INTO user_balances \(userID, balance, withdrawn, updated_at\)`).
		WithArgs(user.ID, decimal.Zero, decimal.Zero, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockPool.ExpectCommit()

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

	// user not found
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

func TestUserRepoGetBalanceSuccess(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	mockPool.
		ExpectQuery(`SELECT userID, balance, withdrawn, updated_at`).
		WithArgs(user.ID).
		WillReturnRows(pgxmock.NewRows([]string{"userID", "balance", "withdrawn", "updated_at"}).
			AddRow(user.ID, decimal.Zero, decimal.Zero, user.UpdatedAt))

	result, err := repo.GetUserBalance(ctx, user.ID)
	require.NoError(t, err)

	require.Equal(t, user.ID, result.UserID)
	assert.Equal(t, result.Balance, decimal.Zero)
	assert.Equal(t, result.Withdrawn, decimal.Zero)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestUserRepoGetBalanceFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	ctx, user, _, repo := SetupUserRepoCreate(t, mockPool)

	// User balance not found
	mockPool.
		ExpectQuery(`SELECT userID, balance, withdrawn, updated_at`).
		WithArgs(user.ID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetUserBalance(ctx, user.ID)
	require.ErrorIs(t, err, e.ErrRepoUserBalanceNotFound)

	// arbitrary error
	mockPool.
		ExpectQuery(`SELECT userID, balance, withdrawn, updated_at`).
		WithArgs(user.ID).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.AdminShutdown})

	_, err = repo.GetUserBalance(ctx, user.ID)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, pgerrcode.AdminShutdown, pgErr.Code)

	err = mockPool.ExpectationsWereMet()
	require.NoError(t, err)
}
