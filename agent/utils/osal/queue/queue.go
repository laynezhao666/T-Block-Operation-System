// Package queue 队列
package queue

import (
	"errors"
	"time"

	queuelib "github.com/Workiva/go-datastructures/queue"
)

var (
	errNil = errors.New("nil receiver")
)

// T interface
type T interface{}

// ThreadQueue thread queue
type ThreadQueue struct {
	container *queuelib.Queue
}

// NewThreadQueue new thread queue
func NewThreadQueue() *ThreadQueue {
	return &ThreadQueue{container: queuelib.New(50)}
}

// Push push
func (q *ThreadQueue) Push(t T) error {
	if q == nil {
		return errNil
	}
	return q.container.Put(t)
}

// Pop pop
func (q *ThreadQueue) Pop() (T, bool) {
	if q == nil || q.container == nil {
		return nil, false
	}
	t, err := q.container.Poll(1, time.Nanosecond)
	if err != nil {
		return nil, false
	}
	return t[0], true
}

// Clear clear
func (q *ThreadQueue) Clear() {
	if q == nil {
		return
	}
	q.container.Dispose()
}
