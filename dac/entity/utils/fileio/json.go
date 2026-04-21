// Package fileio 提供文件读写工具，支持JSON和YAML格式。
package fileio

import (
	"encoding/json"
	"os"
)

// jsonIO JSON格式文件读写器
type jsonIO struct {
}

// Read 从JSON文件中读取数据并反序列化到pointer
func (j *jsonIO) Read(name string, pointer interface{}) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, pointer)
}

// Write 将数据序列化为JSON格式并写入文件
func (j *jsonIO) Write(name string, data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return syncWrite(name, b)
}
