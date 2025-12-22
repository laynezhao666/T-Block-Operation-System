package southdevice

import (
	"math"
	"strings"
	"sync"
	"time"

	"agent/entity/definition"
	"agent/entity/model"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	model2 "agent/logic/collector/rtdb/model"
	"agent/logic/plugin"
	consts "agent/logic/plugin/var"
	"agent/utils"
)

type southPlugin struct {
}

// Notify 通知
func (p *southPlugin) Notify(event plugin.EventType) {}

// Do 执行
func (p *southPlugin) Do(_ interface{}) {
	devices := cm.Worker().GetAllDevices()
	if len(devices) == 0 {
		return
	}

	currentTime := utils.GetNowUTCTime()
	pushPoints := make(model2.DataPoints, 0, len(devices))

	var mutex sync.Mutex
	var wg sync.WaitGroup
	for i := range devices {
		wg.Add(1)
		go func(deviceGid definition.DeviceGidType) {
			defer wg.Done()

			points := p.do(currentTime, deviceGid)
			for i := range points {
				points[i].DeviceGiD = deviceGid
			}

			mutex.Lock()
			defer mutex.Unlock()

			pushPoints = append(pushPoints, points...)
		}(devices[i].Gid)
	}
	wg.Wait()

	if len(pushPoints) > 0 {
		go func(p model2.DataPoints) {
			rtdb.SetDataPoints(p, true)
		}(pushPoints)
	}

	// 此为周期上报，可暂时屏蔽。 因为已经触发变化上报
	//u := &data.DataUnit{
	//	DeviceID: mconsts.DeviceIDVirtual,
	//	Points:   pushPoints,
	//}
	//distArg := utils2.DistributorArgs{
	//	Time:     currentTime,
	//	Interval: 1,
	//}
	//kafka.KafkaDistributor().Distribute(u, &distArg)
}

// ProcessRtd 处理测点数据
func (p *southPlugin) ProcessRtd(deviceID definition.DeviceGidType, points model2.DataPoints, ignore bool) {
	//if ignore {
	//	logOnce.Do(func() {
	//		filterLog = flog.NewFilterLogger(time.Hour, config.Log)
	//	})
	//	filterLog.Infof(deviceID, "southPlugin ignore device %v", deviceID)
	//	return
	//}
	setPointQuaErrByCommPoint(deviceID, points)
}

func calcSubDeviceInterruption(pushPoints model2.DataPoints, currentTime time.Time,
	subDevice *model.SubDeviceData) model2.DataPoints {
	subDeviceGiD := subDevice.InstanceDeviceGid
	if len(subDevice.PointsInfo) == 0 {
		return pushPoints
	}

	points := getCachedPoints(subDeviceGiD, subDevice.PointsInfo)
	commID := definition.GenerateCommID(subDeviceGiD)

	has, ok := hasCommID(subDeviceGiD)
	if !ok {
		// 如果缓存中没有记录是否存在 CommID，则遍历测点判断一次
		has = false
		for j := range points {
			if points[j].ID == commID {
				has = true
				break
			}
		}
		setHasCommID(subDeviceGiD, has)
	}
	// 已存在 CommID，跳过后续计算
	if has {
		return pushPoints
	}

	// 正常测点及中断测点总数，排除其他异常值
	total := 0
	// 中断测点数
	interrupted := 0

	rtdb.GetDataPoints(points)
	// 遍历子设备下所有测点
	// 判断测点质量是否 ok(排除snmp下层错误)，测点值是否均为中断
	// todo: 提前结束循环
	for j := range points {
		val := &points[j].Rtd.Val
		if !val.IsSubDeviceErr() && !val.IsOK() {
			continue
		}
		v := val.Pv.String()
		if v == definition.OfflineValue {
			interrupted++
		} else if strings.HasPrefix(v, "-9999") {
			continue
		}

		total++
	}

	minInterruptionNum := int(math.Ceil(float64(total*consts.InterruptionJudgeThreshold) / 100.0))
	isInterrupted := false
	// 最小中断测点数为 0 时，不能判定中断
	if interrupted >= minInterruptionNum && minInterruptionNum > 0 {
		isInterrupted = true
	}

	// 当所有测点质量 ok 且中断测点数超过阈值时，该子设备判定为中断
	interruptionPoint := &model2.DataPoint{
		ID:  definition.GenerateInterruptionID(subDeviceGiD),
		Rtd: model2.NewVirtualRtDataWithValueTime(isInterrupted, currentTime.Unix()),
	}
	// 复制测点，CommID 由业务使用
	commPoint := &model2.DataPoint{
		ID: commID,
		// 上层使用 0 或 1 而不是布尔值
		Rtd: model2.NewVirtualRtDataWithValueTime(utils.Bool2Int(isInterrupted), currentTime.Unix()),
	}

	pushPoints = append(pushPoints, *interruptionPoint, *commPoint)
	return pushPoints
}

func (p *southPlugin) do(currentTime time.Time, deviceGid definition.DeviceGidType) model2.DataPoints {
	templateData := cm.GetCachedTemplateData(deviceGid)
	subDevices := templateData.SubDevices
	var pushPoints = make(model2.DataPoints, 0, len(subDevices)*2)
	for i := range subDevices {
		pushPoints = calcSubDeviceInterruption(pushPoints, currentTime, &subDevices[i])
	}
	return pushPoints
}

// GetInterval 获取轮训间隔
func (p *southPlugin) GetInterval() int {
	return 30
}
