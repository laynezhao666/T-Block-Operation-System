// Package dispatcher 采集任务调度器
package dispatcher

import (
	"fmt"
	"agent/utils/message"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/entity/model"
	"agent/logic/cm"
	"agent/logic/collector/worker"
	"agent/logic/distribution/interval"
)

var (
	d    dispatcher
	once sync.Once
)

// WorkerChannelMap 设备通道映射
type WorkerChannelMap map[string]*worker.WorkerChannel

// ChannelDevMap 设备通道映射
type ChannelDevMap map[definition.DeviceGidType]string

type dispatcher struct {
	workerChannels WorkerChannelMap
	sync.Mutex

	channelDevs ChannelDevMap
}

// RunningStatus 当前程序的运行状态
type RunningStatus struct {
	Devices   []definition.DeviceGidType `json:"devices"`
	IsRunning bool                       `json:"is_running"`
}

func init() {
	once.Do(func() {
		d = dispatcher{
			workerChannels: make(WorkerChannelMap),
			channelDevs:    make(ChannelDevMap),
		}
	})
}

// Dispatcher 采集任务调度器
func Dispatcher() *dispatcher {
	return &d
}

// GetStatus 获取当前channel的运行状态
func (d *dispatcher) GetStatus() map[string]RunningStatus {
	if d == nil {
		return nil
	}
	d.Lock()
	defer d.Unlock()

	r := make(map[string]RunningStatus)
	for chID, ch := range d.workerChannels {
		r[chID] = RunningStatus{
			Devices:   ch.GetDevices(),
			IsRunning: ch.IsRunning(),
		}
	}
	return r
}

// AddDevices 添加设备开始采集
func (d *dispatcher) AddDevices(devices []model.Device) {
	if d == nil {
		return
	}
	for _, device := range devices {
		msg := worker.NewDeviceMessage(message.MethodPost, device.GetDeviceInfo())
		d.dispatch(msg)

		d.channelDevs[device.Gid] = msg.Info.ChannelID
	}
}

// DeleteDevicesInChannel 删除设备
func (d *dispatcher) DeleteDevicesInChannel(deviceGids definition.DeviceGidArrType, channelID string) {
	if d == nil {
		return
	}
	for _, gid := range deviceGids {
		msg := worker.NewDeviceDeleteInChannelMessage(gid, channelID)
		d.dispatchInChannel(msg)
		log.Infof("delete device \"%v\"@\"%v\"", gid, channelID)
		go interval.CollectProcessorManager().DeleteDevice(gid)
	}
}

// Load 加载设备
func (d *dispatcher) Load() error {
	if d == nil {
		return nil
	}
	devices := cm.Worker().GetAllDevices()
	d.AddDevices(devices)
	return nil
}

// Unload 卸载设备
func (d *dispatcher) Unload() {
	if d == nil {
		return
	}
	d.Lock()
	defer d.Unlock()

	for _, workerChannel := range d.workerChannels {
		workerChannel.Close()
	}
	d.workerChannels = nil
}

// Reload 重新加载设备
func (d *dispatcher) Reload() error {
	d.Unload()
	return d.Load()
}

func (d *dispatcher) dispatchInChannel(msg *worker.DeviceMessage) {
	if d == nil || msg == nil || len(msg.Info.ChannelID) == 0 {
		return
	}
	d.Lock()
	defer d.Unlock()

	ch, ok := d.workerChannels[msg.Info.ChannelID]
	if !ok {
		log.Warnf("not find device \"%v\" in channel \"%v\"", msg.Info.Gid, msg.Info.ChannelID)
		return
	}

	ch.PushMessage(msg)
}

func (d *dispatcher) dispatch(msg *worker.DeviceMessage) {
	if d == nil || msg == nil {
		return
	}
	d.Lock()
	defer d.Unlock()

	if d.workerChannels == nil {
		d.workerChannels = make(WorkerChannelMap)
	}
	workerChannel, ok := d.workerChannels[msg.Info.ChannelID]
	if !ok && len(msg.Info.ChannelID) > 0 {
		workerChannel = worker.NewWorkerChannel(msg.Info.ChannelID)
		d.workerChannels[msg.Info.ChannelID] = workerChannel
	}

	switch msg.Method() {
	case message.MethodPut, message.MethodPost:
		// 修改设备时，通道有可能发生变化，旧通道的该设备位于其它 MapWorkerChannel 中，需将其删除
		// POST 请求和 PUT 请求均有可能修改设备
		d.removeOldDevice(msg, workerChannel)
	case message.MethodDelete:
		// 遍历通道，删除设备
		d.removeOldDevice(msg, &worker.WorkerChannel{})
	default:
		// nop
	}

	if len(msg.Info.ChannelID) > 0 {
		workerChannel.PushMessage(msg)
	}
}

func (d *dispatcher) removeOldDevice(msg *worker.DeviceMessage, workerChannel *worker.WorkerChannel) {
	if d == nil || msg == nil || workerChannel == nil {
		return
	}
	for chID, channel := range d.workerChannels {
		if channel == workerChannel {
			continue
		}
		removeMessage := worker.NewDeviceMessage(message.MethodDelete, msg.Info)
		if chID == removeMessage.Info.ChannelID {
			log.Infof("channel id: \"%v\", remove message: %v, info: %+v", chID, removeMessage, removeMessage.Info)
		}
		channel.PushMessage(removeMessage)
	}
}

// ControlPoint 控制点位
func (d *dispatcher) ControlPoint(info model.PointControlInfo) (int, string) {
	// 获取点位所在的channel
	chid, ok := d.channelDevs[info.DeviceGid]
	if !ok {
		errMsg := fmt.Sprintf("Control failed, device not found. device id: \"%s\", point id: \"%s\"",
			info.DeviceId, info.PointNo)
		log.Errorf(errMsg)
		return model.DeviceNotFind, errMsg
	}
	workerChannel, ok := d.workerChannels[chid]

	// 给channel发控制消息
	msg := worker.NewPointControlMessage(message.MethodPut, info)
	workerChannel.PushMessage(msg)

	return 0, ""
}
