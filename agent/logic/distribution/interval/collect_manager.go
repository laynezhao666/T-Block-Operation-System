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

// CollectIntervalManager 采集间隔管理器
type CollectIntervalManager struct {
	*BaseIntervalManager
}

func newCollectIntervalManager() *CollectIntervalManager {
	m := &CollectIntervalManager{
		BaseIntervalManager: &BaseIntervalManager{
			intervalProcessors: make(map[int]*IntervalProcessor),
			distributors:       base.GetDistributorList(consts.CollectInterval),
		},
	}

	// 初始化默认处理器
	var err error
	m.defaultProcessor, err = newIntervalProcessor(definition.DefaultInterval, m.distributors...)
	if err != nil {
		log.Errorf("Failed to create collect default interval processor: %v", err)
		return nil
	}
	return m
}

func (m *CollectIntervalManager) refreshCollectDevices() {
	if m == nil {
		return
	}

	for {
		if m.stopped {
			return
		}

		// 获取所有设备
		tempDevices := cm.Worker().GetAllDevices()
		if len(tempDevices) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for i := range tempDevices {
			d := &tempDevices[i]
			if m.stopped {
				return
			}
			// 添加驱动模版内的点位
			dataPoints := distributor.PointsDataManager().GetDataPoints(d.Gid)
			pointsID := make(definition.DataPointIDsType, len(dataPoints))
			for i, point := range dataPoints {
				pointsID[i] = point.ID
			}
			// 添加通信状态虚拟点（厂商侧使用）
			commstePoint := definition.DataPointIDType(d.Gid + consts.DefaultIDSep + definition.CommunicationStatusID)
			pointsID = append(pointsID, commstePoint)

			// 通信状态（原版腾讯动环使用）
			commPoint := definition.DataPointIDType(d.Gid + consts.DefaultIDSep + definition.CommID)
			pointsID = append(pointsID, commPoint)

			// 采集器自身监控点（cpu、内存等）
			if d.ID == "EDC_1" {
				pointsID = append(pointsID, base.EdcMonitorPoints(string(d.Gid))...)
			}

			m.defaultProcessor.SetPoints(d.Gid, pointsID)
		}

		time.Sleep(definition.DeviceRefreshTime)
	}
}

// Start 启动采集间隔管理器
func (m *CollectIntervalManager) Start() {
	if m == nil {
		return
	}
	m.defaultProcessor.Start()
	for _, p := range m.intervalProcessors {
		p.Start()
	}
	m.stopped = false
	go m.refreshCollectDevices()

	go m.refreshAllPoints()
	go m.recordPointsNumber()
	go m.recordPoints()
}
