package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/repository"
)

type UserUseCase struct {
	repo repository.UserRepository
	log  *zerolog.Logger
}

func NewUserUseCase(repo repository.UserRepository, log *zerolog.Logger) *UserUseCase {
	return &UserUseCase{
		repo: repo,
		log:  log,
	}
}

func (u *UserUseCase) CreateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error) {
	user := model.NewUser(creds.Username)
	if err := user.SetPassword(creds.Password); err != nil {
		u.log.
			Error().Err(err).
			Str("username", creds.Username).
			Msg("user password generation error")

		return nil, e.ErrUseCaseInternal
	}

	createdUser, err := u.repo.CreateUser(ctx, user)

	if errors.Is(err, e.ErrRepoUserExists) {
		return nil, e.ErrRepoUserExists
	}

	if err != nil {
		u.log.
			Error().Err(err).
			Str("username", creds.Username).
			Msg("internal error")

		return nil, e.ErrUseCaseInternal
	}

	return createdUser, nil
}

func (u *UserUseCase) ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error) {
	user, err := u.repo.ValidateUser(ctx, creds.Username, creds.Password)

	if errors.Is(err, e.ErrRepoUserNotFound) {
		return nil, e.ErrRepoUserNotFound
	}

	if errors.Is(err, e.ErrRepoUserPassMismatch) {
		return nil, e.ErrRepoUserPassMismatch
	}

	if err != nil {
		u.log.
			Error().Err(err).
			Str("username", creds.Username).
			Msg("internal error")

		return nil, e.ErrUseCaseInternal
	}

	return user, nil
}

func (u *UserUseCase) GetUserBalance(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error) {
	userBalance, err := u.repo.GetUserBalance(ctx, userID)

	if errors.Is(err, e.ErrRepoUserBalanceNotFound) {
		return nil, e.ErrRepoUserBalanceNotFound
	}

	if err != nil {
		u.log.
			Error().Err(err).
			Str("user_id", userID.String()).
			Msg("internal error")

		return nil, e.ErrUseCaseInternal
	}

	return userBalance, nil
}
