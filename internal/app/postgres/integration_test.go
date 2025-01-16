package postgres_test

import (
	"context"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patraden/ya-practicum-go-mart/internal/app/config"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres"
	"github.com/patraden/ya-practicum-go-mart/internal/app/postgres/database"
)

const disabled = true

func setupOrderRepoDevDB(t *testing.T) (
	*postgres.UserRepository,
	*postgres.OrderRepository,
	*postgres.OrderTransactionsRepository,
	*model.User,
	*model.User,
) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.DatabaseURI = "postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	pgdb := database.NewDatabase(cfg.DatabaseURI, log)
	user1 := model.NewUserWithID("0a0118a4-d516-4b69-bdf7-e1fd49703798", "test_user1")
	user2 := model.NewUserWithID("0a0118a4-d516-4b69-bdf7-e1fd49703799", "test_user2")

	err := pgdb.Init(context.Background())
	require.NoError(t, err)

	err = pgdb.Ping(context.Background())
	require.NoError(t, err)

	repoUser := postgres.NewUserRepository(pgdb.ConnPool, log)
	repoOrder := postgres.NewOrderRepository(pgdb.ConnPool, log)
	repoTrx := postgres.NewOrderTransactionsRepository(pgdb.ConnPool, log)

	return repoUser, repoOrder, repoTrx, user1, user2
}

func TestCreateTestUsers(t *testing.T) {
	t.Parallel()

	if disabled {
		return
	}

	ctx := context.Background()
	userRepo, _, _, user1, user2 := setupOrderRepoDevDB(t)

	_, err := userRepo.CreateUser(ctx, user1)
	require.NoError(t, err)

	_, err = userRepo.CreateUser(ctx, user2)
	require.NoError(t, err)
}

func TestCreateTestAccrualOrders(t *testing.T) {
	t.Parallel()

	if disabled {
		return
	}

	ctx := context.Background()
	_, orderRepo, _, user1, user2 := setupOrderRepoDevDB(t)
	o1u1 := model.NewOrder(101, user1.ID)
	o2u1 := model.NewOrder(102, user1.ID)
	o1u2 := model.NewOrder(201, user2.ID)
	o2u2 := model.NewOrder(202, user2.ID)

	_, err := orderRepo.CreateOrder(ctx, o1u1)
	require.NoError(t, err)

	_, err = orderRepo.CreateOrder(ctx, o2u1)
	require.NoError(t, err)

	_, err = orderRepo.CreateOrder(ctx, o1u2)
	require.NoError(t, err)

	_, err = orderRepo.CreateOrder(ctx, o2u2)
	require.NoError(t, err)

	o1u1Status := model.
		NewOrderStatus(o1u1.ID, user1.ID, decimal.NewFromFloat(20.55)).
		ChangeStatus(model.StatusRegistered)

	o2u1Status := model.
		NewOrderStatus(o2u1.ID, user1.ID, decimal.NewFromFloat(25.466)).
		ChangeStatus(model.StatusProcessed)

	o1u2Status := model.
		NewOrderStatus(o1u2.ID, user2.ID, decimal.NewFromFloat(35.55)).
		ChangeStatus(model.StatusProcessed)

	o2u2Status := model.
		NewOrderStatus(o2u2.ID, user2.ID, decimal.NewFromFloat(32.4)).
		ChangeStatus(model.StatusProcessed)

	err = orderRepo.UpdateStatus(ctx, o1u1Status)
	require.NoError(t, err)

	err = orderRepo.UpdateStatus(ctx, o2u1Status)
	require.NoError(t, err)

	err = orderRepo.UpdateStatus(ctx, o1u2Status)
	require.NoError(t, err)

	err = orderRepo.UpdateStatus(ctx, o2u2Status)
	require.NoError(t, err)
}

func TestConcurrentWithdrawal(t *testing.T) {
	t.Parallel()

	if disabled {
		return
	}

	_, _, repoTrx, user1, user2 := setupOrderRepoDevDB(t)
	cent, twg := decimal.NewFromFloat(0.01), sync.WaitGroup{}

	for range 3000 {
		twg.Add(1)

		go func() {
			defer twg.Done()

			trx := model.NewWithdrawal(101, user1.ID, cent)
			_ = repoTrx.CreateWithdrawal(context.Background(), trx)
		}()

		twg.Add(1)

		go func() {
			defer twg.Done()

			trx := model.NewWithdrawal(102, user1.ID, cent)
			_ = repoTrx.CreateWithdrawal(context.Background(), trx)
		}()

		twg.Add(1)

		go func() {
			defer twg.Done()

			trx := model.NewWithdrawal(201, user2.ID, cent)
			_ = repoTrx.CreateWithdrawal(context.Background(), trx)
		}()

		twg.Add(1)

		go func() {
			defer twg.Done()

			trx := model.NewWithdrawal(202, user2.ID, cent)
			_ = repoTrx.CreateWithdrawal(context.Background(), trx)
		}()
	}

	twg.Wait()
}

func TestBalanceAndWithdrawals(t *testing.T) {
	t.Parallel()

	if disabled {
		return
	}

	_, _, repoTrx, user1, user2 := setupOrderRepoDevDB(t)

	balUser1, err := repoTrx.GetUserBalance(context.Background(), user1.ID)
	require.NoError(t, err)

	balUser2, err := repoTrx.GetUserBalance(context.Background(), user2.ID)
	require.NoError(t, err)

	assert.True(t, balUser1.Balance.Equal(decimal.NewFromFloat32(0.0)))
	assert.True(t, balUser1.Withdrawn.Equal(decimal.NewFromFloat32(25.47)))
	assert.True(t, balUser2.Balance.Equal(decimal.NewFromFloat32(7.95)))
	assert.True(t, balUser2.Withdrawn.Equal(decimal.NewFromFloat32(60.00)))
}
