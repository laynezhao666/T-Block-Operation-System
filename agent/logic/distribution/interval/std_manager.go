package interval

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/distribution/base"
	"agent/logic/distribution/distributor"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// StdIntervalManager 标准间隔管理器
type StdIntervalManager struct {
	*BaseIntervalManager
}

func newStdIntervalManager() *StdIntervalManager {
	m := &StdIntervalManager{
		BaseIntervalManager: &BaseIntervalManager{
			intervalProcessors: make(map[int]*IntervalProcessor),
			distributors:       base.GetDistributorList(consts.StdInterval),
		},
	}

	// 初始化默认处理器
	var err error
	m.defaultProcessor, err = newIntervalProcessor(definition.DefaultInterval, m.distributors...)
	if err != nil {
		log.Errorf("Failed to create std default interval processor: %v", err)
		return nil
	}
	return m
}

func (m *StdIntervalManager) refreshStdPoints() {
	if m == nil {
		return
	}

	for {
		if m.stopped {
			return
		}

		m.loadAllDevices()

		time.Sleep(definition.StdPointsRefreshTime)
	}
}

func (m *StdIntervalManager) loadAllDevices() {
	// 获取所有设备
	data := cm.Worker().GetStdData()
	copyData := data.Copy()
	if copyData == nil || len(copyData.StdPointsInfo) == 0 {
		return
	}

	pointsID := make(definition.DataPointIDsType, 0, len(copyData.StdPointsInfo))
	for _, v := range copyData.StdPointsInfo {
		point := definition.DataPointIDType(v.StdDevice + consts.DefaultIDSep + v.StdPoint)
		pointsID = append(pointsID, point)
	}
	m.defaultProcessor.SetPoints(definition.StdDevice, pointsID)
}

// ReloadAllDevice 重新加载全部设备
func (m *StdIntervalManager) ReloadAllDevice() {
	if m == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.defaultProcessor.ClearAllDevice()
	for _, p := range m.intervalProcessors {
		p.ClearAllDevice()
	}
	distributor.PointsDataManager().ClearAllDevice()

	m.loadAllDevices()
}

// Start 启动
func (m *StdIntervalManager) Start() {
	if m == nil {
		return
	}
	m.defaultProcessor.Start()
	for _, p := range m.intervalProcessors {
		p.Start()
	}
	m.stopped = false
	go m.refreshStdPoints()

	go m.refreshAllPoints()
	go m.recordPointsNumber()
	go m.recordPoints()
}
