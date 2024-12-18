package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
	"github.com/patraden/ya-practicum-go-mart/internal/app/usecase"
)

const (
	ContentType     = "Content-Type"
	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
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

func (h *UserHandler) UserBalance(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	balance, err := h.usecase.GetUserBalance(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if _, err := easyjson.MarshalToWriter(balance, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set(ContentType, ContentTypeJSON)
}

func (h *UserHandler) RegisterRoutes(router chi.Router, auth *jwtauth.JWTAuth) {
	router.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.RegisterUser)
		r.Post("/api/user/login", h.LoginUser)
	})

	router.Group(func(r chi.Router) {
		r.Use(middleware.Verifier(auth))
		r.Use(middleware.Authenticator())
		r.Get("/api/user/balance", h.UserBalance)
	})
}
