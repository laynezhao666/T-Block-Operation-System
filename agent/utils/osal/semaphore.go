package osal

import (
	"time"
)

// Semaphore 信号量
type Semaphore struct {
	ch chan int
	n  int
}

// NewSemaphore 创建信号量
func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		n = 1
	}
	return &Semaphore{
		n:  n,
		ch: make(chan int, n),
	}
}

// Post 释放信号量
func (s *Semaphore) Post() {
	if s == nil {
		return
	}
	if len(s.ch) == s.n {
		return
	}
	s.ch <- 1
}

// Wait 等待信号量
func (s *Semaphore) Wait(t time.Duration) bool {
	if s == nil {
		return false
	}
	select {
	case <-s.ch:
		return true
	case <-time.After(t):
		return false
	}
}
