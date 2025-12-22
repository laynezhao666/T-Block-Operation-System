// Package interval 周期数据处理
package interval

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"

	"github.com/robfig/cron/v3"

	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model/data"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"agent/logic/distribution/distributor"
	utils2 "agent/logic/distribution/distributor/utils"
	"agent/repo/monitor"
	"agent/utils"
)

// IntervalProcessorDevices 管理推送间隔下的设备
type IntervalProcessorDevices map[definition.DeviceGidType]*DeviceProcessor

// IntervalProcessor 管理推送间隔下的测点
type IntervalProcessor struct {
	c            *cron.Cron
	interval     int
	devices      IntervalProcessorDevices
	distributors distributor.Distributors
	mutex        sync.RWMutex
}

func newIntervalProcessor(interval int, distributors ...distributor.Distributor) (*IntervalProcessor, error) {
	if interval < 1 || 60%interval != 0 {
		return nil, fmt.Errorf("invalid interval: %v", interval)
	}

	p := &IntervalProcessor{
		c:            cron.New(cron.WithSeconds()),
		interval:     interval,
		distributors: distributors,
		devices:      make(IntervalProcessorDevices),
	}

	// 周期上报的基准时间点每个pod随机
	second := rand.Intn(60)
	//config.LoadIntOrDefault(config.GetRB().Distributor.Common.IntervalReportSecond, consts.DefaultNorthReportSecond)
	_, err := p.c.AddFunc(
		fmt.Sprintf("%v/%v * * * * *", second, interval), func() {
			p.Distribute()
		},
	)

	// 程序初始启动时，全量上报一次
	if time.Now().Second() < 50 {
		go func() {
			time.Sleep(10 * time.Second) // 等待采集一次完成。后续增加获得一个采集数据全部Read的状态，代替这种sleep的方式
			p.Distribute()
		}()
	}

	return p, err
}

// Start 启动
func (p *IntervalProcessor) Start() {
	if p == nil {
		return
	}
	log.Infof("interval[%v] processor start", p.interval)
	p.c.Start()
}

// Stop 停止
func (p *IntervalProcessor) Stop() {
	if p == nil {
		return
	}
	p.c.Stop()
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

	for _, d := range p.devices {
		go p.distribute(d.deviceGiD, d.GetPointsID())
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

func (p *IntervalProcessor) distribute(deviceGiD definition.DeviceGidType, points definition.DataPointIDsType) {
	if len(points) == 0 {
		return
	}

	dataPoints := rtdb.GetDataPointsByID(points)
	collected := verifyCollectedDataPoints(deviceGiD, dataPoints)
	currentDataUnit := &data.DataUnit{
		DeviceGid: deviceGiD,
		Points:    collected,
	}

	args := utils2.DistributorArgs{
		Time:     utils.GetNowUTCTime(),
		Interval: p.interval,
	}
	p.distributors.BatchDistribute(currentDataUnit, &args)

	p.report(deviceGiD, float64(len(collected)), float64(len(dataPoints)))
}

func (p *IntervalProcessor) report(deviceGiD definition.DeviceGidType, success float64, all float64) {
	attrsList := []*metrics.Dimension{
		{
			Name:  consts.AttrDeviceName,
			Value: string(deviceGiD),
		},
	}

	m := []*metrics.Metrics{
		metrics.NewMetrics("std_interval_success", success, metrics.PolicySUM),
		metrics.NewMetrics("std_interval_all", all, metrics.PolicySUM),
	}

	log.Debugf("%v interval: success=%v, all=%v", deviceGiD, success, all)
	monitor.ReportMultiMetricsWithDimensions(m, attrsList)
}
