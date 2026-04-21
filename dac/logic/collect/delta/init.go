// Package delta 提供测点数据变化检测和增量写入Redis的功能。
package delta

import (
	"context"

	"dac/logic/collect/rtdb"
)

// Init 初始化增量数据模块，注册测点更新回调并启动工作协程
func Init(ctx context.Context) {
	rtdb.RegisterPointsUpdatedCallback(callback, nil)

	w.Start(ctx)
}

// UnInit 清理增量数据模块，取消测点更新回调
func UnInit() {
	rtdb.UnRegisterPointsUpdatedCallback(callback)
}
