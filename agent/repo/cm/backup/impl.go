package backup

import (
	"encoding/json"
	"fmt"
	"agent/entity/consts"
	"agent/entity/model"
	model3 "agent/logic/collector/device/model"
	"agent/repo/cm/utils"
	"agent/utils/file/io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
)

// ReadImpl 读取备份数据
type ReadImpl struct {
	chConfigChanged chan bool
	versionMap      map[string]*model.ConfigVersion
}

// JSONData json数据
type JSONData struct {
	Data Data `json:"data"`
}

// Data 配置
type Data struct {
	List []model.StdDevice `json:"list"`
}

// GetStdDevice 获取标准设备
func (r *ReadImpl) GetStdDevice(stdMap map[string]bool) (*model.StdDeviceData, error) {
	var jsonData JSONData
	stdDevice := &model.StdDeviceData{
		StdDevices:     make([]model.StdDevice, 0),
		StdDeviceMap:   make(map[string]model.StdDevice),
		StdPoints:      make(map[string]model3.StdInstancePointsInfo),
		ConciseCodeMap: make(map[string]string),
	}
	devicesNumbers := utils.GetTargetDevice()
	for _, d := range devicesNumbers {
		data, err := os.ReadFile(consts.ProjectPath + "/" + d + "/std_device" + consts.SuffixJSON)
		if err != nil {
			log.Warnf("backup std device config: read fail: %v", err)
			continue

			// return nil, err
		}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("backup std device config: parse fail: %v", err)
		}
		// 筛选出与采集器关联的标准设备
		list := utils.FilterStdDevice(jsonData.Data.List, stdMap)
		// 展示设备编号处理
		list = utils.AddShowDeviceNumber(list)
		// 获取短编号索引
		codeMap, list := utils.GetConciseCodeMap(list)

		for _, d := range list {
			stdDevice.StdDeviceMap[d.DeviceGid] = d
		}
		stdDevice.StdDevices = append(stdDevice.StdDevices, list...)
		for k, v := range codeMap {
			stdDevice.ConciseCodeMap[k] = v
		}
	}
	return stdDevice, nil
}

// GetCmdbVersion 获取cmdb版本
func (r *ReadImpl) GetCmdbVersion() (map[string]*model.ConfigVersion, error) {
	devicesNumbers := utils.GetTargetDevice()
	r.versionMap = make(map[string]*model.ConfigVersion, len(devicesNumbers))
	for _, d := range devicesNumbers {
		dir := consts.ProjectPath + "/" + d
		collectorVersion, err := getLatestFileVersionFromDirectory(dir, consts.DeviceTag)
		if err != nil {
			log.Warnf("get cmdb collector version fail: %v", err)
			continue
		}
		pointVersion, err := getLatestFileVersionFromDirectory(dir, consts.StdTag)
		if err != nil {
			log.Warnf("get cmdb collector version fail: %v", err)
			continue
		}
		r.versionMap[d] = &model.ConfigVersion{
			Collector: collectorVersion,
			Point:     pointVersion,
		}
	}
	return r.versionMap, nil
}

// NewReadImpl 创建读取实例
func NewReadImpl(chConfigChanged chan bool) *ReadImpl {
	r := &ReadImpl{
		chConfigChanged: chConfigChanged,
	}
	return r
}

// GetDevices 获取设备列表
func (r *ReadImpl) GetDevices() ([]model.Device, []model.Device, map[string]any, error) {
	deviceMap := make(map[string]any, 0)
	totalCollectDevices := []model.Device{}
	totalTboxDevices := []model.Device{}
	if r.versionMap == nil {
		_, err := r.GetCmdbVersion()
		if err != nil {
			log.Warnf("get cmdb version fail: %v", err)
			return nil, nil, nil, err
		}
	}
	for d, v := range r.versionMap {
		version := v.Collector
		targetFile := filepath.Join(consts.ProjectPath, d, consts.DeviceTag+"@"+version+consts.SuffixJSON)
		// 读取文件内容
		rsp := utils.Response{}
		err := io.JSON.Read(targetFile, &rsp)
		if err != nil {
			log.Errorf("backup file %s config: read fail: %v", targetFile, err)
			continue
		}
		collectDevices, tboxDevices, err := utils.ParseCollectDeviceConfigMap(rsp.Data.ConfigMap)
		if err != nil {
			return nil, nil, deviceMap, fmt.Errorf("backup device config: parse fail: %v", err)
		}
		totalCollectDevices = append(totalCollectDevices, collectDevices...)
		totalTboxDevices = append(totalTboxDevices, tboxDevices...)
	}

	log.Warnf("backup parse ok: GetDevices count: %v", len(totalCollectDevices))

	return totalCollectDevices, totalTboxDevices, deviceMap, nil
}

// GetTemplate 获取模板
func (r *ReadImpl) GetTemplate(name string) (*model.TemplateData, error) {
	// if !strings.HasSuffix(name, consts.SuffixJSON) {
	// 	name += consts.SuffixJSON
	// }

	// f := filepath.Join(config.GetRB().GetProjectLocalPath()+"/"+consts.RelativeTemplateFile, name)
	// t := new(model.TemplateData)
	// err := io.JSON.Read(f, t)
	// return t, err
	return nil, fmt.Errorf("backup get template not implemented")
}

// GetTemplates 获取模板
func (r *ReadImpl) GetTemplates(list []string) (map[string]any, error) {
	configMap := make(map[string]any)
	deviceNumbers := utils.GetTargetDevice()
	temMap := map[string]*model.TemplateData{}
	for _, d := range deviceNumbers {
		folderPath := consts.ProjectPath + "/" + d + "/" + consts.RelativeTemplateDir + "/"
		files, err := os.ReadDir(folderPath)
		if err != nil {
			// return nil, fmt.Errorf("failed to read directory: %v", err)
			log.Warnf("backup failed to read directory: %v", err)
			continue
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
	}
	log.Infof("backup parse ok: template count: %v", len(temMap))
	return configMap, nil
}

// GetStdData 获取标准数据
func (r *ReadImpl) GetStdData(_ map[string]*model.ConfigVersion) (*model.StdData, error) {
	std := new(model.StdData)
	if r.versionMap == nil {
		_, err := r.GetCmdbVersion()
		if err != nil {
			log.Warnf("get cmdb version fail: %v")
			return nil, err
		}
	}
	for d, v := range r.versionMap {
		version := v.Point
		targetFile := filepath.Join(consts.ProjectPath, d, consts.StdTag+"@"+version+consts.SuffixJSON)
		// 读取文件内容
		rsp := utils.Response{}
		err := io.JSON.Read(targetFile, &rsp)
		if err != nil {
			log.Errorf("backup file %s config: read fail: %v", targetFile, err)
			continue
		}
		stdPoints, err := utils.ParseStdPointConfigMap(rsp.Data.ConfigMap)
		if err != nil {
			log.Errorf("local file std point config: parse fail: %v", err)
			return nil, err
		}
		std.StdPointsInfo = append(std.StdPointsInfo, *stdPoints...)
	}

	log.Infof("backup parse ok: stdPoint count: %v", len(std.StdPointsInfo))
	return std, nil
}

func getLatestFileVersionFromDirectory(dir, configType string) (string, error) {
	// 获取项目路径
	projectPath := dir

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
	var maxTimestamp string = ""

	// 遍历所有匹配的文件，找到时间戳最大的文件
	for _, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 2 { // 匹配成功
			timestamp := matches[1]
			if err != nil {
				log.Warnf("invalid timestamp in filename %s: %v", file, err)
				continue
			}

			// 更新最大时间戳的文件
			if timestamp > maxTimestamp {
				maxTimestamp = timestamp
			}
		}
	}

	return maxTimestamp, nil
}

// WatchCallback 监听回调
func (r *ReadImpl) WatchCallback() {
	r.chConfigChanged <- true
}
