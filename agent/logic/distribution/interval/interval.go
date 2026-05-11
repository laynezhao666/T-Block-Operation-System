// Package interval 周期数据处理
package interval

import (
	"agent/logic/cm"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model/data"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"agent/logic/distribution/distributor"
	utils2 "agent/logic/distribution/distributor/utils"
	"agent/utils"
)

// IntervalProcessorDevices 管理推送间隔下的设备
type IntervalProcessorDevices map[definition.DeviceGidType]*DeviceProcessor

// IntervalProcessor 管理推送间隔下的测点
type IntervalProcessor struct {
	interval     int
	devices      IntervalProcessorDevices
	distributors distributor.Distributors
	mutex        sync.RWMutex
	stopChan     chan struct{}
}

func newIntervalProcessor(interval int, distributors ...distributor.Distributor) (*IntervalProcessor, error) {
	if interval < 1 || 60%interval != 0 {
		return nil, fmt.Errorf("invalid interval: %v", interval)
	}

	p := &IntervalProcessor{
		interval:     interval,
		distributors: distributors,
		devices:      make(IntervalProcessorDevices),
		stopChan:     make(chan struct{}),
	}

	return p, nil
}

// Start 启动
func (p *IntervalProcessor) Start() {
	if p == nil {
		return
	}
	log.Infof("interval[%v] processor start", p.interval)

	go func() {
		// 让每个tbox的上报时间随机打散
		initialDelay := time.Duration(rand.Intn(p.interval)) * time.Second
		log.Warnf("initial delay of %v for interval %v", initialDelay, p.interval)
		time.Sleep(initialDelay)

		// 启动定时器，使用单调时钟，极端情况下不受系统时间突变影响
		ticker := time.NewTicker(time.Duration(p.interval) * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Debugf("interval processor detected %d seconds", initialDelay)
				p.Distribute()
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	// 程序初始启动时，全量上报一次
	if time.Now().Second() < 50 {
		go func() {
			time.Sleep(15 * time.Second) // 等待采集一次完成。后续增加获得一个采集数据全部Read的状态，代替这种sleep的方式
			p.Distribute()
		}()
	}

}

// Stop 停止
func (p *IntervalProcessor) Stop() {
	if p == nil {
		return
	}
	close(p.stopChan)
}

// SetPoints 设置测点
func (p *IntervalProcessor) SetPoints(deviceGiD definition.DeviceGidType, pointsID definition.DataPointIDsType) {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	d := p.getDeviceProcessor(deviceGiD)
	d.SetPointsID(pointsID)
}

func (p *IntervalProcessor) getPointsNumber() int {
	n := 0
	for _, d := range p.devices {
		n += d.GetPointsNumber()
	}
	return n
}

// GetPointsNumber 获取测点数量
func (p *IntervalProcessor) GetPointsNumber() int {
	if p == nil {
		return 0
	}
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.getPointsNumber()
}

// GetPoints 获取测点
func (p *IntervalProcessor) GetPoints() definition.DataPointIDsType {
	if p == nil {
		return nil
	}
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	ids := make(definition.DataPointIDsType, 0, p.getPointsNumber())
	for _, d := range p.devices {
		points := d.GetPointsID()
		ids = append(ids, points...)
	}
	return ids
}

// AddPoint 添加测点
func (p *IntervalProcessor) AddPoint(point definition.IDPair) {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 推送间隔低于 60s 的测点都放在一条消息内，不按照设备划分为多条消息
	deviceGid := getDeviceGid(p.interval, point)
	d := p.getDeviceProcessor(deviceGid)
	d.AddPointID(point.PointInstanceID)
}

// DeletePoint 删除测点
func (p *IntervalProcessor) DeletePoint(point definition.IDPair) {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	deviceGid := getDeviceGid(p.interval, point)
	d := p.getDeviceProcessorOrNil(deviceGid)
	d.DeletePointID(point.PointInstanceID)
	if d.IsEmpty() {
		p.deleteDevice(deviceGid)
	}
}

// PrunePoints 清理空测点
func (p *IntervalProcessor) PrunePoints() {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	emptyDevices := make([]definition.DeviceGidType, 0)
	for _, d := range p.devices {
		d.PrunePoints()
		if d.IsEmpty() {
			emptyDevices = append(emptyDevices, d.deviceGiD)
		}
	}
	for _, deviceGid := range emptyDevices {
		p.deleteDevice(deviceGid)
	}
}

// DeleteDevice 删除设备
func (p *IntervalProcessor) DeleteDevice(deviceGiD definition.DeviceGidType) {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.deleteDevice(deviceGiD)
}

// ClearAllDevice 清空全部设备
func (p *IntervalProcessor) ClearAllDevice() {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.devices = make(IntervalProcessorDevices)
}

func (p *IntervalProcessor) deleteDevice(deviceGiD definition.DeviceGidType) {
	delete(p.devices, deviceGiD)
}

func (p *IntervalProcessor) getDeviceProcessor(deviceGiD definition.DeviceGidType) *DeviceProcessor {
	if p == nil {
		return nil
	}
	d, ok := p.devices[deviceGiD]
	if !ok {
		d = NewDeviceProcessor(deviceGiD)
		p.devices[deviceGiD] = d
	}
	return d
}

func (p *IntervalProcessor) getDeviceProcessorOrNil(deviceGiD definition.DeviceGidType) *DeviceProcessor {
	if p == nil {
		return nil
	}
	d, ok := p.devices[deviceGiD]
	if !ok {
		return nil
	}
	return d
}

// Distribute 分发采集数据
func (p *IntervalProcessor) Distribute() {
	if p == nil {
		return
	}
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 确定 dataType（同一次 Distribute 请求中都一致）
	var dataType int
	for _, d := range p.devices {
		switch d.deviceGiD {
		case definition.StdDevice:
			dataType = definition.KafkaDataTypeStd
		default:
			dataType = definition.KafkaDataTypeCollector
		}
		break // 取第一个即可，因为同一次请求中 dataType 一致
	}

	// 全局按 mozu 聚合，不再按 device 拆分
	mozuMap := make(map[string][]definition.DataPointIDType)
	for _, d := range p.devices {
		pointsId := d.GetPointsID()
		for _, pointId := range pointsId {
			mozuId := cm.Worker().GetDeviceMozuID(definition.DeviceGidType(pointId.GetPointGid()))
			mozuMap[mozuId] = append(mozuMap[mozuId], pointId)
		}
	}

	// 按 mozu 分发
	for mozuId, ps := range mozuMap {
		if len(ps) > 0 {
			// 收集该 mozu 下所有设备 gid（去重）
			gidSet := make(map[definition.DeviceGidType]struct{})
			for _, pid := range ps {
				g := definition.DeviceGidType(pid.GetPointGid())
				gidSet[g] = struct{}{}
			}
			deviceGids := make([]definition.DeviceGidType, 0, len(gidSet))
			for g := range gidSet {
				deviceGids = append(deviceGids, g)
			}
			// 取第一个 gid 作为主标识（兼容下游），同时传递完整列表
			gid := deviceGids[0]
			go p.distribute(gid, deviceGids, ps, dataType, mozuId)
		}
	}
}

// 分割已采集数据和未采集数据, 以及时间戳延时验证；添加统计类虚拟点
func verifyCollectedDataPoints(deviceGiD definition.DeviceGidType, points model.DataPoints) model.DataPoints {
	stat := NewQualityStatistics(deviceGiD)
	collected := make(model.DataPoints, 0, len(points))
	for i := range points {
		// 时间戳延时判断和统计, 当测点时间戳晚于当前10-60秒之间，调整测点时间戳（缓冲采集延时异常，防止上层无法获得该分钟级数据）。
		diffSecond := points[i].Rtd.Val.TmsDiffByNow()
		if diffSecond >= definition.TenSecondTmsDelay && diffSecond <= definition.OneMinutesTmsDelay {
			points[i].Rtd.Val.Tms = utils.GetNowUTCTimeStamp() - definition.TenSecondTmsDelay
		}
		stat.CountTmsDelay(points[i].ID, diffSecond)

		// 质量标签降噪和统计
		quaOriginal := points[i].Rtd.Val.Qua
		points[i].Rtd.Val.Qua = deNoiseQualityErr(points[i].Rtd.Val)
		stat.CountQuaError(points[i].ID, quaOriginal, points[i].Rtd.Val.Qua)

		if points[i].Rtd.Val.NotCollected() {
			continue
		}
		collected = append(collected, points[i])
	}
	stat.Report()
	return collected
}

// verifyCollectedDataPointsByDevice 按设备拆分点位进行质量统计，
// 每台设备的统计类虚拟点写入各自的 rtdb key 中
func verifyCollectedDataPointsByDevice(
	deviceGids []definition.DeviceGidType,
	points model.DataPoints,
) model.DataPoints {
	// 按设备 gid 分组点位索引
	devicePointIdx := make(map[definition.DeviceGidType][]int)
	for i := range points {
		gid := definition.DeviceGidType(points[i].ID.GetPointGid())
		devicePointIdx[gid] = append(devicePointIdx[gid], i)
	}

	// 每台设备分别统计
	for _, gid := range deviceGids {
		idxList, ok := devicePointIdx[gid]
		if !ok {
			continue
		}
		stat := NewQualityStatistics(gid)
		for _, idx := range idxList {
			diffSecond := points[idx].Rtd.Val.TmsDiffByNow()
			if diffSecond >= definition.TenSecondTmsDelay &&
				diffSecond <= definition.OneMinutesTmsDelay {
				points[idx].Rtd.Val.Tms = utils.GetNowUTCTimeStamp() - definition.TenSecondTmsDelay
			}
			stat.CountTmsDelay(points[idx].ID, diffSecond)

			quaOriginal := points[idx].Rtd.Val.Qua
			points[idx].Rtd.Val.Qua = deNoiseQualityErr(points[idx].Rtd.Val)
			stat.CountQuaError(points[idx].ID, quaOriginal, points[idx].Rtd.Val.Qua)
		}
		stat.Report()
	}

	// 过滤出已采集的点位
	collected := make(model.DataPoints, 0, len(points))
	for i := range points {
		if !points[i].Rtd.Val.NotCollected() {
			collected = append(collected, points[i])
		}
	}
	return collected
}

// deNoiseQualityErr 过滤短暂的质量异常
func deNoiseQualityErr(val model.RTValue) consts.Quality {
	if val.Qua == consts.QualityOk {
		return consts.QualityOk
	}

	if val.Qua != consts.QualityCommDisconnected { // 通讯中断类的质量异常不过滤
		if val.LastQuaOkTmsDiffByNow() <= definition.OneMinutesTmsDelay { // 隔最近质量正常时间不超过1分钟,返回质量正常
			return consts.QualityOk
		}
	}
	return val.Qua
}

func (p *IntervalProcessor) distribute(deviceGiD definition.DeviceGidType,
	deviceGids []definition.DeviceGidType, points definition.DataPointIDsType,
	dataType int, mozuId string) {
	if len(points) == 0 {
		return
	}

	dataPoints := rtdb.GetDataPointsByID(points)
	// 按设备拆分点位，分别做质量统计
	collected := verifyCollectedDataPointsByDevice(deviceGids, dataPoints)
	currentDataUnit := &data.DataUnit{
		DeviceGid:  deviceGiD,
		DeviceGids: deviceGids,
		Points:     collected,
	}

	args := utils2.DistributorArgs{
		Time:     utils.GetNowUTCTime(),
		Interval: p.interval,
		DataType: dataType,
		MozuID:   mozuId,
	}
	p.distributors.BatchDistribute(currentDataUnit, &args)
}
