// Package utils 工具包
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultFilePerm = 0666
	DefaultDirPerm  = 0755
)

// WriteBytesToFile 写字节到文件，文件不存在则创建文件
func WriteBytesToFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DefaultFilePerm)
	if err != nil {
		return fmt.Errorf("fail to open file: %v, err: %v", filePath, err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("fail to write file: %v, err: %v", filePath, err)
	}
	return nil
}

// IsExist 判断文件/文件夹是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// CreateDir 创建文件夹，如过文件夹已存在则不创建
func CreateDir(path string) error {
	if IsExist(path) {
		return nil
	}
	err := MakeDirAll(path, DefaultDirPerm)
	if err != nil {
		return err
	}
	return nil
}

// MakeDirAll 创建文件夹，以及其依赖的所有父文件夹，
// 由于umask的影响，直接使用os.MkdirAll创建的文件夹权限可能不是我们想要的
// 因此创建后还显式子配置了权限
func MakeDirAll(dirPath string, perm os.FileMode) error {
	// 分割路径并逐层创建目录
	dirs := strings.Split(dirPath, string(os.PathSeparator))
	for i := range dirs {
		dirPath := filepath.Join(dirs[:i+1]...)
		if dirPath == "" {
			continue
		}
		if !IsExist(dirPath) {
			err := os.Mkdir(dirPath, perm)
			if err != nil {
				return fmt.Errorf("make dir %v fail: %v", dirPath, err)
			}
		}
		// 显式设置权限
		if err := os.Chmod(dirPath, perm); err != nil {
			return fmt.Errorf("change dir %v mode fail: %v", dirPath, err)
		}
	}
	return nil
}
