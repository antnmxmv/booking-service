package queue

import "time"

// DelayedQueue generic not persistent queue made with goroutines and time.Sleep
type DelayedQueue[T any] struct {
	queue chan T
}

func NewDelayedQueue[T any]() *DelayedQueue[T] {
	return &DelayedQueue[T]{
		queue: make(chan T),
	}
}

func (q *DelayedQueue[T]) SendMessage(message T, delay time.Duration) error {
	go func() {
		time.Sleep(delay)
		q.queue <- message
	}()
	return nil
}

func (q *DelayedQueue[T]) Subscribe() <-chan T {
	return q.queue
}
