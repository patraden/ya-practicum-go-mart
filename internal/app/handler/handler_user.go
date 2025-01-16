package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
	"github.com/patraden/ya-practicum-go-mart/internal/app/usecase"
)

type UserHandler struct {
	usecase      usecase.IUserUseCase
	tokenEncoder jwtauth.TokenEncoder
	log          *zerolog.Logger
}

func NewUserHandler(usecase usecase.IUserUseCase, encoder jwtauth.TokenEncoder, log *zerolog.Logger) *UserHandler {
	return &UserHandler{
		usecase:      usecase,
		tokenEncoder: encoder,
		log:          log,
	}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var creds dto.UserCredentials

	if err := easyjson.UnmarshalFromReader(r.Body, &creds); err != nil {
		h.log.Error().
			Err(err).
			Msg(e.ErrJSONUnmarshal.Error())

		http.Error(w, e.ErrJSONUnmarshal.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.usecase.CreateUser(r.Context(), &creds)
	if errors.Is(err, e.ErrRepoUserExists) {
		http.Error(w, err.Error(), http.StatusConflict)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if token, err := h.tokenEncoder(user.Username, user.ID); err != nil {
		h.log.Error().
			Err(err).
			Msg("failed to add auth token")
	} else {
		jwtauth.StoreTokenInCookie(w, token)
		jwtauth.StoreTokenInHeader(w, token)
	}
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds dto.UserCredentials

	if err := easyjson.UnmarshalFromReader(r.Body, &creds); err != nil {
		h.log.Error().
			Err(err).
			Msg(e.ErrJSONUnmarshal.Error())

		http.Error(w, e.ErrJSONUnmarshal.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.usecase.ValidateUser(r.Context(), &creds)
	if errors.Is(err, e.ErrRepoUserNotFound) || errors.Is(err, e.ErrRepoUserPassMismatch) {
		http.Error(w, err.Error(), http.StatusUnauthorized)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if token, err := h.tokenEncoder(user.Username, user.ID); err != nil {
		h.log.Error().
			Err(err).
			Msg("failed to add auth token")
	} else {
		jwtauth.StoreTokenInCookie(w, token)
		jwtauth.StoreTokenInHeader(w, token)
	}
}

func (h *UserHandler) RegisterRoutes(router chi.Router, _ *jwtauth.JWTAuth) {
	router.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.RegisterUser)
		r.Post("/api/user/login", h.LoginUser)
	})
}