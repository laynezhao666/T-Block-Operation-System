package utils

import (
	"agent/entity/config"
	"agent/entity/definition"
	"agent/logic/task"
	"agent/utils/file/io"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	relativeOtherConfigFile string = "other.json"
	GeneratedDeviceGidKey   string = "generated_device_gid"
)

// GetAlternativeDeviceGid 获取备选设备GID
func GetAlternativeDeviceGid() definition.DeviceGidType {
	otherConfigContent := make(map[string]any)
	err := io.JSON.Read(config.GetRB().GetProjectLocalPath()+"/"+relativeOtherConfigFile, &otherConfigContent)
	if err != nil {
		log.Errorf("read local other config err: %v", err)
		return definition.DeviceGidType("0")
	}
	gid, ok := otherConfigContent[GeneratedDeviceGidKey]
	if !ok {
		log.Error("device gid not found")
		return definition.DeviceGidType("0")
	}
	return definition.DeviceGidType(gid.(string))
}

// SetAlternativeDeviceGid 设置备选设备GID
func SetAlternativeDeviceGid(gid definition.DeviceGidType) {
	otherConfigContent := make(map[string]any)
	err := io.JSON.Read(config.GetRB().GetProjectLocalPath()+"/"+relativeOtherConfigFile, &otherConfigContent)
	if err != nil {
		log.Errorf("read local other config err: %v", err)
	}
	otherConfigContent[GeneratedDeviceGidKey] = gid
	err = io.JSON.Write(config.GetRB().GetProjectLocalPath()+"/"+relativeOtherConfigFile, otherConfigContent)
	if err != nil {
		log.Errorf("write local other config err: %v", err)
	}
}

// GetTargetDevice 获取目标设备
func GetTargetDevice() []string {
	var deviceNums []string
	// 获取任务模块的最新数据
	if config.GetRB().IsDevTaskLocalEnable() {
		// 任务使用测试数据
		simulationDevices := config.GetRB().Task.Local.Devs
		deviceNums = simulationDevices
	} else {
		// 获取任务模块的最新数据
		deviceNums = task.GetInstance().GetTasks()
	}
	return deviceNums
}
