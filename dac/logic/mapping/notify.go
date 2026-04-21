// Package mapping 提供门禁控制器的映射和GID同步功能。
package mapping

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/repo/dac"
)

// notifyTime GID同步通知的批量等待时间
const (
	notifyTime = time.Second * 10
)

// notifyFetchLoop 监听GID同步通知，批量获取并更新控制器GID。
// 使用定时器合并短时间内的多次通知，减少数据库操作频率。
func (w *worker) notifyFetchLoop(ctx context.Context) {
	codeMap := make(map[string]struct{}, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case code := <-w.notifyFetchChan:
			// 收集待同步的code
			codeMap[code] = struct{}{}
		case <-time.After(notifyTime):
			l := len(codeMap)
			if l == 0 {
				break
			}
			// 批量获取GID
			codes := make([]string, 0, l)
			for code := range codeMap {
				codes = append(codes, code)
			}
			codeGIDs, err := w.FetchGIDs(codes)
			if err != nil {
				break
			}
			// 更新数据库中的GID
			err = dac.GetRW().UpdateControllerAndDoorGIDsByCode(
				ctx, codeGIDs)
			if err != nil {
				config.Log.Warnf(
					"update code gid error: %v", err)
				break
			}

			codeMap = make(map[string]struct{})
		}
	}
}

// Notify 异步通知GID同步（非阻塞发送到通知通道）
func (w *worker) Notify(code string, mozuID string) {
	if config.C.IgnoreGID(mozuID) {
		return
	}
	go func() {
		w.notifyFetchChan <- code
	}()
}
