// Package file load local config file.
package file

import (
	"etrpc-go/config/util"
	"flag"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go/config"
)

// LocalConfigPath 本地配置地址
var LocalConfigPath = defaultConfigPath

const (
	defaultConfigPath = "./trpc_go.yaml"
)

// Load 加载本地配置
func Load(path string) (map[string]any, error) {
	// 加载本地配置文件
	fileCfg := make(map[string]any)
	fileLoader, err := config.Load(path, config.WithCodec("yaml"),
		config.WithProvider("file"))
	if err != nil {
		return fileCfg, errors.Wrapf(err, "加载本地配置文件错误, 文件路径:%s", path)
	}
	if err = fileLoader.Unmarshal(fileCfg); err != nil {
		return fileCfg, errors.Wrapf(err, "本地配置反序列化Map格式配置错误")
	}
	fileCfgYaml, err := yaml.Marshal(fileCfg)
	if err != nil {
		return fileCfg, errors.Wrapf(err, "本地配置序列化成Yaml格式错误")
	}
	fileCfgYaml = util.ExpandSystemEnv(fileCfgYaml, false)
	newFileCfg := make(map[string]any)
	if err := yaml.Unmarshal(fileCfgYaml, &newFileCfg); err != nil {
		return newFileCfg, errors.Wrapf(err, "替换环境变量后配置解析错误")
	}
	return newFileCfg, nil
}

// GetLocalConfigPath 本地配置文件地址
func GetLocalConfigPath() string {
	if LocalConfigPath == defaultConfigPath && !flag.Parsed() {
		flag.StringVar(&LocalConfigPath, "conf", defaultConfigPath, "local config path")
		flag.Parse()
	}
	return LocalConfigPath
}
