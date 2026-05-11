// Package iec103_siemens 西门子IEC103协议驱动
package iec103_siemens

import (
	"fmt"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"
)

const iec103SiemensDriverName = "iec103"

func init() {
	if err := driver.Register(iec103SiemensDriverName,
		&Driver{drvLib: iec103SiemensDriverName, devices: make(map[definition.DeviceGidType]*Device)}); err != nil {
		panic(fmt.Sprintf("register iec103_siemens driver failed: %v", err))
	}
}

// Driver IEC103西门子协议驱动
type Driver struct {
	drvLib      string
	devicesLock sync.RWMutex
	devices     map[definition.DeviceGidType]*Device
}

// Init 初始化驱动
func (d *Driver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d *Driver) UnInit() consts.Quality {
	d.devicesLock.Lock()
	defer d.devicesLock.Unlock()
	for _, device := range d.devices {
		device.Close()
	}
	d.devices = make(map[definition.DeviceGidType]*Device)
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d *Driver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	d.devicesLock.Lock()
	defer d.devicesLock.Unlock()
	if device, exists := d.devices[gid]; exists {
		log.Warnf("device %s %s already exists", gid, name)
		return device
	}
	device := NewDevice(gid, name)
	d.devices[gid] = device
	return device
}

// CreateValParseObj 创建解析对象
func (d *Driver) CreateValParseObj(params *model.ValParseParams) interface{} {
	return &ValueParser{
		Addr:     params.DataAddr,
		Extend:   params.Extend,
		DataType: params.DataType,
	}
}

// GetDevice 获取设备（用于测试）
func (d *Driver) GetDevice(gid definition.DeviceGidType) *Device {
	d.devicesLock.RLock()
	defer d.devicesLock.RUnlock()
	if value, ok := d.devices[gid]; ok {
		return value
	}
	return nil
}
