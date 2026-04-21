package utils

import (
	"sync"
	"sync/atomic"
)

// SyncCall 同步顺序执行多个函数，任一失败立即返回错误。
func SyncCall(fs ...func() error) error {
	var err error
	for _, f := range fs {
		if err = f(); err != nil {
			return err
		}
	}
	return nil
}

// AsyncCall 异步并发执行多个函数，任一失败立即返回错误。
// 使用原子计数器检测错误，避免后续协程继续执行。
func AsyncCall(fs ...func() error) error {
	var wg sync.WaitGroup

	var errNum uint32 = 0

	errs := make(chan error, 1)
	finished := make(chan struct{}, 1)

	for _, f := range fs {
		wg.Add(1)
		go func(f func() error) {
			defer wg.Done()

			if atomic.LoadUint32(&errNum) > 0 {
				return
			}

			if err := f(); err != nil {
				atomic.AddUint32(&errNum, 1)
				errs <- err
			}
		}(f)
	}

	go func() {
		wg.Wait()
		finished <- struct{}{}
	}()

	select {
	case <-finished:
		return nil
	case err := <-errs:
		return err
	}
}
