package device

import (
	"agent/utils"
	"runtime"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/logic/collector/device/model"
	"agent/logic/collector/device/virtualpoints"
)

func (d *Device) tryOpenDriver(channelIndex int) {
	if d.isDriverOpenCalled {
		return
	}
	r := d.doDriverOpenDevice(channelIndex)
	if r != consts.QualityOk {
		// 打开失败则返回失败，改为保持一直尝试
		d.Warnf("open driver fail，channel id: \"%v\", template: \"%v\"", d.CurrentChannelID(), d.TemplateName())
		d.isDriverOpenCalled = false
		return
	}
	d.Infof("open driver success，channel id: \"%v\", template: \"%v\"", d.CurrentChannelID(), d.TemplateName())
	d.isDriverOpenCalled = true
}

func (d *Device) needMoveToFirstAvailableIndex() (int, int, bool) {
	firstAvailableChannelIndex := d.firstAvailableIndex.Get()
	firstIndex := firstAvailableChannelIndex % d.channelNumber
	currentIndex := d.currentChannelIndex % d.channelNumber
	return firstIndex, currentIndex, firstIndex < currentIndex && firstAvailableChannelIndex >= 0
}

func (d *Device) tryUpdateFirstAvailableChannelIndex() {
	if !d.needProbeChannelComm() {
		return
	}
	if d.isFirstAvailableChannelIndexUpdateCalled {
		return
	}

	d.isFirstAvailableChannelIndexUpdateCalled = true

	d.wg.Add(1)
	go d.updateFirstAvailableChannelIndexLoop()
}

func (d *Device) updateFirstAvailableChannelIndex() {
	firstAvailableIndex := -1
	for i := 0; i < d.channelNumber; i++ {
		if !d.virtualPoints.GetChannelInterruptionStatus(i) {
			// 第一个通讯正常的通道
			firstAvailableIndex = i
			break
		}
	}

	// 更新第一个可用通道的索引
	d.firstAvailableIndex.Set(firstAvailableIndex)
}

func (d *Device) updateFirstAvailableChannelIndexLoop() {
	defer d.wg.Done()
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(firstIndexUpdateTime):
			break
		}
		d.updateFirstAvailableChannelIndex()
	}
}

// MoveToChannel 切换到指定 channel
func (d *Device) MoveToChannel(index int) {
	index %= d.channelNumber
	d.currentChannel = &d.Info.Channels[index]
	d.currentChannelIndex = index
	d.nextChannelIndex = d.currentChannelIndex + 1
	d.Warnf("switch to channel: \"%+v\"", *d.CurrentChannel())
}

// MoveNextChannel 切换到下一 channel
func (d *Device) MoveNextChannel() {
	l := len(d.Info.Channels)
	d.currentChannelIndex = d.nextChannelIndex
	if !d.hasReachEnd && d.currentChannelIndex >= l {
		// 标记所有地址均已尝试
		// 用于后续判断任务的失败状态
		d.hasReachEnd = true
	}

	d.currentChannel = &d.Info.Channels[d.currentChannelIndex%l]
	// 仅当有多个地址时才记录日志
	if l > 1 {
		if d.currentChannelIndex == 0 {
			d.Warnf("first use channel: \"%+v\"", *d.CurrentChannel())
		} else {
			d.Warnf("now use channel: \"%+v\"", *d.CurrentChannel())
		}
	}

	d.nextChannelIndex++
	return
}

func (d *Device) probeChannelsComm() {
	for i := range d.driverDeviceForChannels {
		d.wg.Add(1)
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					stack := make([]byte, 102400)
					length := runtime.Stack(stack, true)
					log.Errorf("panic:%v,stack:%s",
						r, string(stack[:length]))
				}
			}()
			defer d.wg.Done()
			d.probeChannelComm(i)
		}(i)
	}
}

// probeChannelComm 定期嗅探各通道的通讯状态
func (d *Device) probeChannelComm(index int) {
	chDev := d.driverDeviceForChannels[index]
	packets := d.templateProtocol.GetCollectPackets()
	var packet model.CollectProtocolPacket
	packetNum := len(packets)
	packetIndex := 0
	failedCount := 0
	isInterruption := false
	communicationStatusChangedTime := time.Now()
	currentSuccess := true
	channelName := d.Info.Channels[index].Name

	d.Infof("start probe channel: \"%v\"", channelName)
	defer func() {
		d.Infof("stop probe channel: \"%v\"", channelName)
	}()

	for {
		if packetNum > 0 {
			packet = *packets[packetIndex]
			packetIndex++
			packetIndex %= packetNum
		}
		if chDev != nil {
			currentSuccess = chDev.RequestPing(d.ctx, packet) == consts.QualityOk
		}
		if currentSuccess {
			failedCount = 0
		} else {
			failedCount++
		}

		if failedCount > 0 {
			if virtualpoints.IsChannelCommunicationInterruption(failedCount) {
				if !isInterruption {
					d.Warnf("通道: \"%v\" %v", channelName, virtualpoints.CommunicationInterruptionStr)
				}
				isInterruption = true
				communicationStatusChangedTime = utils.GetNowUTCTime()
				d.virtualPoints.UpdateChannelInterruptionStatus(isInterruption, index,
					communicationStatusChangedTime.Unix())
			}
		} else {
			if isInterruption {
				d.Infof("通道: \"%v\" %v", channelName, virtualpoints.CommunicationNormalStr)
			}
			isInterruption = false
			communicationStatusChangedTime = utils.GetNowUTCTime()
			d.virtualPoints.UpdateChannelInterruptionStatus(isInterruption, index,
				communicationStatusChangedTime.Unix())
		}

		// 因多通道的设备不多，故每秒上报数据量不大
		d.virtualPoints.ReportChannelInterruption(index, isInterruption)

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 6):
			break
		}
	}
}
