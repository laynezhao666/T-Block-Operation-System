package interval

import (
	"agent/entity/config"
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

		m.loadAllDevices()

		time.Sleep(definition.DeviceRefreshTime)
	}
}

func (m *CollectIntervalManager) loadAllDevices() {
	// 获取所有设备
	tempDevices := cm.Worker().GetAllDevices()
	if len(tempDevices) == 0 {
		return
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
		if d.ID == consts.DeviceIDEDC {
			pointsID = append(pointsID, base.EdcMonitorPoints(string(d.Gid))...)
		}

		// gw模式，对于所有子设备确认是否有Comm，没有则添加到上报列表
		if config.GetRB().IsGatewayMode() {
			subCommToAdd := completeSubDeviceComm(pointsID)
			if len(subCommToAdd) > 0 {
				pointsID = append(pointsID, subCommToAdd...)
			}
		}

		m.defaultProcessor.SetPoints(d.Gid, pointsID)
	}
}

func completeSubDeviceComm(points definition.DataPointIDsType) definition.DataPointIDsType {
	// 按gid分组，并记录是否有Comm
	sub2Comm := map[definition.DeviceGidType]bool{}
	for _, point := range points {
		deviceId, pointId, err := definition.SplitDataPointID(point)
		if err != nil {
			log.Infof("invalid point id [%v], split failed", point)
			continue
		}
		hasComm := pointId == definition.CommID
		if _, ok := sub2Comm[deviceId]; !ok {
			// 创建时赋值
			sub2Comm[deviceId] = hasComm
		} else {
			if hasComm {
				// 有Comm点时赋值
				sub2Comm[deviceId] = hasComm

				continue
			}
		}
	}
	commToAdd := make(definition.DataPointIDsType, 0, len(sub2Comm))
	for subDeviceGid, has := range sub2Comm {
		if !has {
			commToAdd = append(commToAdd,
				definition.DataPointIDType(subDeviceGid+consts.DefaultIDSep+definition.CommID))
		}
	}
	return commToAdd
}

// ReloadAllDevice 重新加载全部设备
func (m *CollectIntervalManager) ReloadAllDevice() {
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
