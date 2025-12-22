package plugin

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

// ProcessRtd 处理rtd数据
func ProcessRtd(deviceID definition.DeviceGidType, points model.DataPoints, ignoreSubDeviceComm bool) {
	Manager().processRtd(deviceID, points, ignoreSubDeviceComm)
}
