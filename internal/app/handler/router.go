package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware"
)

func NewRouter(log *zerolog.Logger, handlers ...Handler) http.Handler {
	router := chi.NewRouter()

	// Apply common middleware to all routes
	router.Use(middleware.Recoverer())
	router.Use(middleware.StripSlashes())
	router.Use(middleware.Compress())
	router.Use(middleware.Logger(log))

	// Register handlers
	for _, h := range handlers {
		h.RegisterRoutes(router)
	}

	return router
}
