package handler

import (
	"github.com/go-chi/chi/v5"

	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
)

const (
	ContentType     = "Content-Type"
	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
)

type Handler interface {
	RegisterRoutes(router chi.Router, auth *jwtauth.JWTAuth)
}
