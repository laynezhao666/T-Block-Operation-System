// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"os"
)

// hostName 当前主机名，初始化时自动获取
var (
	hostName = "unknown"
)

// init 初始化时获取当前主机名
func init() {
	h, err := os.Hostname()
	if err == nil {
		hostName = h
	}
}

// GetHostName 返回当前主机名
func GetHostName() string {
	return hostName
}
