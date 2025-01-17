package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
	"github.com/patraden/ya-practicum-go-mart/internal/app/usecase"
)

type OrdersHandler struct {
	usecase usecase.IOrderUseCase
	log     *zerolog.Logger
}

func NewOrdersHandler(usecase usecase.IOrderUseCase, log *zerolog.Logger) *OrdersHandler {
	return &OrdersHandler{
		usecase: usecase,
		log:     log,
	}
}

func (h *OrdersHandler) PostOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if r.Body == http.NoBody || err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = h.usecase.CreateOrder(r.Context(), claims.UserID, string(body))

	switch {
	case errors.Is(err, e.ErrUseCaseBadOrder):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, e.ErrRepoOrderExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, e.ErrRepoOrderUserExists):
		w.WriteHeader(http.StatusOK)
	case err != nil:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	orders, err := h.usecase.GetOrders(r.Context(), claims.UserID)
	if errors.Is(err, e.ErrRepoOrderNoOrders) || len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	orderResponse := make(dto.OrderStatusResponseBatch, len(orders))
	for i, order := range orders {
		orderResponse[i] = dto.OrderStatusResponse{
			ID:        strconv.FormatInt(order.ID, 10),
			Status:    order.Status,
			Accrual:   order.Accrual.InexactFloat64(),
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set(ContentType, ContentTypeJSON)

	if _, err = easyjson.MarshalToWriter(orderResponse, w); err != nil {
		h.log.Error().Err(err).
			Str("user_id", claims.UserID.String()).
			Msg(e.ErrJSONMarshal.Error())

		http.Error(w, e.ErrJSONMarshal.Error(), http.StatusInternalServerError)

		return
	}
}

func (h *OrdersHandler) RegisterRoutes(router chi.Router, auth *jwtauth.JWTAuth) {
	router.Group(func(r chi.Router) {
		r.Use(middleware.Verifier(auth))
		r.Use(middleware.Authenticator())
		r.Post("/api/user/orders", h.PostOrder)
		r.Get("/api/user/orders", h.GetOrders)
	})
}
