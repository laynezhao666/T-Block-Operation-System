// Package debug 调试
package debug

import (
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	enableTime = 300
)

var (
	enable    uint32 = 0
	notify           = make(chan int, 1)
	resetTime        = time.Duration(enableTime) * time.Second

	handler disableHandler
)

func init() {
	go reset()
}

// RegisterDisableHandler 注册禁用函数
func RegisterDisableHandler(f func()) {
	handler.AddHandler(f)
}

// SetEnable 设置启用调试
func SetEnable(t int) {
	if t <= 0 || t > 3600 {
		t = enableTime
	}
	log.Infof("enable debug, time: %vs", t)
	notify <- t
	atomic.StoreUint32(&enable, 1)
}

// IsEnable 判断是否启用调试
func IsEnable() bool {
	return atomic.LoadUint32(&enable) > 0
}

func callDisableHandlers() {
	handler.CallHandlers()
}

func disable() {
	if IsEnable() {
		log.Infof("disable debug...")
		callDisableHandlers()
	}
	atomic.StoreUint32(&enable, 0)
}

func reset() {
	for {
		select {
		case t, _ := <-notify:
			resetTime = time.Duration(t) * time.Second
			break
		case <-time.After(resetTime):
			disable()
			break
		}
	}
}
