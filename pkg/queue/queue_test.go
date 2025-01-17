package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patraden/ya-practicum-go-mart/pkg/queue"
)

func TestQueueNew(t *testing.T) {
	t.Parallel()

	chq := queue.New(10)

	assert.NotNil(t, chq)
	assert.Equal(t, int32(0), chq.Size())
	assert.Equal(t, 10, chq.Cap())
	assert.False(t, chq.IsClosed())
}

func TestQueueEnqueueAndDequeue(t *testing.T) {
	t.Parallel()

	chq := queue.New(5)

	err := chq.Enqueue("item1")
	require.NoError(t, err)

	err = chq.Enqueue("item2")
	require.NoError(t, err)

	assert.Equal(t, int32(2), chq.Size())

	item := chq.Dequeue()
	assert.Equal(t, "item1", item)
	assert.Equal(t, int32(1), chq.Size())

	item = chq.Dequeue()
	assert.Equal(t, "item2", item)
	assert.Equal(t, int32(0), chq.Size())

	item = chq.Dequeue()
	assert.Nil(t, item) // Dequeue from an empty queue
}

func TestQueueEnqueueFull(t *testing.T) {
	t.Parallel()

	chq := queue.New(2)

	require.NoError(t, chq.Enqueue("item1"))
	require.NoError(t, chq.Enqueue("item2"))

	err := chq.Enqueue("item3")
	require.ErrorIs(t, err, queue.ErrFull)

	assert.Equal(t, int32(2), chq.Size())
}

func TestQueueClosure(t *testing.T) {
	t.Parallel()

	chq := queue.New(5)

	require.NoError(t, chq.Enqueue("item1"))
	require.NoError(t, chq.Enqueue("item2"))

	chq.Close()

	assert.True(t, chq.IsClosed())

	err := chq.Enqueue("item3")
	require.ErrorIs(t, err, queue.ErrClosed)

	item := chq.Dequeue()
	assert.Equal(t, "item1", item)

	item = chq.Dequeue()
	assert.Equal(t, "item2", item)

	item = chq.Dequeue()
	assert.Nil(t, item) // Dequeue from an empty and closed queue
}

func TestQueueConcurrentEnqueue(t *testing.T) {
	t.Parallel()

	var wgr sync.WaitGroup
	totalItems := 10000
	chq := queue.New(totalItems)

	for i := range totalItems {
		wgr.Add(1)

		go func(i int) {
			defer wgr.Done()

			err := chq.Enqueue(i)
			assert.NoError(t, err)
		}(i)
	}

	wgr.Wait()
	assert.Equal(t, int32(totalItems), chq.Size())
}

func TestQueueConcurrentDequeue(t *testing.T) {
	t.Parallel()

	var wgr sync.WaitGroup
	totalItems := 10000
	chq := queue.New(totalItems)

	// Fill the queue
	for i := range totalItems {
		require.NoError(t, chq.Enqueue(i))
	}

	assert.Equal(t, int32(totalItems), chq.Size())

	// Concurrently dequeue items
	for range totalItems {
		wgr.Add(1)

		go func() {
			defer wgr.Done()

			item := chq.Dequeue()
			assert.NotNil(t, item)
		}()
	}

	wgr.Wait()
	assert.Equal(t, int32(0), chq.Size())
}

func TestQueueConcurrentEnqueueDequeue(t *testing.T) {
	t.Parallel()

	var wgr sync.WaitGroup
	totalItems := 2000
	chq := queue.New(1000)

	for i := range totalItems {
		wgr.Add(1)

		go func(i int) {
			defer wgr.Done()

			if i%2 == 0 {
				err := chq.Enqueue(i)
				assert.NoError(t, err)
			} else {
				chq.Dequeue()
			}
		}(i)
	}

	wgr.Wait()
	assert.GreaterOrEqual(t, chq.Size(), int32(0))
}

func TestQueueCloseWhileEnqueuing(t *testing.T) {
	t.Parallel()

	var wgr sync.WaitGroup
	chq := queue.New(100)

	// Start multiple goroutines to enqueue
	for i := range 50 {
		wgr.Add(1)

		go func(i int) {
			defer wgr.Done()

			if err := chq.Enqueue(i); err != nil {
				assert.ErrorIs(t, err, queue.ErrClosed)
			}
		}(i)
	}

	// Close the queue while enqueuing
	wgr.Add(1)

	go func() {
		defer wgr.Done()
		time.Sleep(3 * time.Millisecond)
		chq.Close()
	}()

	wgr.Wait()
	assert.True(t, chq.IsClosed())
}
