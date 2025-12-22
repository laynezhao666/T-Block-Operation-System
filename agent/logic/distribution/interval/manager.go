package interval

import (
	"agent/entity/config"
	"agent/logic/distribution/distributor"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
)

var (
	stdManager     *StdIntervalManager
	collectManager *CollectIntervalManager
)

// Init 初始化
func Init() {
	if config.GetRB().IsCollectReportEnable() {
		collectManager = newCollectIntervalManager()
		collectManager.Start()
	}

	if config.GetRB().IsStdCalEnable() {
		stdManager = newStdIntervalManager()
		stdManager.Start()
	}
}

// UnInit 释放资源
func UnInit() {
	if config.GetRB().IsCollectReportEnable() {
		collectManager.Stop()
	}
	if config.GetRB().IsStdCalEnable() {
		stdManager.Stop()
	}
}

// CollectProcessorManager 获取采集间隔处理器
func CollectProcessorManager() *CollectIntervalManager {
	return collectManager
}

// StdProcessorManager 获取标准间隔处理器
func StdProcessorManager() *StdIntervalManager {
	return stdManager
}

// BaseIntervalManager 基础间隔处理器
type BaseIntervalManager struct {
	intervalProcessors map[int]*IntervalProcessor // 可自定义其他间隔时间
	defaultProcessor   *IntervalProcessor         // 默认interval是60
	distributors       distributor.Distributors
	mutex              sync.RWMutex
	stopped            bool
}

// DeleteDevice 删除设备
func (m *BaseIntervalManager) DeleteDevice(deviceGiD definition.DeviceGidType) {
	if m == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.defaultProcessor.DeleteDevice(deviceGiD)
	for _, p := range m.intervalProcessors {
		p.DeleteDevice(deviceGiD)
	}

	distributor.PointsDataManager().DeleteDevice(deviceGiD)
}

// AddPoint 添加测点
func (m *BaseIntervalManager) AddPoint(interval int, point definition.IDPair) {
	if m == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果间隔为默认间隔，表示不需要以更低间隔推送数据
	// 将其删除
	if interval == definition.DefaultInterval {
		for _, p := range m.intervalProcessors {
			p.DeletePoint(point)
		}
		return
	}

	processor := m.getIntervalProcessor(interval)
	processor.AddPoint(point)
	for i, p := range m.intervalProcessors {
		if i == interval {
			continue
		}
		p.DeletePoint(point)
	}
}

// Stop 停止
func (m *BaseIntervalManager) Stop() {
	if m == nil {
		return
	}
	m.defaultProcessor.Stop()
	for _, p := range m.intervalProcessors {
		p.Stop()
	}
	m.stopped = true
}

func (m *BaseIntervalManager) getIntervalProcessor(interval int) *IntervalProcessor {
	var err error
	processor, ok := m.intervalProcessors[interval]
	if !ok {
		processor, err = newIntervalProcessor(interval, m.distributors...)
		if err != nil {
			log.Errorf("newIntervalProcessor error: %v", err)
			return nil
		}
		processor.Start()
		m.intervalProcessors[interval] = processor
	}
	return processor
}

// GetPoints 获取测点
func (m *BaseIntervalManager) GetPoints() map[int]definition.DataPointIDsType {
	if m == nil {
		return nil
	}

	data := make(map[int]definition.DataPointIDsType)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for i, p := range m.intervalProcessors {
		data[i] = p.GetPoints()
	}
	return data
}

func (m *BaseIntervalManager) logPoints() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for i, p := range m.intervalProcessors {
		ids := p.GetPoints()
		log.Infof("interval: %v, points id: %v", i, ids)
	}
}

func (m *BaseIntervalManager) recordPoints() {
	if m == nil {
		return
	}
	for {
		if m.stopped {
			return
		}
		m.logPoints()
		time.Sleep(time.Hour)
	}
}

func (m *BaseIntervalManager) logPointsNumber() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	n := 0
	x := 0
	for i, p := range m.intervalProcessors {
		x = p.GetPointsNumber()
		n += x
		log.Infof("interval: %v, points number: %v", i, x)
	}
	log.Infof("all points(push interval < default) number: %v", n)
}

func (m *BaseIntervalManager) recordPointsNumber() {
	if m == nil {
		return
	}
	for {
		if m.stopped {
			return
		}
		m.logPointsNumber()
		time.Sleep(10 * time.Minute)
	}
}

func (m *BaseIntervalManager) refreshAllPointsImpl() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, p := range m.intervalProcessors {
		p.PrunePoints()
	}
}

func (m *BaseIntervalManager) refreshAllPoints() {
	if m == nil {
		return
	}

	for {
		if m.stopped {
			return
		}
		m.refreshAllPointsImpl()
		time.Sleep(definition.DeviceRefreshTime)
	}
}
