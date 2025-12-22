package utils

import (
	"encoding/json"
	"errors"
	"agent/entity/config"
	"agent/entity/definition"
	"agent/entity/model"
	cmodel "agent/logic/collector/device/model"
	"strings"

	cmdbPb "trpcprotocol/cmdb"

	"trpc.group/trpc-go/trpc-go/log"
)

// MergeDeviceOption 合并设备选项
func MergeDeviceOption(devices []model.Device, options []model.Device) []model.Device {
	deviceMap := make(map[definition.DeviceGidType]*model.Device)
	for i := range devices {
		deviceMap[devices[i].Gid] = &devices[i]
	}

	for i := range options {
		device, ok := deviceMap[options[i].Gid]
		if !ok {
			continue
		}
		opt := &options[i]

		if opt.ChData.CmdInterval > 0 {
			device.ChData.CmdInterval = opt.ChData.CmdInterval
		}
		if opt.ChData.WaitTimeMs > 0 {
			device.ChData.WaitTimeMs = opt.ChData.WaitTimeMs
		}
		if opt.ChData.TimeoutMs > 0 {
			device.ChData.TimeoutMs = opt.ChData.TimeoutMs
		}
		if opt.ChData.ParallelCount > 0 {
			device.ChData.ParallelCount = opt.ChData.ParallelCount
		}
		if opt.ChData.PacketMaxPointCount > 0 {
			device.ChData.PacketMaxPointCount = opt.ChData.PacketMaxPointCount
		}
		if len(opt.ChData.Extend) > 0 {
			device.ChData.Extend = opt.ChData.Extend
		}
		if l := len(opt.Extends); l > 0 {
			if (device.Extends) == nil {
				device.Extends = make(map[string]interface{}, l)
			}
			for k, v := range opt.Extends {
				device.Extends[k] = v
			}
		}
	}

	return devices
}

// SimulationCollectAddr 切换模拟采集数据源
func SimulationCollectAddr(devices []model.Device) {
	log.Info("use simulation data!")

	for i := range devices {
		// 替换为模拟器的ip和端口
		// 获取 Channel 和 Tpl 的 Struct
		channel := &devices[i].ChData
		tpl := &devices[i].TemplateData

		if channel.Chtype == "socket" {
			// snmp 模拟
			simulateSnmpAddr := config.GetRB().Test.SnmpAddr
			channel.ChannelID = simulateSnmpAddr

			// modbus tcp 模拟
			//simulateModbusAddr := config.GetRB().Test.ModbusIP
			//channel.ChannelID = simulateModbusAddr + ":502"
		} else {
			// modbus rtu 模拟
			simIP := config.GetRB().Test.ModbusIP
			chid := ""
			switch tpl.TemplateName {
			case "IOT_ACM_YADA_YD6600-N-CT60_V2.0":
				chid = simIP + ":50200"
			case "IOT_ATS_SOCOMEC_ATySC60_V2.0":
				chid = simIP + ":50201"
			case "IOT_BPC_YADA_DEMS-B2V1-TBAI_V2.0":
				chid = simIP + ":50202"
			case "IOT_DCM_YADA_3366P-N-TBAI_V2.0":
				chid = simIP + ":50203"
			case "IOT_HVDC_ZHONGHENG_ZHM20-TBAI_V2.0":
				chid = simIP + ":50204"
			case "IOT_IEAC_SHENLING_JDM_V2.0":
				chid = "/dev/ttyV15" // simIP + ":50205"
			case "IOT_INV_KEHUA_40K-DJN3340-J_V2.0":
				chid = simIP + ":50206"
			case "IOT_THS_TENCENT_TBKH_V2.0": // 串口
				chid = "/dev/ttyV15"
			default:
				continue
			}
			// 设置 chid
			channel.ChannelID = chid
		}
	}

}

// Convert2TemplateData 平铺配置转换为有虚拟子设备的结构
func Convert2TemplateData(rsp *cmdbPb.RspGetCollectorTemplate) (map[string]*model.TemplateData, error) {
	if len(rsp.ConfigMap) == 0 {
		return nil, errors.New("collector config not exist")
	}

	templatesMap := make(map[string]*model.TemplateData, len(rsp.ConfigMap))
	for name, info := range rsp.ConfigMap {
		temOri, _ := json.Marshal(info)
		tem := string(temOri)
		tem = strings.ReplaceAll(tem, "point_name_en", "id")
		tem = strings.ReplaceAll(tem, "point_type", "valtype")
		tem = strings.ReplaceAll(tem, "point_rw", "access")
		tem = strings.ReplaceAll(tem, "reg", "val_key")

		temp := new(model.TemplateData)
		if err := json.Unmarshal([]byte(tem), temp); err != nil {
			log.Errorf("Unmarshal template %s fail: %s", name, err)
			continue
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
		templatesMap[name] = td
	}

	return templatesMap, nil
}
