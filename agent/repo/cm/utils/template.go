package utils

import (
	"encoding/json"
	"agent/entity/definition"
	"agent/entity/model"
	cmodel "agent/logic/collector/device/model"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
)

// ParseCollectTemplateInfo 解析模版配置
func ParseCollectTemplateInfo(info any) (*model.TemplateData, error) {
	temOri, _ := json.Marshal(info)
	tem := string(temOri)
	tem = strings.ReplaceAll(tem, "point_name_en", "id")
	tem = strings.ReplaceAll(tem, "point_name_zh", "name")
	tem = strings.ReplaceAll(tem, "point_type", "valtype")
	tem = strings.ReplaceAll(tem, "point_rw", "access")
	tem = strings.ReplaceAll(tem, "reg", "val_key")

	temp := new(model.TemplateData)
	if err := json.Unmarshal([]byte(tem), temp); err != nil {
		log.Errorf("unmarshal template fail: %s", err)
		return nil, err
	}
	sub2info := map[string]model.SubDeviceData{}
	for _, point := range temp.PointsInfo {
		subDeviceName := SELF
		if len(point.SubDevice) > 0 {
			subDeviceName = point.SubDevice
		}
		if sub, ok := sub2info[subDeviceName]; ok {
			sub.PointsInfo = append(sub.PointsInfo, point)
			sub2info[subDeviceName] = sub
		} else {
			sub2info[subDeviceName] = model.SubDeviceData{
				PointsInfo: []cmodel.TemplateInstancePointInfo{point},
			}
		}
	}
	td := new(model.TemplateData)
	td.DrvInfo = temp.DrvInfo
	for sub, data := range sub2info {
		if sub == SELF {
			td.PointsInfo = data.PointsInfo
		} else {
			// 替换真实子设备数据
			data.InstanceDeviceGid = definition.DeviceGidType(sub)
			td.SubDevices = append(td.SubDevices, data)
		}
	}

	return td, nil
}
