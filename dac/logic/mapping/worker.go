// Package mapping 提供采集编码到全局唯一标识符(GID)的映射管理。
package mapping

import (
	"sync"

	"dac/entity/model/db"
	"dac/entity/model/rt"
)

// www 全局映射Worker单例
var (
	www = newWorker()
)

// newWorker 创建新的映射Worker实例
func newWorker() *worker {
	w := new(worker)
	w.codeGIDMap = make(rt.CodeGIDMapType)
	w.notifyFetchChan = make(chan string, 200)

	return w
}

// worker 采集编码到GID的映射管理器
type worker struct {
	sync.RWMutex
	codeGIDMap      rt.CodeGIDMapType // 编码到GID的映射表
	notifyFetchChan chan string       // 通知拉取GID的通道
}

// GetWorker 获取全局映射Worker实例
func GetWorker() *worker {
	return www
}

// GetGID 根据采集编码获取对应的GID
func (w *worker) GetGID(code string) (db.GIDType, bool) {
	w.RLock()
	defer w.RUnlock()
	gid, ok := w.codeGIDMap[code]
	return gid, ok
}

// setGIDs 批量设置编码到GID的映射
func (w *worker) setGIDs(data rt.CodeGIDMapType) {
	w.Lock()
	defer w.Unlock()

	for k, v := range data {
		if len(v) == 0 {
			continue
		}
		w.codeGIDMap[k] = v
	}
}
