package io

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent/utils/file"
)

type jsonIO struct {
}

// Read 读取文件
func (j *jsonIO) Read(name string, pointer interface{}) error {
	// 清理并验证输入路径
	cleanPath := filepath.Clean(name)

	// 检查路径是否包含路径遍历（如 "../"）
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("非法的文件写入路径")
	}

	b, err := os.ReadFile(cleanPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, pointer)
}

// Write 写入文件
func (j *jsonIO) Write(name string, data interface{}) error {
	return file.SyncWriteJSON(name, data)
}
