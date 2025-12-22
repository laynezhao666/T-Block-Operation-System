package debug

import (
	"sync"
)

type disableHandler struct {
	sync.RWMutex
	handlers []func()
}

// AddHandler 添加处理函数
func (d *disableHandler) AddHandler(f func()) {
	if d == nil {
		return
	}

	d.Lock()
	defer d.Unlock()

	d.handlers = append(d.handlers, f)
}

// CallHandlers 调用处理函数
func (d *disableHandler) CallHandlers() {
	if d == nil {
		return
	}

	d.RLock()
	defer d.RUnlock()
	for _, f := range d.handlers {
		go func(h func()) {
			h()
		}(f)
	}
}
