package handler

import (
	"errors"
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

type TransactionsHandler struct {
	usecase usecase.ITransactionsUseCase
	log     *zerolog.Logger
}

func NewTransactionsHandler(usecase usecase.ITransactionsUseCase, log *zerolog.Logger) *TransactionsHandler {
	return &TransactionsHandler{
		usecase: usecase,
		log:     log,
	}
}

func (h *TransactionsHandler) UserBalance(w http.ResponseWriter, r *http.Request) {
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

	balancResp := &dto.UserBalanceResponse{
		Balance:   balance.Balance.InexactFloat64(),
		Withdrawn: balance.Withdrawn.InexactFloat64(),
	}

	w.Header().Set(ContentType, ContentTypeJSON)

	if _, err := easyjson.MarshalToWriter(balancResp, w); err != nil {
		h.log.Error().Err(err).
			Str("user_id", claims.UserID.String()).
			Msg(e.ErrJSONMarshal.Error())

		http.Error(w, e.ErrJSONMarshal.Error(), http.StatusInternalServerError)

		return
	}
}

func (h *TransactionsHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	wdrl := &dto.Withdrawal{}
	if err := easyjson.UnmarshalFromReader(r.Body, wdrl); err != nil {
		h.log.Error().Err(err).
			Str("user_id", claims.UserID.String()).
			Msg(e.ErrJSONUnmarshal.Error())

		http.Error(w, e.ErrJSONUnmarshal.Error(), http.StatusInternalServerError)

		return
	}

	wdrl.SetUser(claims.UserID)
	err = h.usecase.CreateWithdrawal(r.Context(), wdrl)

	if errors.Is(err, e.ErrUseCaseBadOrder) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

		return
	}

	if errors.Is(err, e.ErrRepoOrderNoFunds) {
		http.Error(w, err.Error(), http.StatusPaymentRequired)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (h *TransactionsHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	trxs, err := h.usecase.GetUserWithdrawals(r.Context(), claims.UserID)
	if errors.Is(err, e.ErrRepoOrderNoWithdrawals) || len(trxs) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	wResponse := make(dto.WithdrawalResponseBatch, len(trxs))
	for i, trx := range trxs {
		wResponse[i] = dto.WithdrawalResponse{
			OrderID:   strconv.FormatInt(trx.OrderID, 10),
			Amount:    trx.Amount.InexactFloat64(),
			CreatedAt: trx.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set(ContentType, ContentTypeJSON)

	if _, err = easyjson.MarshalToWriter(wResponse, w); err != nil {
		h.log.Error().Err(err).
			Str("user_id", claims.UserID.String()).
			Msg(e.ErrJSONMarshal.Error())

		http.Error(w, e.ErrJSONMarshal.Error(), http.StatusInternalServerError)

		return
	}
}

func (h *TransactionsHandler) RegisterRoutes(router chi.Router, auth *jwtauth.JWTAuth) {
	router.Group(func(r chi.Router) {
		r.Use(middleware.Verifier(auth))
		r.Use(middleware.Authenticator())
		r.Get("/api/user/balance", h.UserBalance)
		r.Post("/api/user/balance/withdraw", h.CreateWithdrawal)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})
}
