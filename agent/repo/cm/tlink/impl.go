// Package tlink tlink模式
package tlink

import (
	"agent/entity/consts"
	"agent/entity/model"
	"agent/repo/cm/utils"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	model2 "agent/logic/collector/device/model"

	collectorPb "trpcprotocol/collector"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/naming/registry"
)

const (
	paramNumberPerConfigFetch = 5
)

// ReadImpl 读取配置
type ReadImpl struct {
	chConfigChanged chan bool
	tlinkProxy      collectorPb.ConfigBusClientProxy
}

// NewReadImpl 读取配置
func NewReadImpl(chConfigChanged chan bool) *ReadImpl {
	r := &ReadImpl{
		chConfigChanged: chConfigChanged,
		tlinkProxy:      collectorPb.NewConfigBusClientProxy(),
	}
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
	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		params, err := json.Marshal(deviceNums[begin:end])
		if err != nil {
			return nil, err
		}
		req := &collectorPb.ReqFetchConfig{
			Params:    params,
			FetchType: collectorPb.ReqFetchConfig_FETCH_STD_DEVICES,
		}
		rsp, err := r.tlinkProxy.FetchConfig(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("tlink std device config fetch fail: %v", err)
			return nil, err
		}
		data := rsp.GetData()
		if data == nil {
			return nil, fmt.Errorf("std devices data nil, devices: %v", deviceNums[begin:end])
		}
		var configMap map[string]any
		err = json.Unmarshal(data, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal std device config map err: %v", err)
		}
		list, stdDevicesMap, err := utils.ParseStdDeviceConfigMap(configMap)
		if err != nil {
			return nil, fmt.Errorf("tlink std device config: parse fail: %v", err)
		}
		// 展示设备编号处理
		list = utils.AddShowDeviceNumber(list)
		// 获取短编号索引
		codeMap, list := utils.GetConciseCodeMap(list)

		// 写本地文件
		err = utils.SaveConfigListToDirFile(stdDevicesMap, deviceNums)
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

// GetDevices 获取采集设备列表
func (r *ReadImpl) GetDevices(deviceNums []string) ([]model.Device, []model.Device, map[string]any, error) {
	totalCollectDevices := make([]model.Device, 0)
	totalTboxDevices := make([]model.Device, 0)
	deviceMap := make(map[string]any, 0)
	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		params, err := json.Marshal(deviceNums[begin:end])
		if err != nil {
			return nil, nil, deviceMap, err
		}
		req := &collectorPb.ReqFetchConfig{
			Params:    params,
			FetchType: collectorPb.ReqFetchConfig_FETCH_COLLECTOR_DEVICES,
		}
		rsp, err := r.tlinkProxy.FetchConfig(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("tlink devices config fetch fail: %v", err)
			return nil, nil, deviceMap, err
		}
		data := rsp.GetData()
		if data == nil {
			return nil, nil, deviceMap, fmt.Errorf("devices data nil, devices: %v", deviceNums[begin:end])
		}
		var configMap map[string]any
		err = json.Unmarshal(data, &configMap)
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("unmarshal device config map err: %v", err)
		}
		for k, v := range configMap {
			deviceMap[k] = v
		}
		collectDevices, tboxDevices, err := utils.ParseCollectDeviceConfigMap(configMap)
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("tlink device config: parse fail: %v", err)
		}
		totalCollectDevices = append(totalCollectDevices, collectDevices...)
		totalTboxDevices = append(totalTboxDevices, tboxDevices...)
	}
	log.Infof("tlink fetch ok: device config; device count: %v", len(totalCollectDevices))
	return totalCollectDevices, totalTboxDevices, deviceMap, nil
}

// GetTemplate 获取采集模板
func (r *ReadImpl) GetTemplate(name string) (*model.TemplateData, error) {
	_, fileName := filepath.Split(name)
	nameList := []string{fileName}
	params, err := json.Marshal(nameList)
	if err != nil {
		return nil, err
	}
	req := &collectorPb.ReqFetchConfig{
		Params:    params,
		FetchType: collectorPb.ReqFetchConfig_FETCH_COLLECTOR_TEMPLATES,
	}
	rsp, err := r.tlinkProxy.FetchConfig(trpc.BackgroundContext(), req)
	if err != nil {
		log.Errorf("tlink template config fetch fail: %v", err)
		return nil, err
	}
	data := rsp.GetData()
	var configMap map[string]any
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal template config map err: %v", err)
	}
	temMap, err := utils.ParseCollectTemplateConfigMap(configMap)
	tem, ok := temMap[fileName]
	if !ok {
		return nil, errors.New("template not exist")
	}

	log.Infof("tlink fetch ok: single template config")

	return tem, err
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
	rawTemplateMap := make(map[string]any)
	for begin := 0; begin < len(nameList); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(nameList) {
			end = len(nameList)
		}
		params, err := json.Marshal(nameList[begin:end])
		if err != nil {
			return nil, err
		}
		req := &collectorPb.ReqFetchConfig{
			Params:    params,
			FetchType: collectorPb.ReqFetchConfig_FETCH_COLLECTOR_TEMPLATES,
		}
		rsp, err := r.tlinkProxy.FetchConfig(trpc.BackgroundContext(), req)
		if err != nil {
			log.Errorf("tlink templates config fetch fail: %v", err)
			return nil, err
		}
		data := rsp.GetData()
		var configMap map[string]any
		err = json.Unmarshal(data, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal templates config map err: %v", err)
		}
		for k, v := range configMap {
			rawTemplateMap[k] = v
		}
	}
	return rawTemplateMap, nil
}

func (r *ReadImpl) GetStdData(configVersion map[string]*model.ConfigVersion, deviceNums []string) (
	*model.StdData, error) {
	if len(deviceNums) == 0 {
		return &model.StdData{}, nil
	}
	log.Infof("task count:%d, devices: %v", len(deviceNums), deviceNums)

	std := new(model.StdData)
	stdMap := make(map[string]any)
	mu := sync.Mutex{}

	type result struct {
		configMap map[string]any
		stdPoints *model2.StdInstancePointsInfo
		err       error
	}

	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	resultsCh := make(chan result, (len(deviceNums)+paramNumberPerConfigFetch-1)/paramNumberPerConfigFetch)

	for begin := 0; begin < len(deviceNums); begin += paramNumberPerConfigFetch {
		end := begin + paramNumberPerConfigFetch
		if end > len(deviceNums) {
			end = len(deviceNums)
		}
		devicesBatch := deviceNums[begin:end]

		wg.Add(1)
		sem <- struct{}{}
		go func(devices []string) {
			defer wg.Done()
			defer func() { <-sem }()

			params, err := json.Marshal(devices)
			if err != nil {
				resultsCh <- result{err: err}
				return
			}
			req := &collectorPb.ReqFetchConfig{
				Params:    params,
				FetchType: collectorPb.ReqFetchConfig_FETCH_STD_POINTS,
			}
			ctx := trpc.BackgroundContext()
			rsp, err := r.tlinkProxy.FetchConfig(ctx, req)
			if err != nil {
				log.ErrorContextf(ctx, "tlink std point config fetch fail: %v", err)
				resultsCh <- result{err: err}
				return
			}
			log.InfoContext(ctx, "fetch success")
			data := rsp.GetData()
			if data == nil {
				resultsCh <- result{err: fmt.Errorf("std data nil, req devices: %v", devices)}
				return
			}
			var configMap map[string]any
			err = json.Unmarshal(data, &configMap)
			if err != nil {
				resultsCh <- result{err: fmt.Errorf("unmarshal std config map err: %v", err)}
				return
			}
			stdPoints, err := utils.ParseStdPointConfigMap(configMap)
			if err != nil {
				resultsCh <- result{err: fmt.Errorf("tlink std point config: parse fail: %v", err)}
				return
			}
			resultsCh <- result{configMap: configMap, stdPoints: stdPoints}
		}(devicesBatch)
	}

	// 等待所有请求完成后关闭结果通道
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for res := range resultsCh {
		if res.err != nil {
			return nil, res.err
		}
		mu.Lock()
		for k, v := range res.configMap {
			stdMap[k] = v
		}
		std.StdPointsInfo = append(std.StdPointsInfo, *res.stdPoints...)
		mu.Unlock()
	}

	log.Infof("tlink fetch ok: std point config; stdPoint count: %v", len(std.StdPointsInfo))
	// 写本地文件
	err := utils.SaveConfigMapToDirFileWithVersion(stdMap, consts.StdTag, configVersion)
	if err != nil {
		log.Warnf("save std config fail: %v", err)
	}
	return std, nil
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
		params, err := json.Marshal(deviceNums[begin:end])
		if err != nil {
			return nil, err
		}
		node := &registry.Node{}
		req := &collectorPb.ReqFetchConfig{
			Params:    params,
			FetchType: collectorPb.ReqFetchConfig_FETCH_CONFIG_MODIFY_TIME,
		}
		rsp, err := r.tlinkProxy.FetchConfig(trpc.BackgroundContext(), req, client.WithSelectorNode(node))
		log.Debugf("tlink fetch node:----------- %v", node)
		if err != nil {
			// log.Errorf("tlink cm version fetch fail: %v", err)
			return nil, err
		}
		data := rsp.GetData()
		var configMap map[string]any
		err = json.Unmarshal(data, &configMap)
		if err != nil {
			return nil, fmt.Errorf("unmarshal cmdb version config map err: %v", err)
		}
		tempMap, err := utils.ParseCmdbVersionConfigMap(configMap)
		if err != nil {
			return nil, fmt.Errorf("tlink cmdb versin config: parse fail: %v", err)
		}
		for k, v := range tempMap {
			versionMap[k] = v
		}
	}
	return versionMap, nil
}

// WatchCallback 监听回调
func (r *ReadImpl) WatchCallback() {
	r.chConfigChanged <- true
}

// SnReadImpl 读取配置
type SnReadImpl struct {
	tlinkProxy collectorPb.ConfigBusClientProxy
}

// NewSnReadImpl 读取配置
func NewSnReadImpl() *SnReadImpl {
	r := &SnReadImpl{
		tlinkProxy: collectorPb.NewConfigBusClientProxy(),
	}
	return r
}
