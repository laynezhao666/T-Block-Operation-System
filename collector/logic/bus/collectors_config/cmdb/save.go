package cmdb

import (
	"collector/entity/collectors"
	"collector/utils"

	"etrpc-go/log"

	jsoniter "github.com/json-iterator/go"
)

// updateConfigToFile 将配置写入到文件
func updateConfigToFile[V any](dir string, configMap map[string]V) {
	successKeys := []string{}
	var (
		b   []byte
		err error
	)
	for k, v := range configMap {
		filePath := dir + k + collectors.JsonFileSuffix

		if isConfigFileFormatted {
			b, err = jsoniter.MarshalIndent(v, "", "  ")
		} else {
			b, err = jsoniter.Marshal(v)
		}
		if err != nil {
			log.Errorf("fail to marshal config, key %v, err: %v", k, err)
			continue
		}
		err = utils.WriteBytesToFile(filePath, b)
		if err != nil {
			log.Errorf("fail to write config to file, key %v, err: %v", k, err)
			continue
		}
		successKeys = append(successKeys, k)
	}
	log.Infof("update configMap to file success, dir <%v>, keys: %v", dir, successKeys)
}
