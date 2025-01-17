package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"github.com/patraden/ya-practicum-go-mart/internal/app/config"
	"github.com/patraden/ya-practicum-go-mart/internal/app/handler"
	"github.com/patraden/ya-practicum-go-mart/internal/app/integration/accrual"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
	"github.com/patraden/ya-practicum-go-mart/internal/app/usecase"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	pgdb := database.NewDatabase(cfg.DatabaseURI, log)
	auth := jwtauth.NewJWTAuth(
		func(*jwt.Token) (interface{}, error) { return []byte(cfg.JWTSecret), nil },
		log,
	)

	if err := pgdb.Init(context.Background()); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to init db")
	}

	if err := pgdb.Ping(context.Background()); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to ping db")
	}

	accrualClient := accrual.NewClient(cfg.AccrualAddress)
	repoUser := postgres.NewUserRepository(pgdb.ConnPool, log)
	repoOrder := postgres.NewOrderRepository(pgdb.ConnPool, log)
	repoTransactions := postgres.NewOrderTransactionsRepository(pgdb.ConnPool, log)
	adapter := accrual.NewAdapter(accrualClient, repoOrder, cfg.QueueSize, cfg.AdapterJobsDelay, cfg.AdapterJobsDelay, log)
	userUseCase := usecase.NewUserUseCase(repoUser, log)
	orderUseCase := usecase.NewOrderUseCase(repoOrder, adapter, log)
	transactionsUseCase := usecase.NewTransactionsUseCase(repoTransactions, log)
	userHandler := handler.NewUserHandler(userUseCase, auth.Encoder(), log)
	orderHandler := handler.NewOrdersHandler(orderUseCase, log)
	transactionsHandler := handler.NewTransactionsHandler(transactionsUseCase, log)
	router := handler.NewRouter(log, auth, userHandler, orderHandler, transactionsHandler)

	HTTPListenAndServe(router, adapter, cfg, log)
}

func HTTPListenAndServe(router http.Handler, adapter *accrual.Adapter, cfg *config.Config, log *zerolog.Logger) {
	server := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	stopChan := make(chan os.Signal, 1)
	adapterCtx, adapterCancel := context.WithCancel(context.Background())

	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		adapter.Start(adapterCtx)

		log.Info().
			Str("RUN_ADDRESS", cfg.RunAddress).
			Str("ACCRUAL_SYSTEM_ADDRESS", cfg.AccrualAddress).
			Str("DATABASE_URI", cfg.DatabaseURI).
			Msg("Server started")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().
				Err(err).
				Msg("Server failed to start")
		}
	}()

	<-stopChan
	log.Info().
		Msg("Shutdown signal received")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeOut)
	defer cancel()

	adapterCancel()
	adapter.WaitStop(ctxShutdown)

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Error().
			Err(err).
			Msg("Server failed to shutdown gracefully")
	} else {
		log.Info().
			Msg("Server shutdown gracefully")
	}
}
