package backup

import (
	"agent/entity/consts"
	"agent/entity/model"
	model3 "agent/logic/collector/device/model"
	"agent/repo/cm/utils"
	utils2 "agent/utils"
	"agent/utils/file/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
func (r *ReadImpl) GetStdDevice(stdMap map[string]bool, deviceNums []string) (*model.StdDeviceData, error) {
	var jsonData JSONData
	stdDevice := &model.StdDeviceData{
		StdDevices:      make([]model.StdDevice, 0),
		StdDeviceMap:    make(map[string]model.StdDevice),
		StdPoints:       make(map[string]model3.StdInstancePointsInfo),
		ConciseCodeMap:  make(map[string]string),
		DeviceNumberMap: make(map[string]string),
	}
	//devicesNumbers := utils.GetTargetDevice()
	for _, d := range deviceNums {
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
			if d.DeviceNumber != "" {
				stdDevice.DeviceNumberMap[d.DeviceNumber] = d.DeviceGid
			}
		}
		stdDevice.StdDevices = append(stdDevice.StdDevices, list...)
		for k, v := range codeMap {
			stdDevice.ConciseCodeMap[k] = v
		}
	}
	return stdDevice, nil
}

// GetCmdbVersion 从本地文件获取获取版本
func (r *ReadImpl) GetCmdbVersion() (map[string]*model.ConfigVersion, error) {
	devicesNumbers := utils.GetTargetDevice()
	r.versionMap = make(map[string]*model.ConfigVersion, len(devicesNumbers))
	for _, d := range devicesNumbers {
		dir := consts.ProjectPath + "/" + d
		collectorVersion, err := getLatestFileVersionFromDirectory(dir, consts.DeviceTag)
		if err != nil {
			log.Debugf("get cmdb collector version fail: %v", err)
		}
		pointVersion, err := getLatestFileVersionFromDirectory(dir, consts.StdTag)
		if err != nil {
			log.Debugf("get cmdb collector version fail: %v", err)
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
func (r *ReadImpl) GetDevices(devices []string) ([]model.Device, []model.Device, map[string]any, error) {
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
		// 只读取目标列表中的设备
		if !utils2.InSlice(d, devices) {
			continue
		}
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
// list 参数指定需要获取的模板名称列表，仅读取匹配的模板文件，避免全量加载导致内存膨胀
func (r *ReadImpl) GetTemplates(list []string) (map[string]any, error) {
	configMap := make(map[string]any)

	// 构建需要获取的模板名称白名单，用于过滤无关文件
	needSet := make(map[string]bool, len(list))
	for _, name := range list {
		// 去掉路径前缀，只保留文件名部分（与其他 Reader 实现保持一致）
		_, fileName := filepath.Split(name)
		needSet[fileName] = true
	}

	deviceNumbers := utils.GetTargetDevice()
	for _, d := range deviceNumbers {
		folderPath := consts.ProjectPath + "/" + d + "/" + consts.RelativeTemplateDir + "/"
		files, err := os.ReadDir(folderPath)
		if err != nil {
			log.Warnf("backup failed to read directory: %v", err)
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			// 提取文件名（不含扩展名）作为模板标识
			fileKey := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			// 仅读取 list 中指定的模板（白名单过滤）
			if len(needSet) > 0 && !needSet[fileKey] {
				continue
			}

			// 同名模板内容相同，已读取过则跳过，避免重复读取
			if _, exists := configMap[fileKey]; exists {
				continue
			}

			filePath := filepath.Join(folderPath, file.Name())
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
	log.Infof("backup parse ok: template count: %v (requested: %v)", len(configMap), len(list))
	return configMap, nil
}

// GetStdData 获取标准数据
func (r *ReadImpl) GetStdData(_ map[string]*model.ConfigVersion, deviceNums []string) (*model.StdData, error) {
	std := new(model.StdData)
	if len(deviceNums) == 0 {
		return &model.StdData{}, nil
	}
	if r.versionMap == nil {
		// 从本地获取版本映射
		_, err := r.GetCmdbVersion()
		if err != nil {
			log.Warnf("get cmdb version fail: %v")
			return nil, err
		}
	}
	//for _, d := range deviceNums {
	//	if version, exists := r.versionMap[d]; exists && version.Point != "" {
	//		targetFile := filepath.Join(consts.ProjectPath, d, consts.StdTag+"@"+version.Point+consts.SuffixJSON)
	//
	//		rsp := utils.Response{}
	//		if err := io.JSON.Read(targetFile, &rsp); err != nil {
	//			log.Errorf("backup file %s read fail: %v", targetFile, err)
	//			continue
	//		}
	//
	//		if stdPoints, err := utils.ParseStdPointConfigMap(rsp.Data.ConfigMap); err == nil {
	//			std.StdPointsInfo = append(std.StdPointsInfo, *stdPoints...)
	//		} else {
	//			log.Errorf("local std point config parse fail: %v", err)
	//		}
	//	} else {
	//		log.Warnf("version not found for device %s", d)
	//	}
	//}
	results := make(chan model3.StdInstancePointsInfo, 10)
	var wg sync.WaitGroup

	for _, d := range deviceNums {
		wg.Add(1)
		go func(device string) {
			defer wg.Done()
			if points, ok := r.readDeviceStd(device); ok {
				results <- points
			}
		}(d)
	}
	go func() { wg.Wait(); close(results) }()

	for res := range results {
		std.StdPointsInfo = append(std.StdPointsInfo, res...)
	}
	//log.Infof("Loaded from %d/%d devices", len(std.StdPointsInfo), len(devices))
	log.Infof("backup parse ok: stdPoint count: %v", len(std.StdPointsInfo))
	return std, nil
}

func (r *ReadImpl) readDeviceStd(device string) (model3.StdInstancePointsInfo, bool) {
	ver, exist := r.versionMap[device]
	if !exist || ver.Point == "" {
		return nil, false
	}

	file := filepath.Join(consts.ProjectPath, device, consts.StdTag+"@"+ver.Point+consts.SuffixJSON)
	var rsp utils.Response
	if err := io.JSON.Read(file, &rsp); err != nil {
		return nil, false
	}

	points, err := utils.ParseStdPointConfigMap(rsp.Data.ConfigMap)
	if err != nil {
		return nil, false
	}

	return *points, true
}

func getLatestFileVersionFromDirectory(dir, configType string) (string, error) {
	// 获取项目路径
	projectPath := dir

	// 查找匹配的文件
	files, err := filepath.Glob(filepath.Join(projectPath, fmt.Sprintf("%s*.json", configType)))
	if err != nil {
		log.Debugf("failed to find %s*.json files: %v", configType, err)
		return "", fmt.Errorf("failed to find %s*.json files: %v", configType, err)
	}

	if len(files) == 0 {
		log.Debugf("no %s*.json files found", configType)
		return "", fmt.Errorf("no %s*.json files found", configType)
	}

	// 修改正则表达式以适配新格式：时间戳-序列号
	re := regexp.MustCompile(regexp.QuoteMeta(configType) + `@(\d+-\d+)\.json`)
	var maxVer model.FileVersion
	found := false

	for _, file := range files {
		base := filepath.Base(file) // 获取文件名（不含路径）
		matches := re.FindStringSubmatch(base)
		if len(matches) < 2 {
			continue // 跳过不匹配的文件
		}

		fullVersion := matches[1]
		// 拆分时间戳和序列号
		parts := strings.Split(fullVersion, "-")
		if len(parts) != 2 {
			continue
		}

		timestamp, err1 := strconv.ParseInt(parts[0], 10, 64)
		sequence, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			continue // 跳过无效数字
		}

		// 第一次找到有效版本时初始化 maxVer
		if !found {
			maxVer = model.FileVersion{Timestamp: timestamp, Sequence: sequence, FullVersion: fullVersion}
			found = true
			continue
		}

		// 比较版本：时间戳优先，相同则取序列号更小的
		if timestamp > maxVer.Timestamp ||
			(timestamp == maxVer.Timestamp && sequence < maxVer.Sequence) {
			maxVer = model.FileVersion{Timestamp: timestamp, Sequence: sequence, FullVersion: fullVersion}
		}
	}

	if !found {
		return "", fmt.Errorf("no valid version found")
	}
	return maxVer.FullVersion, nil
}

// WatchCallback 监听回调
func (r *ReadImpl) WatchCallback() {
	r.chConfigChanged <- true
}
