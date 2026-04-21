package push

import (
	"context"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/logic/dlm"

	"dac/entity/utils/ttime"
	pb "dac/repo/pb/tcommon_point_data"
)

// changeTime 变化点位上报间隔
const (
	changeTime = time.Second
)

// changedCacheDataType 变化点位缓存数据类型，key为点位ID
type changedCacheDataType map[string]rt.Point

// changedPointsType 变化点位缓存，支持并发安全的读写操作
type changedPointsType struct {
	data  changedCacheDataType
	mutex sync.RWMutex
}

// newChangedCache 创建新的变化点位缓存实例
func newChangedCache() *changedPointsType {
	return &changedPointsType{
		data: make(changedCacheDataType, 10),
	}
}

// Len 返回缓存中变化点位的数量
func (c *changedPointsType) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

// SetPoints 将变化的点位写入缓存
func (c *changedPointsType) SetPoints(points []*rt.Point) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, p := range points {
		c.data[p.ID] = *p
	}
}

// MovePoints 取出并清空缓存中的所有变化点位
func (c *changedPointsType) MovePoints() changedCacheDataType {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	r := c.data
	c.data = make(changedCacheDataType, 10)
	return r
}

// getChangedPoints 获取指定控制器的变化点位缓存，不存在则创建
func (w *worker) getChangedPoints(controllerID db.IDType) *changedPointsType {
	w.controllerChangedPointsMutex.RLock()
	points, ok := w.controllerChangedPoints[controllerID]
	if ok {
		w.controllerChangedPointsMutex.RUnlock()
		return points
	}
	w.controllerChangedPointsMutex.RUnlock()

	w.controllerChangedPointsMutex.Lock()
	defer w.controllerChangedPointsMutex.Unlock()

	if points, ok = w.controllerChangedPoints[controllerID]; !ok {
		points = newChangedCache()
		w.controllerChangedPoints[controllerID] = points
	}
	return points
}

// changedCallback 点位变化回调函数，过滤首次访问和值变化的点位
func (w *worker) changedCallback(points rt.Points, arg interface{}) interface{} {
	controllerID, ok := arg.(db.IDType)
	if !ok {
		return nil
	}

	changedPoints := make([]*rt.Point, 0, len(points))
	for i := range points {
		p := &points[i]
		if !p.IsValueChanged {
			if !w.isPointFirstAccess(p.ID) {
				continue
			}
			w.setPointHasAccessed(p.ID)
		}

		changedPoints = append(changedPoints, p)
	}

	w.getChangedPoints(controllerID).SetPoints(changedPoints)

	return nil
}

// reportChangedPointsLoop 循环上报变化点位数据
func (w *worker) reportChangedPointsLoop(ctx context.Context) {
	config.Log.Infof("start report changed points")
	for {
		if !dlm.GetWorker().HasLock() {
			time.Sleep(time.Minute)
			continue
		}

		w.reportChangedPoints(ctx)
		select {
		case <-time.After(changeTime):
			break
		case <-ctx.Done():
			config.Log.Infof("stop report changed points")
			return
		}
	}
}

// getAllChangedPoints 并发收集所有控制器的变化点位
func (w *worker) getAllChangedPoints() map[db.IDType]rt.Points {
	w.controllerChangedPointsMutex.RLock()
	defer w.controllerChangedPointsMutex.RUnlock()

	var (
		results      = make(map[db.IDType]rt.Points, len(w.controllerChangedPoints))
		resultsMutex sync.RWMutex
		wg           sync.WaitGroup
	)

	for controllerID, cache := range w.controllerChangedPoints {
		wg.Add(1)
		go func(id db.IDType, data *changedPointsType) {
			defer wg.Done()

			l := data.Len()
			if l == 0 {
				return
			}

			points := make(rt.Points, 0, l)
			cachePoints := data.MovePoints()
			for _, p := range cachePoints {
				points = append(points, p)
			}

			resultsMutex.Lock()
			results[id] = points
			resultsMutex.Unlock()
		}(controllerID, cache)
	}

	wg.Wait()

	return results
}

// reportChangedPoints 上报当前所有变化点位
func (w *worker) reportChangedPoints(ctx context.Context) {
	t := ttime.GetNowUTC().UnixMilli()
	points := w.getAllChangedPoints()
	w.reportAllControllerPoints(ctx, t, pb.DataKind_Change, points)
}

// isPointFirstAccess 检查点位是否为首次访问
func (w *worker) isPointFirstAccess(id string) bool {
	w.firstAccessMutex.RLock()
	defer w.firstAccessMutex.RUnlock()

	_, has := w.firstAccess[id]
	return has
}

// setPointHasAccessed 标记点位已被访问过
func (w *worker) setPointHasAccessed(id string) {
	w.firstAccessMutex.Lock()
	defer w.firstAccessMutex.Unlock()

	w.firstAccess[id] = struct{}{}
}
