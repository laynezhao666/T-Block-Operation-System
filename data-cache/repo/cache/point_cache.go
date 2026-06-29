// Package cache 循环窗口缓存，用于存储单个测点近一段时间的数据
package cache

import (
	"data-cache/entity/model"
	"maps"
	"sync"
	"sync/atomic"
	"time"
)

const (
	extraWindowCnt    uint32 = 2  // 额外窗口数量
	defaultWindowSize uint32 = 60 // 默认窗口大小,最大不超过256
	defaultWindowCnt  uint32 = 11 // 默认窗口数量
)

var (
	windowSize = defaultWindowSize                    // 窗口大小,一个窗口60s
	windowCnt  = defaultWindowCnt + extraWindowCnt    // 默认存储11分钟的数据+2个额外窗口
	expire     = defaultWindowSize * defaultWindowCnt // 数据过期时间
)

// SetWindowCfg 修改窗口配置信息
func SetWindowCfg(wSize, wCnt uint32) {
	windowSize = wSize
	windowCnt = wCnt + extraWindowCnt
	expire = wSize * wCnt
}

// PointCache 单个测点数据缓存
type PointCache struct {
	lastChangeTs uint32      // 上次变化时间
	windowData   []PointList // 所有窗口数据,窗口循环使用,一个窗口存储1分钟的数据
	lock         sync.Mutex  // 写锁
}

// NewPointCache 创建一个循环窗口列表
func NewPointCache() *PointCache {
	obj := &PointCache{
		windowData: make([]PointList, windowCnt),
	}
	for i := range windowCnt {
		obj.windowData[i] = PointList{
			points: make([]CompressPoint, 0),
		}
	}
	return obj
}

func (s *PointCache) getWindowIdx(ts uint32) uint32 {
	return (ts / windowSize) % windowCnt
}

func (s *PointCache) nextWindowIdx(idx uint32) uint32 {
	return (idx + 1) % windowCnt
}

func (s *PointCache) preWindowIdx(idx uint32) uint32 {
	return (idx - 1 + windowCnt) % windowCnt
}

// Add 新增一个元素
func (s *PointCache) Add(ts uint32, val *model.CachePoint) {
	expireTs := uint32(time.Now().Unix()) - expire
	// 过期的数据
	if ts < expireTs {
		return
	}
	idx := s.getWindowIdx(ts)
	// 属于哪个窗口
	windowOffset := ts % windowSize
	// 窗口的起始时间
	windowTs := ts - windowOffset
	wData := &s.windowData[idx]
	s.lock.Lock()
	defer s.lock.Unlock()
	if wData.ts < windowTs {
		wData.Clear()
		wData.ts = windowTs
	}
	wData.Add(uint8(windowOffset), val)
}

// SetLastChangeTs 设置最新变更时间
func (s *PointCache) SetLastChangeTs(ts uint32) {
	s.lastChangeTs = ts
}

// GetLastChangeTs 获取上次变更时间
func (s *PointCache) GetLastChangeTs() uint32 {
	return s.lastChangeTs
}

// Range 根据时间范围查询
func (s *PointCache) Range(begin, end uint32) map[uint32]*model.CachePoint {
	bIdx := s.getWindowIdx(begin)
	eIdx := s.getWindowIdx(end)
	res := make(map[uint32]*model.CachePoint)
	for {
		maps.Copy(res, (&s.windowData[bIdx]).Filter(begin, end))
		if bIdx == eIdx {
			break
		}
		bIdx = s.nextWindowIdx(bIdx)
	}
	// begin所在的数据不存在,则往前找一个
	if _, ok := res[begin]; !ok {
		lTs, lVal := s.Latest(begin)
		if lTs != 0 {
			res[lTs] = lVal
		}
	}
	return res
}

// Latest 根据时间查询最新数据
func (s *PointCache) Latest(end uint32) (uint32, *model.CachePoint) {
	eIdx := s.getWindowIdx(end)
	bIdx := s.nextWindowIdx(eIdx) // 从当前窗口往前找, 因为是个环，最大找到下一个窗口
	begin := end - end%windowSize

	for bIdx != eIdx {
		vals := (&s.windowData[eIdx]).Filter(begin, end)
		if len(vals) > 0 {
			var maxTs uint32
			for ts := range vals {
				if ts > maxTs {
					maxTs = ts
				}
			}
			return maxTs, vals[maxTs]
		}
		begin -= windowSize
		eIdx = s.preWindowIdx(eIdx)
	}
	return 0, nil
}

// CompressPoint 压缩测点
type CompressPoint struct {
	val      float64 // 测点值
	quality  int16   // 数据质量
	tsOffset uint8   // 时间偏移量
}

// PointList 测点列表
type PointList struct {
	ts     uint32
	len    int32 // 已写入的元素数量，原子操作，保证读取一致性
	points []CompressPoint
}

func (obj *PointList) Add(tsOffset uint8, addPoint *model.CachePoint) {
	// 判断测点是否已经存在
	// 似乎没必要，多存储几个点不影响，读的时候后到的点会覆盖前面的点的值
	//for _, point := range obj.points {
	//	if point.tsOffset == tsOffset {
	//		point.val = addPoint.Value
	//		point.quality = addPoint.Quality
	//		return
	//	}
	//}
	obj.points = append(obj.points, CompressPoint{
		tsOffset: tsOffset,
		quality:  addPoint.Quality,
		val:      addPoint.Value,
	})
	// len++ 置于完成之后，提供release语义
	atomic.StoreInt32(&obj.len, int32(len(obj.points)))
}

// Filter 根据时间区间过滤测点
func (obj *PointList) Filter(begin, end uint32) map[uint32]*model.CachePoint {
	res := make(map[uint32]*model.CachePoint)
	var points = obj.points         // 获取切片快照
	n := atomic.LoadInt32(&obj.len) // 读取已发布长度(即acquire语义)
	n = min(n, int32(len(points)))
	ts := obj.ts
	for i := int32(0); i < n; i++ {
		point := points[i]
		pointTs := ts + uint32(point.tsOffset)
		if pointTs >= begin && pointTs <= end {
			res[pointTs] = &model.CachePoint{
				Value:   point.val,
				Quality: point.quality,
			}
		}
	}
	return res
}

func (obj *PointList) Clear() {
	atomic.StoreInt32(&obj.len, 0)        // 归零长度，读取方读取到0时，不会再对切片进行遍历
	obj.points = make([]CompressPoint, 0) // 然后，替换切片
}
