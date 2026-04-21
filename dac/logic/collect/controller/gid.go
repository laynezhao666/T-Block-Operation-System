// Package controller 提供门禁控制器的采集管理功能。
package controller

import (
	"context"
	"time"

	"dac/entity/model/db"
	"dac/logic/mapping"
)

// GetGID 获取指定门编号对应的全局唯一标识符
func (c *Controller) GetGID(door int) (db.GIDType, bool) {
	c.doorGIDMutex.RLock()
	defer c.doorGIDMutex.RUnlock()

	gid, ok := c.doorGIDs[door]
	return gid, ok
}

// refreshGID 刷新所有门的GID映射
func (c *Controller) refreshGID(ctx context.Context) {
	for i := range c.record.Doors {
		d := &c.record.Doors[i]

		code := d.GetCollectCode(c.GetCollectCode())

		n := d.Number
		gid, ok := mapping.GetWorker().GetGID(code)
		if !ok {
			mapping.GetWorker().Notify(code, c.MozuID())
			continue
		}

		c.doorGIDMutex.Lock()
		c.doorGIDs[n] = gid
		c.doorGIDMutex.Unlock()
	}
}

// refreshGIDLoop 后台循环刷新GID映射
func (c *Controller) refreshGIDLoop(ctx context.Context) {
	for {
		c.refreshGID(ctx)
		select {
		case <-ctx.Done():
			c.Infof("refreshGIDLoop exit")
			return
		case <-c.refreshChan:
			break
		case <-time.After(time.Hour):
			break
		}
	}
}

// NotifyRefreshGID 通知刷新GID映射
func (c *Controller) NotifyRefreshGID() {
	go func() {
		c.refreshChan <- struct{}{}
	}()
}
