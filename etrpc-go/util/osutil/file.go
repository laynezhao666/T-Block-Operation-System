// Package osutil provides various os tools
package osutil

import (
	"os"
)

// IsDir 判断是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile 判断是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// IsExistDir 判断目录是否存在
func IsExistDir(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirIfNotExist 目录不存在则创建
func CreateDirIfNotExist(path string) error {
	if IsExistDir(path) {
		return nil
	}
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}
