package accrual_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	"github.com/patraden/ya-practicum-go-mart/internal/app/integration/accrual"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/mock"
)

func setupOrderEventHandlerTest(t *testing.T) (
	*gomock.Controller,
	*mock.MockClient,
	*mock.MockOrderRepository,
	*accrual.OrderEventHandler,
	*dto.OrderStatus,
	*accrual.Event,
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockClient := mock.NewMockClient(ctrl)
	mockRepo := mock.NewMockOrderRepository(ctrl)
	handler := accrual.NewOrderEventHandler(mockClient, mockRepo)

	orderStatus := &dto.OrderStatus{ID: 1, Status: "NEW", Accrual: decimal.Zero}
	event := accrual.NewEvent(orderStatus)

	return ctrl, mockClient, mockRepo, handler, orderStatus, event
}

func TestOrderEventHandler_Handle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Client not alive", func(t *testing.T) {
		t.Parallel()

		ctrl, mockClient, _, handler, _, event := setupOrderEventHandlerTest(t)
		defer ctrl.Finish()

		mockClient.EXPECT().IsAlive().Return(false)

		_, err := handler.Handle(ctx, event)
		require.ErrorIs(t, err, e.ErrAdpaterAccrualNotAlive)
	})

	t.Run("Client GetOrderStatus error", func(t *testing.T) {
		t.Parallel()

		ctrl, mockClient, _, handler, _, event := setupOrderEventHandlerTest(t)
		defer ctrl.Finish()

		mockClient.EXPECT().IsAlive().Return(true)
		mockClient.EXPECT().GetOrderStatus(ctx, int64(1)).Return(nil, e.ErrTesting)

		_, err := handler.Handle(ctx, event)
		require.Contains(t, err.Error(), "accrual system error")
		require.ErrorIs(t, err, e.ErrTesting)
	})

	t.Run("Repository UpdateOrderStatus error", func(t *testing.T) {
		t.Parallel()

		ctrl, mockClient, mockRepo, handler, orderStatus, event := setupOrderEventHandlerTest(t)
		defer ctrl.Finish()

		mockClient.EXPECT().IsAlive().Return(true)
		mockClient.EXPECT().GetOrderStatus(ctx, int64(1)).Return(orderStatus, nil)
		mockRepo.EXPECT().UpdateOrderStatus(ctx, orderStatus).Return(e.ErrTesting)

		_, err := handler.Handle(ctx, event)
		require.Contains(t, err.Error(), "repo error")
		require.ErrorIs(t, err, e.ErrTesting)
		assert.Equal(t, uint32(1), event.Failures, "internal errors with repo register failure")
	})

	t.Run("Successful handling", func(t *testing.T) {
		t.Parallel()

		ctrl, mockClient, mockRepo, handler, orderStatus, event := setupOrderEventHandlerTest(t)
		defer ctrl.Finish()

		mockClient.EXPECT().IsAlive().Return(true)
		mockClient.EXPECT().GetOrderStatus(ctx, int64(1)).Return(orderStatus, nil)
		mockRepo.EXPECT().UpdateOrderStatus(ctx, orderStatus).Return(nil)

		_, err := handler.Handle(ctx, event)
		require.NoError(t, err)
	})
}

func getTestQueueManager(t *testing.T) *accrual.QueueManager {
	t.Helper()

	return accrual.NewQueueManager(
		accrual.NewEventQueue(1),
		accrual.NewEventQueue(1),
		accrual.NewEventQueue(1),
	)
}

func TestOrderEventHandler_OrderJobFn_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	qmanager := getTestQueueManager(t)
	ctrl, mockClient, mockRepo, handler, orderStatus, event := setupOrderEventHandlerTest(t)

	defer ctrl.Finish()

	orderStatus = orderStatus.ChangeStatus(model.StatusProcessing)

	mockClient.EXPECT().IsAlive().Return(true).Times(3)
	mockClient.EXPECT().GetOrderStatus(ctx, int64(1)).Return(orderStatus, nil).Times(3)
	mockRepo.EXPECT().UpdateOrderStatus(ctx, orderStatus).Return(nil).Times(3)

	orderStatusProcessed := &dto.OrderStatus{ID: 2, Status: "PROCESSED", Accrual: decimal.Zero}

	mockClient.EXPECT().IsAlive().Return(true).Times(2)
	mockClient.EXPECT().GetOrderStatus(ctx, int64(2)).Return(orderStatusProcessed, nil).Times(2)
	mockRepo.EXPECT().UpdateOrderStatus(ctx, orderStatusProcessed).Return(nil).Times(2)

	jobFn := handler.OrderJobFn(ctx, qmanager, log)

	jobFn(1, event)
	jobFn(2, event)
	jobFn(3, event)
	jobFn(4, accrual.NewEvent(orderStatusProcessed))
	jobFn(5, accrual.NewEvent(orderStatusProcessed))

	assert.Equal(t, int32(0), qmanager.QueueSize(accrual.EventTypeNew))
	assert.Equal(t, int32(1), qmanager.QueueSize(accrual.EventTypeInProgress), "one event in progress")
	assert.Equal(t, int32(1), qmanager.QueueSize(accrual.EventTypeDLQ), "one event in dql")

	stats := qmanager.GetStats()
	assert.Equal(t, uint64(0), stats.Submitted, "no submitted")
	assert.Equal(t, uint64(2), stats.Processed, "one processed")
	assert.Equal(t, uint64(2), stats.Failures, "two failures")
	assert.Equal(t, uint64(1), stats.Lost, "one lost")
}

func TestOrderEventHandler_OrderJobFn_RetryOnTooManyRequests(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	qmanager := getTestQueueManager(t)
	ctrl, mockClient, _, handler, _, event := setupOrderEventHandlerTest(t)

	defer ctrl.Finish()

	waitErr := &e.AccrualClientError{
		StatusCode: http.StatusTooManyRequests,
		RetryAfter: 10 * time.Second,
		Err:        e.ErrTesting,
	}

	mockClient.EXPECT().IsAlive().Return(true)
	mockClient.EXPECT().GetOrderStatus(ctx, int64(1)).Return(nil, waitErr)

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	jobFn := handler.OrderJobFn(ctx, qmanager, log)
	jobFn(1, event)

	assert.Equal(t, int32(1), qmanager.QueueSize(accrual.EventTypeNew), "event is back to queue")
	assert.Equal(t, int32(0), qmanager.QueueSize(accrual.EventTypeInProgress), "no event in progress")
	assert.Equal(t, int32(0), qmanager.QueueSize(accrual.EventTypeDLQ), "no event in DLQ")

	stats := qmanager.GetStats()
	assert.Equal(t, uint64(0), stats.Submitted, "no submitted")
	assert.Equal(t, uint64(0), stats.Processed, "no job processed")
	assert.Equal(t, uint64(0), stats.Failures, "no failures")
	assert.Equal(t, uint64(0), stats.Lost, "no lost")
}

func TestOrderEventHandler_OrderJobFn_NotAlive(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	log := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	qmanager := getTestQueueManager(t)
	ctrl, mockClient, _, handler, _, event := setupOrderEventHandlerTest(t)

	defer ctrl.Finish()

	mockClient.EXPECT().IsAlive().Return(false)

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	jobFn := handler.OrderJobFn(ctx, qmanager, log)
	jobFn(1, event)

	assert.Equal(t, int32(1), qmanager.QueueSize(accrual.EventTypeNew), "event is back to queue")
	assert.Equal(t, int32(0), qmanager.QueueSize(accrual.EventTypeInProgress), "no event in progress")
	assert.Equal(t, int32(0), qmanager.QueueSize(accrual.EventTypeDLQ), "no event in DLQ")

	stats := qmanager.GetStats()
	assert.Equal(t, uint64(0), stats.Submitted, "no submitted")
	assert.Equal(t, uint64(0), stats.Processed, "no job processed")
	assert.Equal(t, uint64(0), stats.Failures, "no failures")
	assert.Equal(t, uint64(0), stats.Lost, "no lost")
}
