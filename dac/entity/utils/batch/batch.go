// Package batch 提供并发批量执行工具函数。
package batch

import (
	"context"
	"sync"
	"sync/atomic"
)

// Execute 并发执行多个任务，任一失败立即返回错误。
// 使用原子计数器检测错误，避免后续任务继续执行。
func Execute(ctx context.Context, args []interface{}, handler func(context.Context, interface{}) error) error {
	var wg sync.WaitGroup

	errorNum := uint32(0)

	errChan := make(chan error, len(args))
	finishChan := make(chan struct{}, 1)
	for i := range args {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if atomic.LoadUint32(&errorNum) > 0 {
				return
			}
			if err := handler(ctx, args[i]); err != nil {
				errChan <- err
				atomic.AddUint32(&errorNum, 1)
				return
			}
		}(i)
	}
	go func() {
		wg.Wait()
		finishChan <- struct{}{}
	}()

	select {
	case <-finishChan:
		return nil
	case err := <-errChan:
		return err
	}
}

// ExecuteAll 并发执行所有任务，等待全部完成后返回第一个错误。
// 与 Execute 不同，即使某个任务失败，其他任务仍会继续执行。
func ExecuteAll(ctx context.Context, args []interface{},
	handler func(context.Context, interface{}) error) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(args))
	for i := range args {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if err := handler(ctx, args[i]); err != nil {
				errChan <- err
				return
			}
		}(i)
	}
	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
