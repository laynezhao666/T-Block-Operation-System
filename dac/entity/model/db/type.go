// Package db 提供门禁系统数据库模型定义。
package db

// IDType 数据库主键ID类型
type IDType = int

// Protocol 门禁控制器通讯协议信息
type Protocol struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Profile 门禁控制器设备信息
type Profile struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	SN     string `json:"sn"`
}

// GIDType 全局唯一标识符类型
type GIDType string
