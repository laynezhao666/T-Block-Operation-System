package distributor

import (
	"agent/utils/osal"
	"errors"
	"fmt"
	"sync"

	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	model2 "agent/logic/collector/rtdb/model"
)

var (
	manager *PointsDataMap
)

func init() {
	manager = &PointsDataMap{
		m: make(PointsDataMapType),
	}
}

// PointsData 定时刷新的PointsData 测点数据
type PointsData struct {
	timer  *osal.ExpireTimer
	points model2.DataPoints
}

// PointsDataMapType 测点数据map
type PointsDataMapType map[definition.DeviceGidType]*PointsData

// PointsDataMap 管理所有的测点数据
type PointsDataMap struct {
	m PointsDataMapType
	sync.RWMutex
}

// PointsDataManager 获取PointsDataMap
func PointsDataManager() *PointsDataMap {
	return manager
}

// DeleteDevice 删除设备
func (p *PointsDataMap) DeleteDevice(deviceGiD definition.DeviceGidType) {
	if p == nil {
		return
	}
	p.Lock()
	defer p.Unlock()
	delete(p.m, deviceGiD)
}

// ClearAllDevice 清空全部设备缓存
func (p *PointsDataMap) ClearAllDevice() {
	if p == nil {
		return
	}
	p.Lock()
	defer p.Unlock()
	p.m = make(PointsDataMapType)
}

// HasPoint 判断是否有测点
func (p *PointsDataMap) HasPoint(pointIDPair *definition.IDPair) bool {
	_, has := rtdb.GetPv(pointIDPair.PointInstanceID)
	return has
}

func (p *PointsDataMap) fetchPoints(deviceGiD definition.DeviceGidType) (model2.DataPoints, error) {
	var points model2.DataPoints
	if p == nil {
		return points, errors.New("PointsDataMap is nil")
	}

	templateData, ok := cm.Worker().GetDeviceTemplateByGid(deviceGiD)
	if !ok {
		return points, fmt.Errorf("not find device %v templatte", deviceGiD)
	}

	pointsInfo := templateData.GetPoints()
	points = make(model2.DataPoints, 0, len(pointsInfo))
	for i := range pointsInfo {
		point := &pointsInfo[i]
		points = append(points, model2.DataPoint{
			ID:             point.ID,
			Rtd:            model2.NewRTData(),
			IsValueChanged: false,
		})
	}

	data := &PointsData{
		timer:  osal.NewExpireTimer(definition.TemplateCacheDataDuration),
		points: points,
	}
	data.timer.SetAccess()

	p.Lock()
	defer p.Unlock()

	p.m[deviceGiD] = data
	return points, nil
}

// GetDataPoints 获取测点数据
func (p *PointsDataMap) GetDataPoints(deviceGiD definition.DeviceGidType) model2.DataPoints {
	var points model2.DataPoints
	if p == nil {
		return points
	}

	p.RLock()
	d, ok := p.m[deviceGiD]
	p.RUnlock()

	if ok {
		if !d.timer.IsExpired() {
			return d.points
		}

		oldPoints := d.points
		points, err := p.fetchPoints(deviceGiD)
		if err != nil {
			// 如果获取测点失败，使用已存在的数据
			return oldPoints
		}
		return points
	}

	points, _ = p.fetchPoints(deviceGiD)
	return points
}
