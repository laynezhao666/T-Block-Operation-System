// Package worker worker
package worker

import (
	"fmt"
	"agent/entity/config"
	"agent/utils/message"
	"agent/utils/osal/queue"
	"runtime"
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/entity/model"
	dev "agent/logic/collector/device"
)

// ListDevice 设备列表
type ListDevice []*dev.Device

// WorkerChannel 工作通道
type WorkerChannel struct {
	channelID              string
	stopCh                 chan int
	devices                ListDevice
	successRequestCountMap *requestCountMap
	q                      *queue.ThreadQueue
	currentDeviceIndex     int
	startCh                chan int
	isRunning              int32
}

// NewWorkerChannel 新建工作通道
func NewWorkerChannel(channelID string) *WorkerChannel {
	return &WorkerChannel{
		channelID:              channelID,
		stopCh:                 make(chan int, 1),
		devices:                make(ListDevice, 0, 1),
		q:                      queue.NewThreadQueue(),
		startCh:                make(chan int, 1),
		successRequestCountMap: NewRequestCountMap(),
		isRunning:              0,
	}
}

// GetDevices 获取设备列表
func (w *WorkerChannel) GetDevices() []definition.DeviceGidType {
	if w == nil {
		return nil
	}

	devices := make([]definition.DeviceGidType, 0, len(w.devices))
	for _, d := range w.devices {
		devices = append(devices, d.ID())
	}
	return devices
}

// GetDevice 获取设备
func (w *WorkerChannel) GetDevice(id definition.DeviceGidType) *dev.Device {
	if w == nil {
		return nil
	}

	for _, d := range w.devices {
		if id == d.ID() {
			return d
		}
	}
	return nil
}

// IsRunning 判断是否正在运行
func (w *WorkerChannel) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadInt32(&w.isRunning) >= 1
}

func (w *WorkerChannel) setRunning(r bool) {
	if w == nil {
		return
	}
	if r {
		atomic.StoreInt32(&w.isRunning, 1)
	} else {
		atomic.StoreInt32(&w.isRunning, 0)
	}
}

// Close 关闭工作通道
func (w *WorkerChannel) Close() {
	if w == nil {
		return
	}
	w.Stop()
	w.q.Clear()
}

// PushMessage 推送消息
func (w *WorkerChannel) PushMessage(msg message.IMessage) {
	if w == nil {
		return
	}
	if !w.IsRunning() {
		w.Start()
	}
	err := w.q.Push(msg)
	if err != nil {
		log.Errorf("push message [%v] error: %v", msg, err)
	}
}

// Start 启动工作通道
func (w *WorkerChannel) Start() {
	w.setRunning(true)
	go w.run()
	<-w.startCh
	for len(w.stopCh) > 0 {
		<-w.stopCh
	}
}

// Stop 停止工作通道
func (w *WorkerChannel) Stop() {
	if w == nil {
		return
	}
	if len(w.stopCh) == 0 {
		w.stopCh <- 1
	}
	w.setRunning(false)
}

func (w *WorkerChannel) run() {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 102400)
			length := runtime.Stack(stack, true)
			log.Errorf("WorkerChannel[%s] panic:%v,stack:%s",
				w.channelID, err, string(stack[:length]))
			panic(err)
		}
	}()

	w.startCh <- 1
	if w == nil {
		return
	}
	w.currentDeviceIndex = 0

	// 采集间隔，默认1秒
	collectionInterval := config.LoadIntOrDefault(config.GetRB().Collector.Common.CollectionInterval, 1000)
	if collectionInterval < 800 {
		collectionInterval = 800 // 最小为0.8秒
	}

	for {
		select {
		case <-w.stopCh:
			// close 设备连接
			for _, device := range w.devices {
				device.Close()
				device.Post()
			}
			return
		case <-time.After(time.Duration(collectionInterval) * time.Millisecond):
		}
		w.doMessageQueue()
		if len(w.devices) == 0 {
			w.setRunning(false)
			return
		}
		w.handleDeviceCollect()

		// 没有采集任务 或者主动关闭时，退出
		if !w.IsRunning() {
			return
		}
	}
}

func (w *WorkerChannel) handleDeviceCollect() {
	if !pool.Acquire() {
		return
	}
	defer pool.Release()

	for ; w.currentDeviceIndex < len(w.devices); w.currentDeviceIndex++ {
		onePeriodFinished := true
		device := w.devices[w.currentDeviceIndex]
		for device.DoRequestNext() {
			if !w.IsRunning() {
				onePeriodFinished = false
				break
			}
			device.CmdWait()
		}
		if onePeriodFinished {
			w.successRequestCountMap.Increase(device.ID())
		}
		if !w.IsRunning() {
			return
		}
		device.Wait()
	}

	if w.currentDeviceIndex >= len(w.devices) {
		w.currentDeviceIndex = 0
	}
}

func (w *WorkerChannel) doMessageQueue() {
	if w == nil {
		return
	}
	var msg message.IMessage
	for {
		t, ok := w.q.Pop()
		if !ok {
			return
		}
		if msg, ok = t.(message.IMessage); !ok {
			return
		}
		switch msg.Topic() {
		case message.TopicDevice:
			if deviceMsg, ok := msg.(*DeviceMessage); ok {
				w.doDeviceMessage(deviceMsg)
			}
		case message.TopicPointControl:
			if controlMsg, ok := msg.(*PointControlMessage); ok {
				w.doPointControlMessage(controlMsg)
			}
		default:
			log.Warnf("error topic msg: \"%v\"", msg.String())
			continue
		}

		switch msg.Pattern() {
		case message.PatternNotice:
			// 无需处理
			break
		default:
		}
	}
}

func (w *WorkerChannel) doDeviceMessage(msg *DeviceMessage) {
	if message.MethodDelete == msg.Method() {
		w.removeDevice(msg.Info.Gid)
	} else {
		// 如果添加设备失败，停止采集任务
		if err := w.addDevice(msg.Info); err != nil {
			log.Warnf("add device error: %v", err)
		}
	}
	// 设备链表变更后，重置 currentDeviceIndex
	w.currentDeviceIndex = 0
}

func (w *WorkerChannel) doPointControlMessage(msg *PointControlMessage) {
	d := w.GetDevice(msg.CtlInfo.DeviceGid)
	if d == nil {
		msg.ReplyCode = model.DeviceNotFind
		return
	}

	msg.ReplyCode = d.DoControl(msg.CtlInfo.PointGid, msg.CtlInfo.Value)
	log.Debugf("do point control:%v, replyCode=%v", msg.CtlInfo, msg.ReplyCode)
}

func (w *WorkerChannel) removeDevice(id definition.DeviceGidType) {
	availableDeviceIndex := 0
	for _, device := range w.devices {
		if device.ID() == id {
			device.Close()
			w.successRequestCountMap.Delete(id)
			log.Infof(
				"delete device: %+v, channel: \"%+v\", address: \"%+v\", template: \"%+v\"",
				device.ID(), device.ChannelID(), device.Address(), device.TemplateName(),
			)
		} else {
			w.devices[availableDeviceIndex] = device
			availableDeviceIndex++
		}
	}
	w.devices = w.devices[:availableDeviceIndex]
}

func (w *WorkerChannel) addDevice(info model.DeviceInfo) error {
	if w == nil {
		return fmt.Errorf("WorkerChannel pointer is nil")
	}
	templateProtocol := TemplateProtocolManager().GetTemplateProtocol(info.Template, info.Gid)
	if templateProtocol == nil {
		return fmt.Errorf("load device template failed: %+v", info)
	}

	if info.ProtocolVersion == "" {
		info.ProtocolVersion = templateProtocol.GetDrvInfo().ProtocolVersion
	}
	if len(templateProtocol.GetDrvInfo().Extend) > 0 {
		info.DriverExtend = templateProtocol.GetDrvInfo().Extend
	}

	device := dev.NewDevice(info, templateProtocol)

	i := 0
	for ; i < len(w.devices); i++ {
		if device.ID() == w.devices[i].ID() {
			break
		}
	}

	if i < len(w.devices) {
		if info.NeedReopen {
			w.devices[i].Close()
			w.devices[i] = device
		}
	} else {
		w.devices = append(w.devices, device)
	}
	log.Infof("add device: %+v", info)
	return nil
}
