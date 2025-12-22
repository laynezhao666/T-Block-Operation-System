package file

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SyncWriteJSON 将数据写入 json 文件
func SyncWriteJSON(filename string, data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return SyncWrite(filename, b)
}

// SyncWrite 将数据写入文件
func SyncWrite(filename string, data []byte) error {
	var err error
	currentDir := filepath.Dir(filename)
	if err = os.MkdirAll(currentDir, os.ModePerm); err != nil {
		return err
	}
	// 在当前目录下生成临时文件
	// 避免跨文件系统导致的重命名失败
	tempFile, err := os.CreateTemp(currentDir, "misc*")
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	if _, err = tempFile.Write(data); err != nil {
		return err
	}
	if err = tempFile.Sync(); err != nil {
		return err
	}
	if err = tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempFile.Name(), filename)
}
