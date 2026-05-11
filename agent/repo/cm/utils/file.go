package utils

import (
	"agent/entity/consts"
	"agent/entity/model"
	"agent/utils/file"
	"agent/utils/file/io"
	"fmt"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	"trpc.group/trpc-go/trpc-go/log"
)

// SaveConfigMapToFile 以configMap形式保存配置为单个文件
func SaveConfigMapToFile(filename string, configMap map[string]any) error {
	var rsp Response
	rsp.Data.ConfigMap = configMap
	err := file.SyncWriteJSON(filename, rsp)
	if err != nil {
		return err
	}
	return nil
}

// SaveConfigMapToDirFileWithVersion 以configMap形式保存为文件夹下的文件
func SaveConfigMapToDirFileWithVersion(configMap map[string]any, configName string, versionMap map[string]*model.ConfigVersion) error {
	// 不同设备的配置收到各自文件夹下
	for k, v := range configMap {
		// 提取版本
		versionModel, ok := versionMap[k]
		if !ok {
			return fmt.Errorf("get version of device %v fail", k)
		}
		var version string
		switch configName {
		case consts.StdTag:
			version = versionModel.Point
		case consts.DeviceTag:
			version = versionModel.Collector
		}
		// 构建文件内容，与local的格式保持一致
		newConfigMap := make(map[string]any)
		newConfigMap[k] = v
		var rsp Response
		rsp.Data.ConfigMap = newConfigMap
		// k为采集设备编号,作为文件夹
		filePath := consts.ProjectPath + "/" + k + "/" + configName + "@" + version + consts.SuffixJSON
		err := file.SyncWriteJSON(filePath, rsp)
		if err != nil {
			return err
		}
	}
	return nil
}

// StdDeviceFile 标准设备文件
type StdDeviceFile struct {
	Data StdDeviceList `json:"data"`
}

// StdDeviceList 配置
type StdDeviceList struct {
	List []model.StdDevice `json:"list"`
}

// SaveConfigListToDirFile 保存标准设备配置
func SaveConfigListToDirFile(dataList map[string][]model.StdDevice, dirs []string) error {
	for _, dir := range dirs {
		var f StdDeviceFile
		list, ok := dataList[dir]
		if !ok {
			continue
		}
		f.Data = StdDeviceList{
			List: list,
		}
		// dir为采集设备编号,作为文件夹
		filePath := consts.ProjectPath + "/" + dir + "/" + consts.StdDeviceTag + consts.SuffixJSON
		err := file.SyncWriteJSON(filePath, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func dataToString(date time.Time) (string, error) {
	// 将 int64 类型的时间戳转换为 time.Time 类型
	//date := time.Unix(timestamp, 0)
	// 将 time.Time 类型转换为 string 类型
	dateString := date.Format("15:04:05")
	return dateString, nil
}

// SaveConfigMapToMultipleFile 将配置写入为单文件夹下的多文件
func SaveConfigMapToMultipleFile[V any](dir string, configMap map[string]V) error {
	successKeys := []string{}
	var (
		b   []byte
		err error
	)
	for k, v := range configMap {
		filePath := dir + k + consts.SuffixJSON
		log.Infof("save templates to multiple file path: %s", filePath)
		b, err = jsoniter.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		err = file.SyncWrite(filePath, b)
		if err != nil {
			return err
		}
		successKeys = append(successKeys, k)
	}
	return nil
}

// DeleteConfigFile 删除配置文件
func DeleteConfigFile(dir, configKey string) error {
	filePath := dir + configKey + consts.SuffixJSON
	exist, err := file.TestExist(filePath)
	if !exist {
		return fmt.Errorf("file not existed")
	}
	if err != nil {
		return err
	}
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

// UpdateConfig 将[filename]指定的配置中，[key]对应的配置项更新/添加成[value]
func UpdateConfig(filename, key string, value any) error {
	var rsp Response
	err := io.JSON.Read(filename, &rsp)
	if err != nil {
		return fmt.Errorf("readconfig failed")
	}
	configMap := rsp.Data.ConfigMap
	configMap[key] = value
	return SaveConfigMapToFile(filename, configMap)
}

// UpdateConfigByMap 将[configMap]的内容更新到[filename]指定的配置中
func UpdateConfigByMap(filename string, configMap map[string]any) error {
	var rsp Response
	err := io.JSON.Read(filename, &rsp)
	if err != nil {
		return fmt.Errorf("readconfig failed")
	}
	oldConfigMap := rsp.Data.ConfigMap
	for k, v := range configMap {
		oldConfigMap[k] = v
	}
	return SaveConfigMapToFile(filename, oldConfigMap)
}
