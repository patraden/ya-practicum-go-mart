package accrual

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/repository"
)

// events might be limited by size of 256b
// current dlq capacity will be around 256Mb.
const (
	delayClientNotAlive = 10 * time.Second
	jobDelayNew         = 100 * time.Millisecond
	jobConcurency       = 2
	maxEventFailures    = 5
	dlqCapcity          = 1_000_000
)

type JobFunction func(jobID int64, event *Event)

type Adapter struct {
	qmgr               *QueueManager
	orderHandler       *OrderEventHandler
	dlqHandler         *DQLEventHandler
	jobDelayDLQ        time.Duration
	jobDelayInProgress time.Duration
	log                *zerolog.Logger
	wg                 sync.WaitGroup
}

func NewAdapter(
	client IClient,
	repo repository.OrderRepository,
	queueCapacity int,
	jobDelayDLQ time.Duration,
	jobDelayInProgress time.Duration,
	log *zerolog.Logger,
) *Adapter {
	qmanager := NewQueueManager(
		NewEventQueue(queueCapacity),
		NewEventQueue(queueCapacity),
		NewEventQueue(dlqCapcity),
	)
	orderHandler := NewOrderEventHandler(client, repo)
	dlqHandler := NewDQLEventHandler()

	return &Adapter{
		qmgr:               qmanager,
		orderHandler:       orderHandler,
		dlqHandler:         dlqHandler,
		log:                log,
		wg:                 sync.WaitGroup{},
		jobDelayInProgress: jobDelayInProgress,
		jobDelayDLQ:        jobDelayDLQ,
	}
}

func (adp *Adapter) SubmitOrder(orderStatus *model.OrderStatus) bool {
	return adp.qmgr.submitOrder(orderStatus)
}

func (adp *Adapter) scheduler(ctx context.Context, source EventType, jobFn JobFunction, delay time.Duration) {
	defer adp.wg.Done()

	adp.log.Info().
		Str("EventType", string(source)).
		Msg("Adapter: scheduler started.")

	queue := adp.qmgr.getQueue(source)
	if queue == nil {
		adp.log.Error().
			Str("EventType", string(source)).
			Msg("Adapter: invalid event type, stopping scheduler...")

		return
	}

	for {
		select {
		case <-ctx.Done():
			adp.log.Info().
				Str("EventType", string(source)).
				Msg("Adapter: scheduler stopped gracefully.")

			return
		default:
			// potentially make upper limit here for job size
			if jobSize := queue.Size(); jobSize > 0 {
				adp.job(queue, source, int(jobSize), jobConcurency, jobFn)
			}

			// rest between job runs
			sleepWithContext(ctx, delay, func() {
				adp.log.Info().
					Str("EventType", string(source)).
					Msg("Adapter: scheduler stopped while sleeping.")
			})
		}
	}
}

func (adp *Adapter) job(queue *EventQueue, source EventType, size int, concurrency int, jobFn JobFunction) {
	tasks := make(chan *Event, size)
	start := time.Now()
	jobID := start.UnixMicro()

	adp.log.Info().
		Int64("JobID", jobID).
		Str("EventType", string(source)).
		Int("Size", size).
		Int("Workers", concurrency).
		Msg("Adapter: job started")

	// spawning workers
	var jwg sync.WaitGroup
	for range concurrency {
		jwg.Add(1)

		go func() {
			defer jwg.Done()

			for event := range tasks {
				jobFn(jobID, event)
			}
		}()
	}

	// sending events to be processed
	for range size {
		event := queue.Dequeue()
		if event == nil {
			continue
		}

		select {
		case tasks <- event:
		default: // just in case make it non-blocking
		}
	}

	close(tasks)
	jwg.Wait()

	adp.log.Info().
		Int64("JobID", jobID).
		Str("EventType", string(source)).
		Int("Size", size).
		Int("Workers", concurrency).
		Int64("Duration(ms)", time.Since(start).Milliseconds()).
		Msg("Adapter: job finished")
}

func (adp *Adapter) GetStats() *EventsStats {
	return adp.qmgr.stats
}

func (adp *Adapter) logStats() {
	stats := adp.GetStats()

	adp.log.Info().
		Uint64("Submitted", stats.Submitted).
		Uint64("Processed", stats.Processed).
		Uint64("Failures", stats.Failures).
		Uint64("Lost", stats.Lost).
		Int32("RemainNew", adp.qmgr.QueueSize(EventTypeNew)).
		Int32("RemainInProgress", adp.qmgr.QueueSize(EventTypeInProgress)).
		Int32("RemainDLQ", adp.qmgr.QueueSize(EventTypeDLQ)).
		Msg("Adapter: stats summary")
}

func (adp *Adapter) Start(ctx context.Context) {
	jobFn := adp.orderHandler.OrderJobFn(ctx, adp.qmgr, adp.log)
	dlqFn := adp.dlqHandler.JobFn(ctx, adp.qmgr, adp.log)

	adp.wg.Add(1)
	go adp.scheduler(ctx, EventTypeNew, jobFn, jobDelayNew)

	adp.wg.Add(1)
	go adp.scheduler(ctx, EventTypeInProgress, jobFn, adp.jobDelayInProgress)

	adp.wg.Add(1)
	go adp.scheduler(ctx, EventTypeDLQ, dlqFn, adp.jobDelayDLQ)
}

func (adp *Adapter) WaitStop(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		adp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		adp.log.Info().Msg("Adapter: stopped gracefully")
	case <-ctx.Done():
		adp.log.Error().Msg("Adapter: shutdown timed out")
	}

	adp.logStats()
}
