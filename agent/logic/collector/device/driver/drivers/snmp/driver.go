// Package snmp snmp协议相关
package snmp

import (
	model2 "agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model"
	"agent/logic/collector/device/driver"
	model3 "agent/logic/collector/device/model"
	"agent/utils"
)

const driverName = "snmp"

func init() {
	if err := driver.Register(driverName, DriverSNMP{}); err != nil {
		panic(err)
	}
}

// DriverSNMP struct
type DriverSNMP struct {
}

// Init driver
func (s DriverSNMP) Init() model2.Quality {
	return model2.QualityOk
}

// UnInit driver
func (s DriverSNMP) UnInit() model2.Quality {
	return model2.QualityOk
}

// CreateDevice create device
func (s DriverSNMP) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	device := &DeviceSNMP{
		Data: model.IDeviceData{
			Gid:  gid,
			Name: name,
		},
		isConnected: false,
	}

	return device
}

// CreateValParseObj create val parse obj
func (s DriverSNMP) CreateValParseObj(params *model3.ValParseParams) interface{} {
	p := &SnmpValueParser{
		OID:      params.DataAddr,
		Extend:   params.Extend,
		DataType: utils.GetDataType(params.DataType, nil, nil),
	}
	if len(p.OID) > 0 && p.OID[0] == '.' {
		p.OID = p.OID[1:]
	}

	return p
}
