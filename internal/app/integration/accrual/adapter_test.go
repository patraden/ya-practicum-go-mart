package accrual_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/integration/accrual"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/mock"
)

func setupAccrualAdapterTest(t *testing.T, queueCapacity int) (
	*gomock.Controller,
	*mock.MockIClient,
	*mock.MockOrderRepository,
	*accrual.Adapter,
) {
	t.Helper()

	jobDelayDLQ := 100 * time.Millisecond
	jobDelayInProgress := 100 * time.Millisecond
	ctrl := gomock.NewController(t)
	mockClient := mock.NewMockIClient(ctrl)
	mockRepo := mock.NewMockOrderRepository(ctrl)
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	adapter := accrual.NewAdapter(mockClient, mockRepo, queueCapacity, jobDelayInProgress, jobDelayDLQ, log)

	return ctrl, mockClient, mockRepo, adapter
}

func TestAccrrualAdapter_Stress(t *testing.T) {
	t.Parallel()

	queueCapacity := 1000
	totalOrders := 10000
	ctrl, mockClient, mockRepo, adapter := setupAccrualAdapterTest(t, queueCapacity)
	ctx, cancel := context.WithCancel(context.Background())

	defer ctrl.Finish()

	mockClient.EXPECT().IsAlive().Return(true).AnyTimes()
	mockRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(nil).AnyTimes()

	mockClient.EXPECT().GetOrderStatus(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, orderID int64, userID uuid.UUID) (*model.OrderStatus, error) {
			return &model.OrderStatus{
				ID:      orderID,
				UserID:  userID,
				Status:  model.StatusProcessing,
				Accrual: decimal.Zero,
			}, nil
		}).Times(totalOrders)

	mockClient.EXPECT().GetOrderStatus(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, orderID int64, userID uuid.UUID) (*model.OrderStatus, error) {
			return &model.OrderStatus{
				ID:      orderID,
				UserID:  userID,
				Status:  model.StatusProcessed,
				Accrual: decimal.Zero,
			}, nil
		}).Times(totalOrders)

	go func() {
		for i := range totalOrders {
			orderStatus := &model.OrderStatus{
				ID:      int64(i),
				UserID:  uuid.Nil,
				Status:  model.StatusNew,
				Accrual: decimal.Zero,
			}
			adapter.SubmitOrder(orderStatus)
		}
	}()

	adapter.Start(ctx)
	time.Sleep(5 * time.Second)

	cancel()
	adapter.WaitStop(context.Background())

	stats := adapter.GetStats()
	assert.Equal(t, uint64(totalOrders), stats.Submitted, "all successfully submitted")
	assert.Equal(t, uint64(totalOrders), stats.Processed, "all successfully processed")
	assert.Positive(t, stats.Failures, "failures due to low queue capacity")
	assert.Equal(t, uint64(0), stats.Lost, "nothing is lost")
}
