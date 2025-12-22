// Package cmdb 从CMDB获取采集配置的实现
package cmdb

import (
	"collector/entity/collectors"
	"fmt"
	"sync"

	cmdbPb "trpcprotocol/cmdb"

	jsoniter "github.com/json-iterator/go"
	"trpc.group/trpc-go/trpc-go"
)

// FetcherImpl 从CMDB获取配置的获取器实现
type FetcherImpl struct {
	cmdbProxy cmdbPb.ConfigQueryClientProxy
}

const (
	stdDeviceListName string = "list"
)

var (
	f                     *FetcherImpl
	once                  sync.Once
	isConfigFileFormatted = true
)

// Fetcher 获取配置获取器
func Fetcher() *FetcherImpl {
	once.Do(func() {
		f = &FetcherImpl{
			cmdbProxy: cmdbPb.NewConfigQueryClientProxy(),
		}
	})
	return f
}

// Name 采集器名称
func (f *FetcherImpl) Name() string {
	return "cmdb config fetcher"
}

// FetchCollectDevices 将请求转发到CMDB获取设备配置，并将结果存储到本地文件
func (f *FetcherImpl) FetchCollectDevices(deviceNumbers []string) ([]byte, error) {
	req := &cmdbPb.ReqGetCollectorDevice{
		DeviceNumbers: deviceNumbers,
	}
	rsp, err := f.cmdbProxy.GetCollectorDevice(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, fmt.Errorf("devices config fetch fail: %v", err)
	}
	configMap := rsp.GetConfigMap()
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal collector devices config configMap fail: %v", err)
	}
	go updateConfigToFile(collectors.CollectDevicesConfigDir, configMap)
	return value, nil
}

// FetchCollectTemplates 将请求转发到CMDB获取模板配置，并将结果存储到本地文件
func (f *FetcherImpl) FetchCollectTemplates(templateNames []string) ([]byte, error) {
	req := &cmdbPb.ReqGetCollectorTemplate{
		TemplateNames: templateNames,
	}
	rsp, err := f.cmdbProxy.GetCollectorTemplate(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, fmt.Errorf("templates config fetch fail: %v", err)
	}
	configMap := rsp.GetConfigMap()
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal collector templates config configMap fail: %v", err)
	}
	go updateConfigToFile(collectors.CollectTemplatesConfigDir, configMap)
	return value, nil
}

// FetchStdPoints 将请求转发到CMDB获取测点配置，并将结果存储到本地文件
func (f *FetcherImpl) FetchStdPoints(deviceNumbers []string) ([]byte, error) {
	req := &cmdbPb.ReqGetCollectorPoint{
		DeviceNumbers: deviceNumbers,
	}
	rsp, err := f.cmdbProxy.GetCollectorPoint(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, fmt.Errorf("points config fetch fail: %v", err)
	}
	configMap := rsp.GetConfigMap()

	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal points config configMap fail: %v", err)
	}
	go updateConfigToFile(collectors.StdPointsConfigDir, configMap)
	return value, nil
}

// FetchConfigModifyTime 将请求转发到CMDB获取配置修改时间，并将结果存储到本地文件
func (f *FetcherImpl) FetchConfigModifyTime(deviceNumbers []string) ([]byte, error) {
	req := &cmdbPb.ReqGetCollectorDevice{
		DeviceNumbers: deviceNumbers,
	}
	rsp, err := f.cmdbProxy.GetConfigModifyTime(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, fmt.Errorf("config modify time fetch fail: %v", err)
	}
	configMap := rsp.GetConfigMap()

	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal config modify time configMap fail: %v", err)
	}
	go updateConfigToFile(collectors.ConfigModifyTimeDir, configMap)
	return value, nil
}

// FetchStdDevices 将请求转发到CMDB获取标准设备，并将结果存储到本地文件
func (f *FetcherImpl) FetchStdDevices(collectDeviceNumbers []string) ([]byte, error) {
	configMap := map[string]any{}
	for _, d := range collectDeviceNumbers {
		req := &cmdbPb.ReqGetDeviceEntity{
			BelongCollector: d,
		}
		rsp, err := f.cmdbProxy.GetDeviceEntity(trpc.BackgroundContext(), req)
		if err != nil {
			return nil, fmt.Errorf("std device config fetch fail: %v", err)
		}
		list := rsp.GetList()
		configMap[d] = map[string]any{
			stdDeviceListName: list,
		}
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal std device config configMap fail: %v", err)
	}
	go updateConfigToFile(collectors.StdDevicesConfigDir, configMap)
	return value, nil
}
