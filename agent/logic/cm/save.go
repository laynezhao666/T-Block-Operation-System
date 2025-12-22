package cm

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model"
	"agent/repo/cm"
	cmUtils "agent/repo/cm/utils"

	"etrpc-go/log"
)

const (
	defaultDeviceName string = "default"
	defaultLocalDir   string = "project/local"
)

// NotifyConfigChange 通知任务变化
func NotifyConfigChange() {
	cm.ConfigChangedChan() <- true
}

// NotifyDeviceConfigChange 通知设备配置变化
func NotifyDeviceConfigChange() {
	cm.DeviceConfigChangedChan() <- true
}

// NotifyStdConfigChange 通知标准测点变化
func NotifyStdConfigChange() {
	cm.StdConfigChangedChan() <- true
}

// SaveCurrentDevicesConfig 保存当前采集设备配置
func (w *worker) SaveCurrentDevicesConfig(saveDeviceType int) error {
	devices := w.CopyAllDevices()
	configMap := make(map[string]any, len(devices))
	switch saveDeviceType {
	default:
		fallthrough
	case definition.CollectorDeviceTypeTBox:
		w.handleTBoxDevices(devices, configMap)
	case definition.CollectorDeviceTypeVendor:
		w.handleVendorDevices(devices, configMap)
	}
	deviceFilePath, err := getLatestDevicePath()
	if err != nil {
		return fmt.Errorf("获取devices*.json文件错误:%s", err.Error())
	}
	err = cmUtils.SaveConfigMapToFile(
		deviceFilePath,
		configMap,
	)
	if err != nil {
		return err
	}
	return nil
}

// handleTBoxDevices 处理TBox设备配置
func (w *worker) handleTBoxDevices(devices map[definition.DeviceGidType]*model.Device, configMap map[string]any) {
	defaultParentDevice := cmUtils.DeviceConfig{
		DeviceCode:    defaultDeviceName,
		DeviceName:    defaultDeviceName,
		DeviceNumber:  defaultDeviceName,
		CollectorType: definition.CollectorDeviceTypeTBox,
		SubDevices:    make([]cmUtils.DeviceConfig, 0, len(devices)),
	}

	for _, d := range devices {
		defaultParentDevice.SubDevices = append(defaultParentDevice.SubDevices,
			w.createDeviceConfig(d, definition.CollectorDeviceTypeTBox))
	}
	configMap[defaultDeviceName] = defaultParentDevice
}

// handleVendorDevices 处理Vendor设备配置
func (w *worker) handleVendorDevices(devices map[definition.DeviceGidType]*model.Device, configMap map[string]any) {
	for _, d := range devices {
		conf := w.createDeviceConfig(d, definition.CollectorDeviceTypeVendor)
		for _, sub := range d.SubDevices {
			conf.SubDevices = append(conf.SubDevices, w.createSubDeviceConfig(sub, d.ChData.ChannelID))
		}
		configMap[d.Name] = conf
	}
}

// createDeviceConfig 创建基础设备配置
func (w *worker) createDeviceConfig(d *model.Device, collectorType int32) cmUtils.DeviceConfig {
	return cmUtils.DeviceConfig{
		DeviceGID:     string(d.Gid),
		DeviceCode:    d.ID,
		DeviceName:    d.Name,
		CollectorType: collectorType,
		MozuId:        int32(d.MozuID),
		Channel:       w.createChannelConfig(d.ChData, d.ChData.ChannelID),
		Tpl:           w.createTemplateConfig(d.TemplateData),
	}
}

// createSubDeviceConfig 创建子设备配置
func (w *worker) createSubDeviceConfig(d model.Device, chId string) cmUtils.DeviceConfig {
	return cmUtils.DeviceConfig{
		DeviceGID:    string(d.Gid),
		DeviceNumber: d.Name,
		// 只有vendor设备走这个逻辑，所以类型是vendor
		CollectorType: definition.CollectorDeviceTypeVendor,
		MozuId:        int32(d.MozuID),
		Channel:       w.createChannelConfig(d.ChData, chId),
		Tpl:           w.createTemplateConfig(d.TemplateData),
	}
}

// createChannelConfig 创建通道配置
func (w *worker) createChannelConfig(chData model.ChannelData, chId string) cmUtils.Channel {
	return cmUtils.Channel{
		Addr: chData.Address,
		// TODO 临时逻辑，将通道id从映射回去，如/usr/dev/serial/com1映射回COM1
		// 子设备的情况下，用上级的channelid
		Chid:        strings.ToUpper(strings.TrimPrefix(chId, "/usr/dev/serial/")),
		Chparams:    chData.ChannelParams,
		Chtype:      chData.Chtype,
		Timeout:     fmt.Sprintf("%v", chData.TimeoutMs),
		CmdInterval: fmt.Sprintf("%v", chData.CmdInterval),
		WaitTime:    fmt.Sprintf("%v", chData.WaitTimeMs),
	}
}

// createTemplateConfig 创建模板配置
func (w *worker) createTemplateConfig(tplData model.TemplateInfo) cmUtils.Template {
	return cmUtils.Template{
		Tplnm:   tplData.TemplateName,
		Tplpath: tplData.TemplatePath,
	}
}

// SaveCurrentTemplatesConfig 保存当前模板配置
func (w *worker) SaveCurrentTemplatesConfig() error {
	templatesMap := w.CopyAllTemplateData()
	configMap := make(map[string]any, len(templatesMap))
	for k, v := range templatesMap {
		configMap[k] = v
	}
	err := cmUtils.SaveConfigMapToFile(
		config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeTemplateFile,
		configMap,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveTemplatesConfig 保存模板配置
func (w *worker) SaveTemplatesConfig(templatesMap map[string]*model.TemplateData) error {
	configMap := make(map[string]any, len(templatesMap))
	for k, v := range templatesMap {
		configMap[k] = v
	}
	err := cmUtils.SaveConfigMapToMultipleFile(
		config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeTemplateDir+"/",
		configMap,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveCurrentStdPointsConfig 保存当前标准测点配置
func (w *worker) SaveCurrentStdPointsConfig() error {
	stdData := w.CopyStdData()
	configMap := make(map[string]any, 1)
	configMap[defaultDeviceName] = stdData.StdPointsInfo
	err := cmUtils.SaveConfigMapToFile(
		config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeStdFile,
		configMap,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTemplateConfig 删除模板配置
func (w *worker) DeleteTemplateConfig(templateName string) error {
	err := cmUtils.DeleteConfigFile(
		config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeTemplateDir+"/",
		templateName,
	)
	if err != nil {
		return err
	}
	return nil
}

// 获取最新的设备路径
func getLatestDevicePath() (string, error) {
	// 获取项目路径
	projectPath := config.GetRB().GetProjectLocalPath()
	configType := consts.DeviceTag

	// 查找匹配的文件
	files, err := filepath.Glob(filepath.Join(projectPath, fmt.Sprintf("%s*.json", configType)))
	if err != nil {
		log.Errorf("failed to find %s*.json files: %v", configType, err)
		return "", fmt.Errorf("failed to find %s*.json files: %v", configType, err)
	}

	if len(files) == 0 {
		log.Errorf("no %s*.json files found", configType)
		return "", fmt.Errorf("no %s*.json files found", configType)
	}

	// 正则表达式匹配文件名中的时间戳
	re := regexp.MustCompile(configType + `@(\d+)\.json`)
	var maxTimestamp int64 = -1
	// 兼容没有时间戳的情况
	targetFile := files[0]

	// 遍历所有匹配的文件，找到时间戳最大的文件
	for _, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 2 { // 匹配成功
			timestamp, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				log.Warnf("invalid timestamp in filename %s: %v", file, err)
				continue
			}

			// 更新最大时间戳的文件
			if timestamp > maxTimestamp {
				maxTimestamp = timestamp
				targetFile = file
			}
		}
	}

	if targetFile == "" {
		log.Errorf("no valid %s @<timestamp>.json file found", configType)
		return "", fmt.Errorf("no %s*.json files found", configType)
	}
	return targetFile, nil
}
