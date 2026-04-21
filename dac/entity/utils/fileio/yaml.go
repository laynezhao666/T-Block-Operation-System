// Package fileio 提供文件读写工具，支持JSON和YAML格式。
package fileio

import (
	"os"

	"gopkg.in/yaml.v3"
)

// yamlIO YAML格式文件读写器
type yamlIO struct {
}

// Read 从YAML文件中读取数据并反序列化到pointer
func (y *yamlIO) Read(name string, pointer interface{}) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, pointer)
}

// Write 将数据序列化为YAML格式并写入文件
func (y *yamlIO) Write(name string, data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return syncWrite(name, b)
}
