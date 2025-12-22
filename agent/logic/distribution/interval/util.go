package interval

import (
	"agent/entity/definition"
	"strconv"
)

func getDeviceGid(interval int, point definition.IDPair) definition.DeviceGidType {
	if interval < definition.DefaultInterval {
		// 使用推送间隔作为虚拟的设备ID
		// return definition.DeviceGidType(interval)
		return definition.DeviceGidType(strconv.Itoa(interval))
	}
	return point.DeviceGid
}
