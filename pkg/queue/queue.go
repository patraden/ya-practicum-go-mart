/*
Overview:

	Package queue provides a thread-safe, channel-based queue implementation
	that supports concurrent enqueue and dequeue operations with graceful shutdown.
	This queue is designed for high-concurrency environments where multiple
	goroutines may enqueue and dequeue items simultaneously.

Limitations:

	Does not support dynamic resizing of the queue capacity.
*/
package queue

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrClosed = errors.New("queue is closed")
	ErrFull   = errors.New("queue is full")
)

type Queue struct {
	ch           chan interface{}
	count        int32
	capacity     int
	closingMutex sync.RWMutex
	closed       uint32
}

func New(capacity int) *Queue {
	return &Queue{
		ch:           make(chan interface{}, capacity),
		capacity:     capacity,
		count:        0,
		closingMutex: sync.RWMutex{},
		closed:       0,
	}
}

func (q *Queue) Enqueue(item interface{}) error {
	q.closingMutex.RLock()
	defer q.closingMutex.RUnlock()

	if q.IsClosed() {
		return ErrClosed
	}

	select {
	case q.ch <- item:
		atomic.AddInt32(&q.count, 1)

		return nil
	default:
		return ErrFull
	}
}

func (q *Queue) Dequeue() interface{} {
	select {
	case item := <-q.ch:
		atomic.AddInt32(&q.count, -1)

		return item
	default:
		return nil
	}
}

func (q *Queue) Size() int32 {
	return atomic.LoadInt32(&q.count)
}

func (q *Queue) Cap() int {
	return cap(q.ch)
}

func (q *Queue) Close() {
	if atomic.CompareAndSwapUint32(&q.closed, 0, 1) {
		q.closingMutex.Lock()
		defer q.closingMutex.Unlock()
		close(q.ch)
	}
}

func (q *Queue) IsClosed() bool {
	return atomic.LoadUint32(&q.closed) == 1
}
