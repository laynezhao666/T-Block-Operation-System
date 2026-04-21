// Package fileio 提供文件读写工具，支持JSON和YAML格式。
package fileio

import (
	"fmt"
	"os"
	"path/filepath"
)

// syncWrite 原子写入文件。先写入临时文件再重命名，确保写入的原子性。
func syncWrite(filename string, data []byte) error {
	var err error
	currentDir := filepath.Dir(filename)
	if err = os.MkdirAll(currentDir, os.ModePerm); err != nil {
		return err
	}
	tempFile, err := os.CreateTemp(currentDir, "dac_fileio*")
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

	// 保持原文件权限
	if srcInfo, err := os.Stat(filename); err == nil {
		if err = os.Chmod(tempFile.Name(), srcInfo.Mode()); err != nil {
			return fmt.Errorf(
				"chmod %v error: %w", tempFile.Name(), err)
		}
	}

	return os.Rename(tempFile.Name(), filename)
}
