package accrual

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/repository"
)

// event handler used by adapter's scheduled jobs.
type OrderEventHandler struct {
	client Client
	repo   repository.OrderRepository
}

func NewOrderEventHandler(
	client Client,
	repo repository.OrderRepository,
) *OrderEventHandler {
	return &OrderEventHandler{
		client: client,
		repo:   repo,
	}
}

func (eh *OrderEventHandler) Handle(ctx context.Context, event *Event) (*Event, error) {
	if !eh.client.IsAlive() {
		return event, e.ErrAdpaterAccrualNotAlive
	}

	orderStatus, err := eh.client.GetOrderStatus(ctx, event.orderStatus.ID)
	if err != nil {
		return event, e.Wrap("Adapter: accrual system error", err)
	}

	err = eh.repo.UpdateOrderStatus(ctx, orderStatus)
	if err != nil {
		// for now register internal failures only.
		event.AddFailure()

		return event, e.Wrap("Adapter: repo error", err)
	}

	newEvent := NewEvent(orderStatus)

	return newEvent, nil
}

func (eh *OrderEventHandler) OrderJobFn(
	ctx context.Context,
	qmgr *QueueManager,
	log *zerolog.Logger,
) JobFunction {
	return func(jobID int64, event *Event) {
		newEvent, err := eh.Handle(ctx, event)
		if err != nil {
			log.Error().Err(err).
				Int64("JobID", jobID).
				Int64("OrderID", newEvent.orderStatus.ID).
				Str("Status", string(newEvent.orderStatus.Status)).
				Msg("Adapter: event processing error")
		}

		// enqueue asap to minimize risks of lost events
		// when do we add failure to event?
		enqErr := qmgr.enqueue(newEvent)
		if errors.Is(enqErr, e.ErrAdapterMissedEvent) {
			log.Error().Err(enqErr).
				Int64("JobID", jobID).
				Int64("OrderID", newEvent.orderStatus.ID).
				Str("Status", string(newEvent.orderStatus.Status)).
				Msg("Adapter: event missed!")
		}

		// handle sleeps
		if errors.Is(err, e.ErrAdpaterAccrualNotAlive) {
			sleepWithContext(ctx, delayClientNotAlive, func() {
				log.Info().
					Int64("JobID", jobID).
					Msg("Adapter: job stopped while sleeping.")
			})

			return
		}

		var clientErr *e.AccrualClientError
		if errors.As(err, &clientErr) && clientErr.StatusCode == http.StatusTooManyRequests {
			sleepWithContext(ctx, clientErr.RetryAfter, func() {
				log.Info().
					Int64("JobID", jobID).
					Msg("Adapter: job stopped while sleeping.")
			})

			return
		}
	}
}

type DQLEventHandler struct{}

func NewDQLEventHandler() *DQLEventHandler { return &DQLEventHandler{} }

func (dlqh *DQLEventHandler) JobFn(
	_ context.Context,
	qmgr *QueueManager,
	log *zerolog.Logger,
) JobFunction {
	return func(jobID int64, event *Event) {
		if event.Failures > maxEventFailures {
			log.Error().
				Int64("JobID", jobID).
				Uint32("Failures", event.Failures).
				Int64("OrderID", event.orderStatus.ID).
				Str("Status", string(event.orderStatus.Status)).
				Msg("Adapter: event discarded!")

			return
		}

		err := qmgr.enqueue(event)
		if errors.Is(err, e.ErrAdapterMissedEvent) {
			log.Error().Err(err).
				Int64("JobID", jobID).
				Int64("OrderID", event.orderStatus.ID).
				Str("Status", string(event.orderStatus.Status)).
				Msg("Adapter: event missed!")
		}
	}
}

func sleepWithContext(ctx context.Context, delay time.Duration, cb func()) {
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		cb()

		return
	}
}
