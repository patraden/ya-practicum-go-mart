package accrual

import (
	"errors"
	"sync/atomic"
	"time"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/pkg/queue"
)

type EventType string

const (
	EventTypeNew        EventType = "NEW"
	EventTypeInProgress EventType = "IN_PROGRESS"
	EventTypeDLQ        EventType = "DEAD_LETTER"
)

// order status event.
type Event struct {
	ts          int64
	orderStatus *model.OrderStatus
	Failures    uint32
}

func NewEvent(orderStatus *model.OrderStatus) *Event {
	return &Event{
		ts:          time.Now().UnixMicro(),
		orderStatus: orderStatus,
		Failures:    0,
	}
}

func (msg *Event) AddFailure() {
	atomic.AddUint32(&msg.Failures, 1)
}

type EventsStats struct {
	Submitted uint64 // events submitted to adapter
	Processed uint64 // events successfully processed by adapter
	Lost      uint64 // events lost or ignored by adapter
	Failures  uint64 // events failures, including repeatable failures on single event
}

func NewEventsStats() *EventsStats {
	return &EventsStats{
		Submitted: 0,
		Processed: 0,
		Lost:      0,
		Failures:  0,
	}
}

func (stats *EventsStats) IncrementSubmitted() {
	atomic.AddUint64(&stats.Submitted, 1)
}

func (stats *EventsStats) IncrementProcessed() {
	atomic.AddUint64(&stats.Processed, 1)
}

func (stats *EventsStats) IncrementFailures() {
	atomic.AddUint64(&stats.Failures, 1)
}

func (stats *EventsStats) IncrementLost() {
	atomic.AddUint64(&stats.Lost, 1)
}

type EventQueue struct {
	*queue.Queue
}

// Order event based thread-safe queue.
func NewEventQueue(capacity int) *EventQueue {
	return &EventQueue{
		Queue: queue.New(capacity),
	}
}

func (eq *EventQueue) Dequeue() *Event {
	item := eq.Queue.Dequeue()
	if item == nil {
		return nil
	}

	event, ok := item.(*Event)
	if !ok {
		return nil
	}

	return event
}

func (eq *EventQueue) Enqueue(event *Event) error {
	err := eq.Queue.Enqueue(event)
	if err != nil {
		return e.Wrap("event enqueue error", err)
	}

	return nil
}

// Manages adapter's queues, event routing and collects statistics.
type QueueManager struct {
	queueNew        *EventQueue
	queueInProgress *EventQueue
	queueDeadLetter *EventQueue
	stats           *EventsStats
}

func NewQueueManager(
	queueNew *EventQueue,
	queueInProgress *EventQueue,
	queueDeadLetter *EventQueue,
) *QueueManager {
	stats := NewEventsStats()

	return &QueueManager{
		queueNew:        queueNew,
		queueInProgress: queueInProgress,
		queueDeadLetter: queueDeadLetter,
		stats:           stats,
	}
}

func (qm *QueueManager) getQueue(eventType EventType) *EventQueue {
	switch eventType {
	case EventTypeNew:
		return qm.queueNew
	case EventTypeInProgress:
		return qm.queueInProgress
	case EventTypeDLQ:
		return qm.queueDeadLetter
	}

	return nil
}

func (qm *QueueManager) emitDLQ(event *Event) error {
	if err := qm.queueDeadLetter.Enqueue(event); err != nil {
		qm.stats.IncrementFailures()
		qm.stats.IncrementLost()

		return e.ErrAdapterMissedEvent
	}

	qm.stats.IncrementFailures()

	return e.ErrAdpaterDLQEvent
}

func (qm *QueueManager) enqueue(event *Event) error {
	switch event.orderStatus.Status {
	case model.StatusNew:
		if err := qm.queueNew.Enqueue(event); err == nil {
			return nil
		}

		return qm.emitDLQ(event)
	case model.StatusProcessing, model.StatusRegistered:
		if err := qm.queueInProgress.Enqueue(event); err == nil {
			return nil
		}

		return qm.emitDLQ(event)
	case model.StatusInvalid, model.StatusProcessed:
		qm.stats.IncrementProcessed()

		return nil
	}

	return nil
}

func (qm *QueueManager) submitOrder(orderStatus *model.OrderStatus) bool {
	event := NewEvent(orderStatus)
	err := qm.enqueue(event)

	if errors.Is(err, e.ErrAdapterMissedEvent) {
		return false
	}

	qm.stats.IncrementSubmitted()

	return true
}

func (qm *QueueManager) GetStats() EventsStats {
	return EventsStats{
		Submitted: atomic.LoadUint64(&qm.stats.Submitted),
		Processed: atomic.LoadUint64(&qm.stats.Processed),
		Lost:      atomic.LoadUint64(&qm.stats.Lost),
		Failures:  atomic.LoadUint64(&qm.stats.Failures),
	}
}

func (qm *QueueManager) QueueSize(eventType EventType) int32 {
	q := qm.getQueue(eventType)
	if q == nil {
		return 0
	}

	return q.Size()
}
