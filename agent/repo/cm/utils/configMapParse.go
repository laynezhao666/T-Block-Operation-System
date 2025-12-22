package utils

import (
	"encoding/json"
	"fmt"
	"agent/entity/config"
	"agent/entity/model"
	"regexp"
	"strconv"
	"strings"

	"agent/entity/definition"
	cmodel "agent/logic/collector/device/model"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	SELF string = "self"
)

// Response 响应体，方便解析用的嵌套
type Response struct {
	// Code    int    `json:"code"`
	// Message string `json:"message"`
	Data Data `json:"data"`
}

// Data 配置
type Data struct {
	ConfigMap map[string]any `json:"config_map"`
}

// DeviceConfigMap 采集设备配置configMap
type DeviceConfigMap struct {
	ConfigMap map[string]DeviceConfig `json:"config_map"`
}

// DeviceConfig 采集设备配置
type DeviceConfig struct {
	DeviceGID             string         `json:"device_gid"`
	DeviceNumber          string         `json:"device_number"`
	DeviceCode            string         `json:"device_code"`
	DeviceName            string         `json:"device_name"`
	CollectorType         int32          `json:"collector_type"`
	DeviceTypeEn          string         `json:"device_type_en"`
	DeviceTypeZh          string         `json:"device_type_zh"`
	MozuId                int32          `json:"mozu_id"`
	Channel               Channel        `json:"channel"`
	Tpl                   Template       `json:"tpl"`
	SubDevices            []DeviceConfig `json:"sub_devices"`
	ConcurrentNum         string         `json:"concurrent_num"`
	MaxPointNum           string         `json:"max_point_num"`
	ExtendParams          string         `json:"extend_params"`
	IsInstantiation       string         `json:"is_instantiation"`
	BelongCollectorDevice string         `json:"belong_collector_device"`
}

// Channel 采集通道配置
type Channel struct {
	Addr         string `json:"addr"`
	Chid         string `json:"chid"`
	Chparams     string `json:"chparams"`
	Chtype       string `json:"chtype"`
	CmdInterval  string `json:"cmd_interval"`
	MaxFailCount string `json:"max_fail_count"`
	MaxFailTime  string `json:"max_fail_time"`
	Timeout      string `json:"timeout"`
	WaitTime     string `json:"wait_time"`
}

// Template 采集模板配置
type Template struct {
	Tplnm   string `json:"tplnm"`
	Tplpath string `json:"tplpath"`
}

// PointConfigMap 标准点配置configMap
type PointConfigMap struct {
	ConfigMap map[string]PointConfig `json:"config_map"`
}

// PointConfig 标准点配置
type PointConfig struct {
	DevicePoints []any `json:"device_points"`
}

// ParseCollectDeviceConfigMap 从configMap解析采集设备配置
func ParseCollectDeviceConfigMap(configMap map[string]any) ([]model.Device, []model.Device, error) {
	if len(configMap) == 0 {
		// return nil, errors.New("devices config not exist")
		log.Warnf("devices config not exist")
		return []model.Device{}, []model.Device{}, nil
	}
	collectDeviceConfigs := make([]any, 0, 1)
	tboxDeviceConfigs := make([]any, 0, len(configMap))
	for taskDevice, info := range configMap {
		var conf DeviceConfig
		b, err := json.Marshal(info)
		if err != nil {
			log.Warnf("marshal device config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, &conf)
		if err != nil {
			log.Warnf("unmarshal device config err: %v", err)
			continue
		}
		switch conf.CollectorType {
		case definition.CollectorDeviceTypeTBox:
			tboxDeviceConfigs = append(tboxDeviceConfigs, conf)
			for i := range conf.SubDevices {
				conf.SubDevices[i].BelongCollectorDevice = taskDevice
				collectDeviceConfigs = append(collectDeviceConfigs, conf.SubDevices[i])
			}
		case definition.CollectorDeviceTypeVendor:
			conf.BelongCollectorDevice = taskDevice
			collectDeviceConfigs = append(collectDeviceConfigs, conf)
		default:
			log.Errorf("unsupported collector type: %v, device config: %v", conf.CollectorType, conf)
		}
	}

	collectDevices, err := convertDeviceConfigs(collectDeviceConfigs)
	if err != nil {
		return nil, nil, fmt.Errorf("convert collect devices failed: %v", err)
	}
	tboxDevices, err := convertDeviceConfigs(tboxDeviceConfigs)
	if err != nil {
		return nil, nil, fmt.Errorf("convert tbox devices failed: %v", err)
	}

	return collectDevices, tboxDevices, nil
}

// convertDeviceConfigs 将原始配置中一些字段名做替换
func convertDeviceConfigs(deviceConfigs []any) ([]model.Device, error) {
	collectDevicesJson, err := json.Marshal(deviceConfigs)
	if err != nil {
		fmt.Printf("Error marshaling devices: %v\n", err)
		return nil, err
	}

	devicesOri := string(collectDevicesJson)
	// 字段替换
	devs := strings.ReplaceAll(devicesOri, "device_gid", "gid")
	devs = strings.ReplaceAll(devs, "device_code", "id")
	devs = strings.ReplaceAll(devs, "device_name", "name")
	// 将 timeout 和 wait_time 字段转换为int
	re := regexp.MustCompile(`"(timeout|wait_time|cmd_interval)":\s*"(.*?)"`)
	devs = re.ReplaceAllStringFunc(devs, removeQuotesAndConvert)

	var devices []model.Device
	err = json.Unmarshal([]byte(devs), &devices)
	if err != nil {
		return nil, err
	}
	// 串口映射为实际设置
	for i := range devices {
		if devices[i].ChData.Chtype == definition.ChannelTypeSerial {
			// 读取配置，例如 COM1 => /usr/dev/serial/com1
			oriCom := devices[i].ChData.ChannelID
			if comConfig, ok := config.GetRB().Collector.Modbus.SerialsMap.COMs[oriCom]; ok {
				devices[i].ChData.ChannelID = comConfig.Dev
			} else {
				log.Errorf("serial device com mapping fail, channel id = %v", oriCom)
			}
		}
		// 如果cmdb无设备中文名称device_name，则临时使用英文标识部分
		if len(devices[i].Name) == 0 {
			devices[i].Name = devices[i].ID
		}
	}
	return devices, nil
}

// ParseCollectTemplateConfigMap 从configMap解析模版配置
func ParseCollectTemplateConfigMap(configMap map[string]any) (map[string]*model.TemplateData, error) {
	if len(configMap) == 0 {
		// return nil, errors.New("templates config not exist")
		log.Warnf("templates config not exist")
		return nil, nil
	}
	templatesMap := make(map[string]*model.TemplateData, len(configMap))
	for name, info := range configMap {
		temOri, _ := json.Marshal(info)
		tem := string(temOri)
		tem = strings.ReplaceAll(tem, "point_name_en", "id")
		tem = strings.ReplaceAll(tem, "point_name_zh", "name")
		tem = strings.ReplaceAll(tem, "point_type", "valtype")
		tem = strings.ReplaceAll(tem, "point_rw", "access")
		tem = strings.ReplaceAll(tem, "reg", "val_key")

		temp := new(model.TemplateData)
		if err := json.Unmarshal([]byte(tem), temp); err != nil {
			log.Errorf("unmarshal template %s fail: %s", name, err)
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

// ParseStdPointConfigMap 从configMap解析标准测点配置
func ParseStdPointConfigMap(configMap map[string]any) (*cmodel.StdInstancePointsInfo, error) {
	stdPoints := new(cmodel.StdInstancePointsInfo)
	if len(configMap) == 0 {
		log.Warnf("std points config not exist")
		return stdPoints, nil
	}
	var allDevicePoints []any
	for _, info := range configMap {
		var conf PointConfig
		b, err := json.Marshal(info)
		if err != nil {
			log.Warnf("marshal std point config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, &conf)
		if err != nil {
			log.Warnf("unmarshal std point config err: %v", err)
			continue
		}
		allDevicePoints = append(allDevicePoints, conf.DevicePoints...)
	}
	stdJson, _ := json.Marshal(allDevicePoints)

	err := json.Unmarshal([]byte(stdJson), stdPoints)
	if err != nil {
		log.Error("std point config: unmarshal fail:%v", err)
		return nil, err
	}
	return stdPoints, nil
}

// removeQuotesAndConvert 函数用于去掉引号并将字符串转换为数字。如果转换失败，则返回 0
func removeQuotesAndConvert(match string) string {
	re := regexp.MustCompile(`"(timeout|wait_time|cmd_interval)":\s*"(.*?)"`)
	matches := re.FindStringSubmatch(match)
	if len(matches) > 2 {
		// 尝试将字符串值转换为数字
		if num, err := strconv.Atoi(matches[2]); err == nil {
			// 返回替换后的字符串
			return fmt.Sprintf("\"%s\": %d", matches[1], num)
		}
	}
	// 如果转换失败，返回 0
	return fmt.Sprintf("\"%s\": 0", matches[1])
}

// ParseStdDeviceConfigMap 从configMap解析标准设备配置
func ParseStdDeviceConfigMap(configMap map[string]any) ([]model.StdDevice, error) {
	if len(configMap) == 0 {
		log.Warnf("std devices config not exist")
		return nil, nil
	}
	// stdDevicesMap := make(map[string]model.StdDevice, len(configMap))
	stdDevicesList := []model.StdDevice{}
	for _, info := range configMap {
		l := struct {
			List []model.StdDevice `json:"list"`
		}{}
		b, err := json.Marshal(info)
		if err != nil {
			log.Warnf("marshal std device config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, &l)
		if err != nil {
			log.Warnf("unmarshal std device config err: %v", err)
			continue
		}
		stdDevicesList = append(stdDevicesList, l.List...)
	}
	return stdDevicesList, nil
}

// ParseCmdbVersionConfigMap 从configMap解析cmdb版本配置
func ParseCmdbVersionConfigMap(configMap map[string]any) (map[string]*model.ConfigVersion, error) {
	if len(configMap) == 0 {
		log.Warnf("cmdb version config not exist")
		return nil, nil
	}
	versionMap := make(map[string]*model.ConfigVersion, len(configMap))
	for k, v := range configMap {
		version := &model.ConfigVersion{}
		b, err := json.Marshal(v)
		if err != nil {
			log.Warnf("marshal cmdb version device config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, version)
		if err != nil {
			log.Warnf("unmarshal cmdb version device config err: %v", err)
			continue
		}
		versionMap[k] = version
	}
	return versionMap, nil
}
