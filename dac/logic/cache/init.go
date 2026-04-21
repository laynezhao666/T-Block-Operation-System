// Package cache 提供门禁系统的运行时数据缓存。
package cache

import (
	"context"
)

// Init 初始化缓存模块并启动后台刷新协程
func Init(ctx context.Context) error {
	var err error

	Get().Start(ctx)

	return err
}
