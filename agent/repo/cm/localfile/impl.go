package localfile

import (
	"encoding/json"
	"fmt"
	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/model"
	model3 "agent/logic/collector/device/model"
	"agent/repo/cm/utils"
	"agent/utils/file/io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
)

// ReadImpl 本地文件读取
type ReadImpl struct {
	chConfigChanged chan bool
}

// JSONData 配置
type JSONData struct {
	// Code    int    `json:"code"`
	// Message string `json:"message"`
	Data Data `json:"data"`
}

// Data 配置
type Data struct {
	List []model.StdDevice `json:"list"`
}

// GetStdDevice 获取标准设备
func (r *ReadImpl) GetStdDevice(stdMap map[string]bool) (*model.StdDeviceData, error) {
	var jsonData JSONData
	data, err := os.ReadFile(config.GetRB().GetProjectLocalPath() + "/" + consts.StdDeviceFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("local device tree config: parse fail: %v", err)
	}
	// 筛选出与采集器关联的标准设备
	list := utils.FilterStdDevice(jsonData.Data.List, stdMap)
	// 展示设备编号处理
	list = utils.AddShowDeviceNumber(list)
	// 获取短编号索引
	codeMap, list := utils.GetConciseCodeMap(list)
	stdDevice := &model.StdDeviceData{
		StdDevices:   make([]model.StdDevice, 0),
		StdDeviceMap: make(map[string]model.StdDevice),
		StdPoints:    make(map[string]model3.StdInstancePointsInfo),
	}
	for _, d := range list {
		stdDevice.StdDeviceMap[d.DeviceGid] = d
	}
	stdDevice.StdDevices = list
	stdDevice.ConciseCodeMap = codeMap
	return stdDevice, nil
}

// GetCmdbVersion 获取cmdb版本
func (r *ReadImpl) GetCmdbVersion() (map[string]*model.ConfigVersion, error) {
	return nil, nil
}

// NewReadImpl 本地文件读取
func NewReadImpl(chConfigChanged chan bool) *ReadImpl {
	r := &ReadImpl{
		chConfigChanged: chConfigChanged,
	}
	return r
}

// GetDevices 获取设备
func (r *ReadImpl) GetDevices() ([]model.Device, []model.Device, map[string]any, error) {
	var rsp utils.Response
	deviceMap := make(map[string]any, 0)
	rsp, err := getLatestFileData(consts.DeviceTag)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("local device config: get latest file data fail: %v", err)
	}
	collectDevices, tboxDevices, err := utils.ParseCollectDeviceConfigMap(rsp.Data.ConfigMap)
	if err != nil {
		return nil, nil, deviceMap, fmt.Errorf("local device config: parse fail: %v", err)
	}
	log.Infof("local parse ok: GetDevices count: %v", len(collectDevices))
	return collectDevices, tboxDevices, deviceMap, err
}

// GetTemplate 获取模板
func (r *ReadImpl) GetTemplate(name string) (*model.TemplateData, error) {
	if !strings.HasSuffix(name, consts.SuffixJSON) {
		name += consts.SuffixJSON
	}

	f := filepath.Join(config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeTemplateFile, name)
	t := new(model.TemplateData)
	err := io.JSON.Read(f, t)
	return t, err
}

// GetTemplates 获取模板
func (r *ReadImpl) GetTemplates(list []string) (map[string]any, error) {
	configMap := make(map[string]any)
	folderPath := config.GetRB().GetProjectLocalPath() + "/" + consts.RelativeTemplateDir + "/"
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(folderPath, file.Name())
		// Extracting file key
		fileKey := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file <%v> fail: %v", filePath, err)
		}
		info := make(map[string]any)
		err = json.Unmarshal(b, &info)
		if err != nil {
			return nil, fmt.Errorf("unmarshal file <%v> fail: %v", filePath, err)
		}
		configMap[fileKey] = info
	}
	return configMap, nil
}

// GetStdData 获取标准数据
func (r *ReadImpl) GetStdData(_ map[string]*model.ConfigVersion) (*model.StdData, error) {
	// 读取文件
	rsp, err := getLatestFileData(consts.StdTag)
	if err != nil {
		return nil, err
	}
	stdPoints, err := utils.ParseStdPointConfigMap(rsp.Data.ConfigMap)
	if err != nil {
		log.Errorf("local file std point config: parse fail: %v", err)
		return nil, err
	}

	std := new(model.StdData)
	std.StdPointsInfo = *stdPoints
	log.Infof("local parse ok: stdPoint count: %v", len(std.StdPointsInfo))

	return std, err
}

// 读取最近版本的文件配置
func getLatestFileData(configType string) (utils.Response, error) {
	var rsp utils.Response
	// 获取项目路径
	projectPath := config.GetRB().GetProjectLocalPath()

	// 查找匹配的文件
	files, err := filepath.Glob(filepath.Join(projectPath, fmt.Sprintf("%s*.json", configType)))
	if err != nil {
		log.Errorf("failed to find %s*.json files: %v", configType, err)
		return rsp, fmt.Errorf("failed to find %s*.json files: %v", configType, err)
	}

	if len(files) == 0 {
		log.Errorf("no %s*.json files found", configType)
		return rsp, fmt.Errorf("no %s*.json files found", configType)
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
		return rsp, fmt.Errorf("no %s*.json files found", configType)
	}

	// 读取文件内容
	err = io.JSON.Read(targetFile, &rsp)
	if err != nil {
		log.Errorf("local file %s config: parse fail: %v", configType, err)
		return rsp, fmt.Errorf("local file %s config: parse fail: %v", configType, err)
	}
	return rsp, nil
}

// WatchCallback 监听回调
func (r *ReadImpl) WatchCallback() {
	r.chConfigChanged <- true
}
