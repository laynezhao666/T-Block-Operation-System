package osal

import (
	"time"
)

// CronTimer 定时执行函数
type CronTimer struct {
	duration time.Duration
	f        func()
}

// NewCronTimer 创建定时器
func NewCronTimer(duration time.Duration, f func()) *CronTimer {
	return &CronTimer{
		duration: duration,
		f:        f,
	}
}

// Start 启动定时器
func (t *CronTimer) Start() {
	if t == nil {
		return
	}
	go t.loop()
}

func (t *CronTimer) loop() {
	for {
		select {
		case <-time.After(t.duration):
			t.f()
		}
	}
}
