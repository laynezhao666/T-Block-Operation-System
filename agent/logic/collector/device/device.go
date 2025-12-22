// Package device 采集设备
package device

import (
	"context"
	"errors"
	"fmt"
	"agent/entity/config"
	"agent/logic/cm"
	"agent/logic/collector/device/driver"
	"agent/logic/logfile"
	"agent/logic/plugin"
	utils2 "agent/utils"
	osal2 "agent/utils/osal"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/logic/collector/device/virtualpoints"
	model4 "agent/logic/collector/processor/iprocessor"
	"agent/logic/collector/rtdb"
	model2 "agent/logic/collector/rtdb/model"
)

const (
	// MaxWaitTimeMs 采集完成后的最长等待时间
	MaxWaitTimeMs = 60000
	// MinTimeoutMs 采集最短超时时间
	MinTimeoutMs = 2000
	// 首个可用通道索引更新时间
	firstIndexUpdateTime = time.Second * 1
)

var (
	emptyChannel model.Channel
)

type filterLogKey struct {
	ChannelID string
	Command   string
}

// Device 采集设备
type Device struct {
	Info model3.DeviceInfo
	// 驱动设备
	driverDevice driver.IDevice
	sem          *osal2.Semaphore
	wg           sync.WaitGroup
	// 模板协议
	templateProtocol *TemplateProtocol
	// 当前指令包的索引
	// 会重复打开驱动，因此不能在打开驱动时置 0
	packetIndex int
	// 是否尝试打开过驱动
	isDriverOpenCalled bool
	// 虚拟测点
	virtualPoints *virtualpoints.VirtualPoints
	// 当前通道的索引
	currentChannelIndex int
	nextChannelIndex    int
	hasReachEnd         bool
	currentChannel      *model.Channel
	// 仅用于标记设备，无其他用途
	randomMark string
	// 通道数
	channelNumber int
	// 针对每个通道的驱动设备
	driverDeviceForChannels []driver.IDevice
	// 是否已尝试为每个通道打开驱动
	isChannelDriverOpenCalled bool
	// 仅当使用的通道数大于 1 时，嗅探每个通道的通讯状态
	// 否则，使用设备的通讯状态即可
	needProbeChannelCommunication bool
	// 标记当前设备的工作状态，当设备关闭后，停止执行嗅探的 goroutine
	ctx                                      context.Context
	cancel                                   context.CancelFunc
	isFirstAvailableChannelIndexUpdateCalled bool
	firstAvailableIndex                      *model.AvailableChannelIndex
	p                                        model4.Processor

	// channel log
	logFileName string
	logFile     *os.File
	logFileSize int64
	logMux      sync.Mutex
}

// NewDevice 根据设备信息 Info 与模板协议 templateProtocol 创建新的采集设备
func NewDevice(info model3.DeviceInfo, templateProtocol *TemplateProtocol) *Device {
	attrs := map[string]string{
		consts.AttrChannel:    info.ChannelID,
		consts.AttrTemplate:   info.Template,
		consts.AttrAContainer: utils2.GetHostName(),
		consts.AttrDeviceName: info.Name,
		consts.AttrDeviceGid:  string(info.Gid),
		consts.Mozu:           cm.Worker().GetDeviceMozuID(info.Gid),
		consts.MozuId:         cm.Worker().GetDeviceMozuID(info.Gid),
	}
	d := &Device{
		Info:                          info,
		driverDevice:                  nil,
		sem:                           osal2.NewSemaphore(1),
		templateProtocol:              templateProtocol,
		virtualPoints:                 virtualpoints.NewVirtualPoints(info.Gid, attrs, info.Channels),
		packetIndex:                   0,
		currentChannelIndex:           0,
		nextChannelIndex:              0,
		hasReachEnd:                   false,
		randomMark:                    fmt.Sprintf("%04d", rand.Intn(10000)),
		channelNumber:                 len(info.Channels),
		needProbeChannelCommunication: len(info.Channels) > 1,
		firstAvailableIndex:           model.NewAvailableChannelIndex(),
		p:                             nil,
	}
	d.driverDeviceForChannels = make([]driver.IDevice, d.channelNumber)

	d.ctx, d.cancel = context.WithCancel(context.Background())

	if d.Info.WaitTimeMs > MaxWaitTimeMs {
		d.Info.WaitTimeMs = MaxWaitTimeMs
	}
	if d.Info.TimeoutMs < MinTimeoutMs {
		d.Info.TimeoutMs = MinTimeoutMs
	}
	return d
}

// needProbeChannelComm 是否需要探测每个通道的通讯状态
func (d *Device) needProbeChannelComm() bool {
	return d.needProbeChannelCommunication
}

// TemplateName 返回设备使用的模板名称
func (d *Device) TemplateName() string {
	if d == nil {
		return ""
	}
	return d.templateProtocol.GetTemplateName()
}

// ID 返回设备 ID
func (d *Device) ID() definition.DeviceGidType {
	if d == nil {
		return definition.DeviceGidType(0)
	}
	return d.Info.Gid
}

// CurrentChannel 返回当前通道信息
func (d *Device) CurrentChannel() *model.Channel {
	if d.currentChannel == nil {
		return &emptyChannel
	}
	return d.currentChannel
}

// CurrentChannelID 返回当前使用的通道 ID，仅用于阅读
func (d *Device) CurrentChannelID() string {
	if !d.Info.IsMultiChannel() {
		return d.ChannelID()
	}
	return d.CurrentChannel().Name + "@" + d.ChannelID()
}

// ChannelID 返回设备的通道ID
func (d *Device) ChannelID() string {
	if d == nil {
		return ""
	}
	return d.Info.ChannelID
}

// Address 返回设备的通道地址
func (d *Device) Address() string {
	if d == nil {
		return ""
	}
	return d.Info.Address
}

func (d *Device) reachChannelEnd() bool {
	return d.hasReachEnd
}

func (d *Device) tryOpenChannelDrivers() {
	if !d.needProbeChannelComm() {
		return
	}
	if d.isChannelDriverOpenCalled {
		return
	}

	if err := d.doOpenChannelDriver(); err != nil {
		d.Warnf("open channel driver failed: %v", err)
	}

	d.isChannelDriverOpenCalled = true

	d.probeChannelsComm()
}

// doOpenChannelDriver 为每个通道打开相关驱动设备，用于嗅探通道的通讯状态
func (d *Device) doOpenChannelDriver() error {
	driver := d.templateProtocol.GetDriver()
	if driver == nil {
		return errors.New("driver is nil")
	}

	for i := range d.driverDeviceForChannels {
		d.driverDeviceForChannels[i] = driver.CreateDevice(d.Info.Gid, d.Info.Name)
		if d.driverDeviceForChannels[i] == nil {
			return fmt.Errorf("create device error, id: %v, name: %v", d.Info.Gid, d.Info.Name)
		}
		c := &d.Info.Channels[i]
		channelInfo := model.ChannelInfo{
			Name:                c.Name,
			Params:              c.Params,
			Address:             c.Address,
			ProtocolVer:         d.Info.ProtocolVersion,
			TimeoutMs:           d.Info.TimeoutMs,
			ParallelCount:       d.Info.ParallelCount,
			PacketMaxPointCount: d.Info.PacketMaxPointCount,
			ExtendKV:            d.Info.ChannelExtendKV,
			DriverExtend:        d.Info.DriverExtend,
		}
		if r := d.driverDeviceForChannels[i].Open(channelInfo,
			d.templateProtocol.GetCollectPackets()); r != consts.QualityOk {
			return fmt.Errorf("open channel %v error, return code: %v", channelInfo, r)
		}
	}
	return nil
}

// DoRequestNext 执行下一次采集任务
func (d *Device) DoRequestNext() bool {
	if d == nil {
		return false
	}
	if firstIndex, currentIndex, ok := d.needMoveToFirstAvailableIndex(); ok {
		d.Warnf("可用通道索引 %v 小于当前通道索引 %v，切换通道采集，\"%v\" -> \"%v\"",
			firstIndex, currentIndex, d.Info.Channels[firstIndex], d.Info.Channels[currentIndex])
		// 若可用通道索引小于当前通道索引，则切换至可用通道
		d.closeDriverDevice()
		d.tryOpenDriver(firstIndex)
	} else {
		// 否则，尝试打开下一通道进行采集
		d.tryOpenDriver(-1)
	}
	d.tryOpenChannelDrivers()
	// 优先使用主通道，仅主异常时才用备通道
	d.tryUpdateFirstAvailableChannelIndex()
	packets := d.templateProtocol.GetCollectPackets()
	packetsLen := len(packets)
	if d.packetIndex >= packetsLen {
		d.packetIndex = 0
		return false
	}
	currentPacket := packets[d.packetIndex]
	d.clearPointsValue(currentPacket.Points)
	reqStartTime := utils2.GetNowUTCTime()
	if d.driverDevice == nil {
		log.Error("d.driverDevice=nil, packet=%v", currentPacket)
		return false
	}

	quality := consts.QualityUncollected
	var msgStat model3.MessageStatistics
	var isInterrupted bool

	// 表达式计算点无需计算
	if currentPacket.Command == "_expression" {
		quality = consts.QualityOk
		isInterrupted = false
	} else {
		// 执行采集逻辑
		d.LogPacket(SendPrefix, fmt.Sprintf(" addr=%v, cmd=%v", d.Info.Address, currentPacket.Command))
		quality, msgStat = d.driverDevice.Request(d.ctx, currentPacket)
		d.LogPacket(RecvPrefix, fmt.Sprintf(" packets=%v, err=%v", msgStat.RecvPackets, msgStat.ErrLog))

		reqEndTime := utils2.GetNowUTCTime()

		// 更新虚拟统计点
		costTime := reqEndTime.Sub(reqStartTime).Milliseconds()
		d.virtualPoints.AddPeriodCostTime(costTime)

		currentReqSuccess := quality == consts.QualityOk
		d.virtualPoints.AddAndUpdateTimeoutNumber(quality == consts.QualityCmdRespTimeout)
		currentPacket.UpdateStat(currentReqSuccess)
		isInterrupted := d.virtualPoints.UpdateAfterOneRequestFinished(currentReqSuccess, len(currentPacket.Points),
			msgStat.SendCount, msgStat.SuccessCount)
		if isInterrupted {
			d.Infof("%v 通讯中断中, Qua=%v...", d.Info.Name, quality)
		}

		// 解析数据并保存
		if currentReqSuccess {
			d.calculatePointsValue(quality, currentPacket.Points)
		} else {
			d.handleRequestFailure(isInterrupted, packets, quality, currentPacket)
		}
	}

	d.packetIndex++
	if d.packetIndex >= packetsLen {
		d.virtualPoints.UpdateAfterOnePeriodFinished(packetsLen)
		d.resetAfterOnePeriod()
		// 在采集一圈结束时上报指标
		d.virtualPoints.ReportInterruption(isInterrupted)

		return false
	}
	return true
}

func (d *Device) handleRequestFailure(isInterrupted bool, packets model.ListCollectPackets, quality consts.Quality,
	currentPacket *model.CollectProtocolPacket) {
	if isInterrupted {
		points := packets.GetPoints()
		for i := range points {
			points[i].RtVal.Qua = consts.QualityCommDisconnected
		}
		d.calculatePointsValue(quality, points)

		// 关闭当前通道，使用下一个通道进行采集，同时可以清空当前fd，解决各种异常场景
		d.closeDriverDevice()
	} else { // 不需要重复将 currentPacket 写入 rtdb
		if currentPacket.IsTimeout() {
			for i := range currentPacket.Points {
				currentPacket.Points[i].RtVal.Qua = consts.QualityCommDisconnected
			}
			key := filterLogKey{ChannelID: d.Info.ChannelID,
				Command: currentPacket.Command}
			filterLog.Errorf(key, "packet timeout, cmd: %v, return value: %v, failed count: %v, device: %+v, mark: %v",
				currentPacket.Command, quality, currentPacket.GetFailedCount(), d.Info, d.randomMark,
			)
			d.calculatePointsValue(quality, currentPacket.Points)
		}
	}
}

func (d *Device) resetAfterOnePeriod() {
	d.packetIndex = 0
	d.virtualPoints.ResetValueAfterOnePeriod()
}

// calculatePointsValue 处理测点值：缩放、偏移、越界，异常检测（全零）
func (d *Device) calculatePointsValue(requestReturnCode consts.Quality, points model.ListPoints) {
	dataPoints := make(model2.DataPoints, len(points))
	currentTime := utils2.GetNowUTCTimeStamp()
	for i, point := range points {
		attr := &point.Attr
		dataPoints[i].ID = attr.ID
		dataPoints[i].DeviceGiD = d.Info.Gid
		if requestReturnCode == consts.QualityOk {
			dataPoints[i].Rtd.Val = point.RtVal
			switch attr.Type {
			case model.AnalogType:
				if valDesc, ok := (attr.ValDesc).(AnalogValueDesc); ok {
					d.calculateAnalogValue(&valDesc, &dataPoints[i].Rtd.Val)
				} else {
					d.Errorf("type assertion failed, value: %v", attr.ValDesc)
					dataPoints[i].Rtd.Val.Qua = consts.QualityConfigError
				}
			case model.DigitalType:
				if valDesc, ok := (attr.ValDesc).(DigitalValDesc); ok {
					d.calculateDigitalValue(&valDesc, &dataPoints[i].Rtd.Val)
				} else {
					d.Errorf("type assertion failed, value: %v", attr.ValDesc)
					dataPoints[i].Rtd.Val.Qua = consts.QualityConfigError
				}
			case model.EnumType:
				if valDesc, ok := (attr.ValDesc).(EnumValDesc); ok {
					d.calculateEnumValue(&valDesc, &dataPoints[i].Rtd.Val)
				} else {
					d.Errorf("type assertion failed, value: %v", attr.ValDesc)
					dataPoints[i].Rtd.Val.Qua = consts.QualityConfigError
				}
			default:
			}
			//d.p.Do(d.templateProtocol.GetPointCount(), dataPoints[i].Gid, &dataPoints[i].Rtd.Val)
		} else {
			dataPoints[i].Rtd.Val.Tms = currentTime
			dataPoints[i].Rtd.Val.Qua = requestReturnCode
		}
	}
	plugin.ProcessRtd(d.ID(), dataPoints, false)
	rtdb.SetDataPoints(dataPoints, true)
}

// calculateAnalogValue 处理模拟量的缩放、偏移以及越界
func (d *Device) calculateAnalogValue(desc *AnalogValueDesc, val *model2.RTValue) {
	if !desc.ScaleEnable {
		return
	}

	v, err := val.Pv.AsFloat()
	if err != nil {
		val.Qua = consts.QualityValueTypeError
		return
	}

	// -9999, -99999, -99998 等特殊值跳过缩放(当consts.SpecialValueForceScaleKey==1时不跳过)
	if utils2.IsSpecialValue(v) {
		status, ok := d.Info.ChannelExtendKV[consts.SpecialValueForceScaleKey]
		if !ok || status != "1" {
			return
		}
	}

	realVal := v*desc.Scale + desc.Offset
	if config.GetRB().IsSimulationEnable() {
		// 模拟生成的值不做缩放、偏移
		realVal = v
	}

	val.Pv.SetValue(realVal)
}

// calculateDigitalValue 获取状态量(特指布尔型)当前值的描述
func (d *Device) calculateDigitalValue(desc *DigitalValDesc, val *model2.RTValue) {
	v, err := val.Pv.AsBool()
	if err != nil {
		return
	}

	if v {
		if s, ok := (*desc)["val1"]; ok {
			val.Desc = s
		}
		val.Pv.SetValue(1)
		return
	}

	if s, ok := (*desc)["val0"]; ok {
		val.Desc = s
	}
	val.Pv.SetValue(0)
}

// calculateEnumValue 获取枚举量当前值的描述
func (d *Device) calculateEnumValue(desc *EnumValDesc, val *model2.RTValue) {
	v, err := val.Pv.AsString()
	if err != nil {
		return
	}
	if s, ok := (*desc)[v]; ok {
		val.Desc = s
	}
}

// inverseCalculatePointsValue 反向处理测点值：缩放、偏移等，控制功能使用
func (d *Device) inverseCalculatePointsValue(point *model.PointInfo, value *string) consts.Quality {
	attr := &point.Attr
	switch attr.Type {
	case model.AnalogType:
		if valDesc, ok := (attr.ValDesc).(AnalogValueDesc); ok {
			d.inverseCalculateAnalogValue(&valDesc, value)
		} else {
			d.Errorf("type assertion failed, value: %v", attr.ValDesc)
			return consts.QualityConfigError
		}
	default:
		// 其他数据类型(布尔、枚举)不做缩放
		return consts.QualityOk
	}
	return consts.QualityOk
}

// calculateAnalogValue 反向处理：模拟量的缩放、偏移以及越界
func (d *Device) inverseCalculateAnalogValue(desc *AnalogValueDesc, val *string) {
	if !desc.ScaleEnable {
		return
	}

	floatValue, err := strconv.ParseFloat(*val, 32)
	if err != nil {
		log.Error("val to float fail:", err)
		return
	}

	realVal := definition.FloatType(floatValue) - desc.Offset
	realVal = realVal / desc.Scale
	*val = fmt.Sprintf("%0.f", realVal)
}

// clearPointsValue 清除测点值
func (d *Device) clearPointsValue(points model.ListPoints) {
	currentTime := utils2.GetNowUTCTimeStamp()
	for i := range points {
		points[i].RtVal.Tms = currentTime
		points[i].RtVal.Qua = consts.QualityUncollected
		points[i].RtVal.Pv = osal2.NewVariant()
	}
}

// doDriverOpenDevice 打开驱动设备
// 打开驱动设备时不能将 packetIndex 置为 0
// 否则会导致 DoRequestNext 永远为 true
func (d *Device) doDriverOpenDevice(channelIndex int) consts.Quality {
	if d == nil {
		return consts.QualityDriverOpenFailed
	}
	d.virtualPoints.Clear()

	if d.driverDevice == nil {
		driver := d.templateProtocol.GetDriver()
		if driver == nil {
			d.Warn("driver is nil")
			return consts.QualityUncertain
		}
		d.driverDevice = driver.CreateDevice(d.Info.Gid, d.Info.Name)
		if d.driverDevice == nil {
			return consts.QualityDriverOpenFailed
		}
	}

	if channelIndex < 0 {
		d.MoveNextChannel()
	} else {
		d.MoveToChannel(channelIndex)
	}

	c := d.CurrentChannel()
	channelInfo := model.ChannelInfo{
		Name:                c.Name,
		Params:              c.Params,
		Address:             c.Address,
		ProtocolVer:         d.Info.ProtocolVersion,
		TimeoutMs:           d.Info.TimeoutMs,
		ParallelCount:       d.Info.ParallelCount,
		PacketMaxPointCount: d.Info.PacketMaxPointCount,
		ExtendKV:            d.Info.ChannelExtendKV,
		DriverExtend:        d.Info.DriverExtend,
	}

	d.logFileName = logfile.GetPacketLogPath(d.Info.ChannelID)
	return d.driverDevice.Open(channelInfo, d.templateProtocol.GetCollectPackets())
}

// Close 关闭采集设备
func (d *Device) Close() {
	if d == nil {
		return
	}
	d.Infof("close device, stop collecting, channel: \"%v\", template: \"%v\"", d.CurrentChannelID(), d.TemplateName())
	d.cancel()
	d.closeDriverDevice()
	d.remove()
	d.virtualPoints.Close()
	d.logClose()
	d.sem.Post()
	d.wg.Wait()
}

// closeDriverDevice 关闭驱动设备
func (d *Device) closeDriverDevice() {
	d.Infof("close driver, channel id: \"%v\"", d.CurrentChannelID())
	if d.driverDevice != nil {
		d.driverDevice.Close()
	}
	d.driverDevice = nil
	d.isDriverOpenCalled = false
}

func (d *Device) getSubDeviceVirtualPoints() definition.DataPointIDsType {
	templateData := cm.GetCachedTemplateData(d.ID())
	subDeviceVirtualPoints := make(definition.DataPointIDsType, 0, len(templateData.SubDevices))
	for i := range templateData.SubDevices {
		// 如果模板中已存在 CommID，会重复删除该测点，但无影响
		subDeviceVirtualPoints = append(subDeviceVirtualPoints,
			definition.GenerateCommID(templateData.SubDevices[i].DeviceGiD))
	}
	return subDeviceVirtualPoints
}

func (d *Device) remove() {
	subVirtualPoints := d.getSubDeviceVirtualPoints()
	realPointIDs := d.templateProtocol.GetCollectPackets().GetPointIDs()
	virtualPointIDs := d.virtualPoints.GetPoints()
	pointIDs := make(definition.DataPointIDsType, 0, len(subVirtualPoints)+len(realPointIDs)+len(virtualPointIDs))
	pointIDs = append(pointIDs, subVirtualPoints...)
	pointIDs = append(pointIDs, realPointIDs...)
	pointIDs = append(pointIDs, virtualPointIDs...)
	// 从实时数据库中删除设备下的所有测点
	rtdb.ClearDataPoints(pointIDs)
}

// CmdWait 同设备的不同命令字等待一段时间
func (d *Device) CmdWait() {
	if d == nil {
		return
	}
	if d.Info.CmdInterval > 0 {
		d.sem.Wait(time.Millisecond * time.Duration(d.Info.CmdInterval))
	} else {
		d.sem.Wait(time.Millisecond * time.Duration(consts.DefaultSerialDeviceCmdIntervalMs))
	}
}

// Wait 同channel的不同设备等待一段时间
func (d *Device) Wait() {
	if d == nil {
		return
	}
	if d.Info.WaitTimeMs > 0 {
		d.sem.Wait(time.Millisecond * time.Duration(d.Info.WaitTimeMs))
	} else {
		d.sem.Wait(time.Millisecond * time.Duration(consts.DefaultChannelDeviceWaitMs))
	}
}

// Post 取消等待
func (d *Device) Post() {
	if d == nil {
		return
	}
	d.sem.Post()
}

// DoControl 控制测点
func (d *Device) DoControl(pointGid definition.DataPointIDType, val string) int {
	packet := d.templateProtocol.findCtlProtoPacket(pointGid)
	if packet == nil {
		return model3.ControlPointNotFind
	}

	// 计算原有值，即进行缩放等反向操作
	if q := d.inverseCalculatePointsValue(packet.Point, &val); q != 0 {
		return int(q)
	}
	if d.driverDevice == nil {
		return model3.ChannelNotFind
	}
	qua := d.driverDevice.Control(packet, val)
	return int(qua)
}

// DoFreeze 冻结测点
func (d *Device) DoFreeze() {

}

// DeviceRemove 设备移除
func (d *Device) DeviceRemove() {

}
