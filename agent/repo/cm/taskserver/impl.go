package taskserver

import (
	"agent/entity/consts"
	"agent/entity/model"
	model2 "agent/logic/collector/device/model"
	"agent/repo/cm/utils"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	cmdbPb "trpcprotocol/cmdb"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	paramNumberPerConfigFetch = 5
)

// ReadImpl 读取配置
type ReadImpl struct {
	chConfigChanged chan bool
	cmdbProxy       cmdbPb.ConfigQueryClientProxy
}

// NewReadImpl 初始化
func NewReadImpl(chConfigChanged chan bool) *ReadImpl {
	r := &ReadImpl{
		chConfigChanged: chConfigChanged,
		cmdbProxy:       cmdbPb.NewConfigQueryClientProxy(),
	}

	// todo 注册task server watcher回调
	return r
}

// GetStdDevice 获取标准设备
func (r *ReadImpl) GetStdDevice(stdMap map[string]bool, deviceNums []string) (*model.StdDeviceData, error) {
	stdDevice := &model.StdDeviceData{
		StdDevices:      make([]model.StdDevice, 0),
		StdDeviceMap:    make(map[string]model.StdDevice),
		StdPoints:       make(map[string]model2.StdInstancePointsInfo),
		ConciseCodeMap:  make(map[string]string),
		DeviceNumberMap: make(map[string]string),
	}
	//deviceNums := utils.GetTargetDevice()
	for _, deviceNum := range deviceNums {
		req := &cmdbPb.ReqGetDeviceEntity{
			BelongCollector: deviceNum,
		}
		rsp, err := r.cmdbProxy.GetDeviceEntity(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("tlink std device config fetch fail: %v", err)
			return nil, err
		}
		list := make([]model.StdDevice, 0)
		for _, device := range rsp.GetList() {
			list = append(list, model.StdDevice{
				DeviceGid:               device.DeviceGid,
				DeviceNumber:            device.DeviceNumber,
				DeviceNumberShow:        device.DeviceNumberShow,
				DeviceNo:                device.DeviceNo,
				DeviceName:              device.DeviceName,
				MozuId:                  device.MozuId,
				MozuName:                device.MozuName,
				IdcArea:                 device.IdcArea,
				FuncRoom:                device.FuncRoom,
				ParentDeviceNumber:      device.ParentDeviceNumber,
				EnableStatus:            device.EnableStatus,
				DeviceTypeEn:            device.DeviceTypeEn,
				DeviceTypeZh:            device.DeviceTypeZh,
				ApplicationTypeEn:       device.ApplicationTypeEn,
				ApplicationTypeZh:       device.ApplicationTypeZh,
				BelongApplicationTypeEn: device.BelongApplicationTypeEn,
			})
		}
		// 展示设备编号处理
		list = utils.AddShowDeviceNumber(list)
		// 获取短编号索引
		codeMap, list := utils.GetConciseCodeMap(list)

		// 写本地文件(目前没有按采集设备分组)
		err = utils.SaveConfigListToDirFile(nil, deviceNums)
		if err != nil {
			return nil, fmt.Errorf("save std device config to file fail: %v", err)
		}
		stdDevice.StdDevices = append(stdDevice.StdDevices, list...)
		for _, d := range list {
			stdDevice.StdDeviceMap[d.DeviceGid] = d
			if d.DeviceNumber != "" {
				stdDevice.DeviceNumberMap[d.DeviceNumber] = d.DeviceGid
			}
		}
		for k, v := range codeMap {
			stdDevice.ConciseCodeMap[k] = v
		}
	}
	return stdDevice, nil
}

// GetCmdbVersion 获取cmdb版本
func (r *ReadImpl) GetCmdbVersion() (map[string]*model.ConfigVersion, error) {
	versionMap := make(map[string]*model.ConfigVersion, 0)
	deviceNums := utils.GetTargetDevice()
	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		req := &cmdbPb.ReqGetCollectorDevice{
			DeviceNumbers: deviceNums[begin:end],
		}
		rsp, err := r.cmdbProxy.GetConfigModifyTime(trpc.BackgroundContext(), req)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(rsp.GetConfigMap())
		if err != nil {
			return nil, fmt.Errorf("marshal cmdb version config map err: %v", err)
		}
		var configMap map[string]any
		err = json.Unmarshal(b, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal cmdb version config map err: %v", err)
		}
		// var tempMap map[string]*model.ConfigVersion
		tempMap, err := utils.ParseCmdbVersionConfigMap(configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal cm version map err: %v", err)
		}
		for k, v := range tempMap {
			versionMap[k] = v
		}
	}
	return versionMap, nil
}

// GetDevices 获取采集设备列表
func (r *ReadImpl) GetDevices(devices []string) ([]model.Device, []model.Device, map[string]any, error) {
	deviceMap := make(map[string]any, 0)
	deviceNums := utils.GetTargetDevice()
	totalCollectDevices := make([]model.Device, 0)
	totalTboxDevices := make([]model.Device, 0)
	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		req := &cmdbPb.ReqGetCollectorDevice{
			DeviceNumbers: deviceNums[begin:end],
		}
		rsp, err := r.cmdbProxy.GetCollectorDevice(trpc.BackgroundContext(), req)
		if err != nil {
			// log.Errorf("task server device config fetch fail: %v", err)
			return nil, nil, deviceMap, err
		}
		b, err := json.Marshal(rsp.GetConfigMap())
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("marshal device config map err: %v", err)
		}
		var configMap map[string]any
		err = json.Unmarshal(b, &configMap)
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("unmarshal device config map err: %v", err)
		}
		collectDevices, tboxDevices, err := utils.ParseCollectDeviceConfigMap(configMap)
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("task server device config parse fail: %v", err)
		}
		totalCollectDevices = append(totalCollectDevices, collectDevices...)
		totalTboxDevices = append(totalTboxDevices, tboxDevices...)
	}

	log.Infof("task server fetch ok: device config; device count: %v", len(totalCollectDevices))

	return totalCollectDevices, totalTboxDevices, deviceMap, nil
}

// GetTemplate 获取采集模板
func (r *ReadImpl) GetTemplate(fullName string) (*model.TemplateData, error) {
	_, fileName := filepath.Split(fullName)

	req := &cmdbPb.ReqGetCollectorTemplate{
		TemplateNames: []string{fileName},
	}
	rsp, err := r.cmdbProxy.GetCollectorTemplate(trpc.BackgroundContext(), req)
	if err != nil {
		log.Errorf("task server template config fetch fail: %v", err)
		return nil, err
	}
	b, err := json.Marshal(rsp.GetConfigMap())
	if err != nil {
		return nil, fmt.Errorf("marshal template config map err: %v", err)
	}
	var configMap map[string]any
	err = json.Unmarshal(b, &configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal template config map err: %v", err)
	}
	templateMap, err := utils.ParseCollectTemplateConfigMap(configMap)
	if err != nil {
		return nil, fmt.Errorf("task server template config parse fail: %v", err)
	}
	tem, ok := templateMap[fileName]
	if !ok {
		return nil, errors.New("template not exist")
	}

	log.Infof("task erver fetch ok: single template config")

	return tem, nil
}

// GetTemplates 获取采集模板
func (r *ReadImpl) GetTemplates(list []string) (map[string]any, error) {
	if len(list) == 0 {
		return nil, errors.New("template path empty")
	}

	var nameList []string
	for _, v := range list {
		_, fileName := filepath.Split(v)
		nameList = append(nameList, fileName)
	}
	templateMap := make(map[string]any)
	for begin := 0; begin < len(nameList); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(nameList) {
			end = len(nameList)
		}
		req := &cmdbPb.ReqGetCollectorTemplate{
			TemplateNames: nameList[begin:end],
		}
		rsp, err := r.cmdbProxy.GetCollectorTemplate(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("task server template config fetch fail: %v", err)
			return nil, err
		}
		b, err := json.Marshal(rsp.ConfigMap)
		if err != nil {
			return nil, fmt.Errorf("marshal template config map err: %v", err)
		}
		var configMap map[string]any
		err = json.Unmarshal(b, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal template config map err: %v", err)
		}
		for k, v := range configMap {
			templateMap[k] = v
		}
	}
	// temMap, err := utils.Convert2TemplateData(rsp) //ParseServerTemplateRsp(rsp)
	log.Infof("task erver fetch ok: batch template config; template count: %v", len(templateMap))

	return templateMap, nil
}

// GetStdData 获取标准数据
func (r *ReadImpl) GetStdData(configVersion map[string]*model.ConfigVersion, deviceNums []string) (*model.StdData, error) {
	deviceNums = utils.GetTargetDevice()
	stdMap := make(map[string]any, 0)
	log.Infof("task count:%d, devices: %v", len(deviceNums), deviceNums)
	std := new(model.StdData)
	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		req := &cmdbPb.ReqGetCollectorPoint{
			DeviceNumbers: deviceNums[begin:end],
		}

		rsp, err := r.cmdbProxy.GetCollectorPoint(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("task server std point config fetch fail: %v", err)
			return nil, err
		}
		b, err := json.Marshal(rsp.ConfigMap)
		if err != nil {
			return nil, fmt.Errorf("marshal std point config map err: %v", err)
		}
		var configMap map[string]any
		err = json.Unmarshal(b, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal std point config map err: %v", err)
		}
		for k, v := range configMap {
			stdMap[k] = v
		}
		stdPoints, err := utils.ParseStdPointConfigMap(configMap)
		// stdJson, err := ParseServerStdRsp(rsp)
		if err != nil {
			log.Errorf("task server std point config: parse fail: %v", err)
			return nil, err
		}
		std.StdPointsInfo = append(std.StdPointsInfo, *stdPoints...)

	}
	log.Infof("task erver fetch ok: std point config; stdPoint count: %v", len(std.StdPointsInfo)) // 写本地文件
	err := utils.SaveConfigMapToDirFileWithVersion(stdMap, consts.StdTag, configVersion)
	if err != nil {
		log.Warnf("save std config map fail: %v", err)
	}
	return std, nil
}

// WatchCallback 监听回调
func (r *ReadImpl) WatchCallback() {
	r.chConfigChanged <- true
}
