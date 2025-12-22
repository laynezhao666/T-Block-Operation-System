// Package localfile 从本地文件获取采集配置的实现
package localfile

import (
	"collector/entity/collectors"
	"collector/utils"
	"fmt"
	"os"

	"etrpc-go/log"

	jsoniter "github.com/json-iterator/go"
)

// FetcherImpl 从本地文件获取配置的获取器实现
type FetcherImpl struct{}

var (
	f *FetcherImpl
)

// Fetcher 获取配置获取器
func Fetcher() *FetcherImpl {
	return f
}

// Name 采集器名称
func (f *FetcherImpl) Name() string {
	return "local file config fetcher"
}

// FetchCollectDevices 从本地配置文件夹获取设备配置
func (f *FetcherImpl) FetchCollectDevices(deviceNumbers []string) ([]byte, error) {
	configMap := make(map[string]any)
	for _, d := range deviceNumbers {
		filePath := collectors.CollectDevicesConfigDir + d + collectors.JsonFileSuffix
		if !utils.IsExist(filePath) {
			return nil, fmt.Errorf("file <%v> not exist", d)
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file <%v> fail: %v", d, err)
		}
		info := make(map[string]any)
		err = jsoniter.Unmarshal(b, &info)
		if err != nil {
			log.Errorf("unmarshal fail: %v, file content [%v]", err, b)
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", d, err)
		}
		configMap[d] = info
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal collector devices config fail: %v", err)
	}
	return value, nil
}

// FetchCollectTemplates 从本地配置文件夹获取模板配置
func (f *FetcherImpl) FetchCollectTemplates(templateNames []string) ([]byte, error) {
	configMap := make(map[string]any)
	for _, t := range templateNames {
		filePath := collectors.CollectTemplatesConfigDir + t + collectors.JsonFileSuffix
		if !utils.IsExist(filePath) {
			return nil, fmt.Errorf("file <%v> not exist", t)
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			log.Errorf("unmarshal fail: %v, file content [%v]", err, b)
			return nil, fmt.Errorf("read file <%v> fail: %v", t, err)
		}
		info := make(map[string]any)
		err = jsoniter.Unmarshal(b, &info)
		if err != nil {
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", t, err)
		}
		configMap[t] = info
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal collector templates config fail: %v", err)
	}
	return value, nil
}

// FetchStdPoints 从本地配置文件夹获取测点配置
func (f *FetcherImpl) FetchStdPoints(deviceNumbers []string) ([]byte, error) {
	configMap := make(map[string]any)
	for _, d := range deviceNumbers {
		filePath := collectors.StdPointsConfigDir + d + collectors.JsonFileSuffix
		if !utils.IsExist(filePath) {
			return nil, fmt.Errorf("file <%v> not exist", d)
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			log.Errorf("unmarshal fail: %v, file content [%v]", err, b)
			return nil, fmt.Errorf("read file <%v> fail: %v", d, err)
		}
		info := make(map[string]any)
		err = jsoniter.Unmarshal(b, &info)
		if err != nil {
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", d, err)
		}
		configMap[d] = info
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal collector points config fail: %v", err)
	}
	return value, nil
}

// FetchConfigModifyTime 从本地配置文件夹获取测点获取配置修改时间
func (f *FetcherImpl) FetchConfigModifyTime(deviceNumbers []string) ([]byte, error) {
	configMap := make(map[string]any)
	for _, d := range deviceNumbers {
		filePath := collectors.ConfigModifyTimeDir + d + collectors.JsonFileSuffix
		if !utils.IsExist(filePath) {
			return nil, fmt.Errorf("file <%v> not exist", d)
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			log.Errorf("unmarshal fail: %v, file content [%v]", err, b)
			return nil, fmt.Errorf("read file <%v> fail: %v", d, err)
		}
		info := make(map[string]any)
		err = jsoniter.Unmarshal(b, &info)
		if err != nil {
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", d, err)
		}
		configMap[d] = info
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal config modify time configMap fail: %v", err)
	}
	return value, nil
}

// FetchStdDevices  从本地配置文件夹获取标准设备配置
func (f *FetcherImpl) FetchStdDevices(collectDeviceNumbers []string) ([]byte, error) {
	configMap := make(map[string]any)
	for _, d := range collectDeviceNumbers {
		filePath := collectors.StdDevicesConfigDir + d + collectors.JsonFileSuffix
		if !utils.IsExist(filePath) {
			return nil, fmt.Errorf("file <%v> not exist", d)
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			log.Errorf("unmarshal fail: %v, file content [%v]", err, b)
			return nil, fmt.Errorf("read file <%v> fail: %v", d, err)
		}
		info := make(map[string]any)
		err = jsoniter.Unmarshal(b, &info)
		if err != nil {
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", d, err)
		}
		configMap[d] = info
	}
	value, err := jsoniter.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal std devices config fail: %v", err)
	}
	return value, nil
}
