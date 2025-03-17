package common

import (
	"context"
	"sync"
)

// WorkerPool is a pool of workers that process items concurrently.
type WorkerPool[T any] struct {
	lock     sync.Mutex
	wg       sync.WaitGroup
	start    bool
	ch       chan T
	size     int
	errors   []error
	procFunc func(context.Context, T) error
	cancel   func()
}

// Add adds an item to the pool.
func (wp *WorkerPool[T]) Add(item T) {
	wp.ch <- item
}

// Errors returns the errors encountered during processing.
func (wp *WorkerPool[T]) Errors() []error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	return wp.errors
}

// Start starts the worker pool.
func (wp *WorkerPool[T]) Start(ctx context.Context) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if wp.start {
		return
	}
	wp.start = true

	var extCtx context.Context
	extCtx, wp.cancel = context.WithCancel(context.Background())
	for i := 0; i < wp.size; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for {
				select {
				case <-extCtx.Done():
					return
				case item := <-wp.ch:
					err := wp.procFunc(ctx, item)
					if err != nil {
						wp.lock.Lock()
						wp.errors = append(wp.errors, err)
						wp.lock.Unlock()
					}
				}
			}
		}()
	}
}

// Join waits for all workers to finish.
func (wp *WorkerPool[T]) Join() {
	wp.lock.Lock()
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.lock.Unlock()
	wp.wg.Wait()
	wp.start = false
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool[T any](procFunc func(context.Context, T) error, size int) *WorkerPool[T] {
	return &WorkerPool[T]{
		ch:       make(chan T),
		procFunc: procFunc,
		size:     size,
	}
}
