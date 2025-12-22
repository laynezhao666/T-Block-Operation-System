package interval

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb"

	"trpc.group/trpc-go/trpc-go/log"
)

// DeviceProcessorPoints 设备测点集合
type DeviceProcessorPoints map[definition.DataPointIDType]bool

// DeviceProcessor 管理设备下的测点
type DeviceProcessor struct {
	deviceGiD definition.DeviceGidType
	points    DeviceProcessorPoints
}

// NewDeviceProcessor 创建设备处理器
func NewDeviceProcessor(deviceGiD definition.DeviceGidType) *DeviceProcessor {
	return &DeviceProcessor{
		deviceGiD: deviceGiD,
		points:    make(DeviceProcessorPoints),
	}
}

// GetPointsID 获取测点ID列表
func (d *DeviceProcessor) GetPointsID() definition.DataPointIDsType {
	if d == nil {
		return nil
	}
	ids := make(definition.DataPointIDsType, 0, len(d.points))
	for point := range d.points {
		ids = append(ids, point)
	}
	return ids
}

// GetPointsNumber 获取测点数量
func (d *DeviceProcessor) GetPointsNumber() int {
	if d == nil {
		return 0
	}
	return len(d.points)
}

// IsEmpty 判断是否为空
func (d *DeviceProcessor) IsEmpty() bool {
	if d == nil {
		return true
	}
	return len(d.points) == 0
}

// AddPointID 添加测点ID
func (d *DeviceProcessor) AddPointID(point definition.DataPointIDType) {
	if d == nil {
		return
	}
	d.points[point] = true
}

// SetPointsID 设置测点ID
func (d *DeviceProcessor) SetPointsID(points definition.DataPointIDsType) {
	if d == nil {
		return
	}
	d.points = make(DeviceProcessorPoints)
	for _, point := range points {
		d.points[point] = true
	}
}

// DeletePointID 删除测点ID
func (d *DeviceProcessor) DeletePointID(point definition.DataPointIDType) {
	if d == nil {
		return
	}
	d.deletePointID(point)
}

func (d *DeviceProcessor) deletePointID(point definition.DataPointIDType) {
	delete(d.points, point)
}

// PrunePoints 删除不存在的测点
func (d *DeviceProcessor) PrunePoints() {
	if d == nil {
		return
	}

	notExistPoints := make([]definition.DataPointIDType, 0)
	for point := range d.points {
		if _, has := rtdb.GetPv(point); !has {
			notExistPoints = append(notExistPoints, point)
		}
	}
	if len(notExistPoints) == 0 {
		return
	}

	for _, point := range notExistPoints {
		d.deletePointID(point)
	}
	log.Infof("remove uncollected points: %+v", notExistPoints)
}
