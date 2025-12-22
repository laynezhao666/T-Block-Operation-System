package model

import (
	"encoding/json"
	"fmt"
)

// DriverInfo 驱动信息
type DriverInfo struct {
	// 设备类型
	Class string `json:"cls"`
	// 厂商
	Vendor string `json:"vendor"`
	// 驱动协议库（协议名称）
	DriverName string `json:"drvlib"`
	// 协议版本
	ProtocolVersion string `json:"protver"`
	// 扩展参数
	Extend string `json:"extend"`
}

// ExpressionDefinition 采集点表达式定义
type ExpressionDefinition struct {
	// 映射表达式
	Expr string `json:"expression"`
	// 映射
	Mapping string `json:"var_mapping"`
}

// ProtocolDefinition 协议定义
type ProtocolDefinition struct {
	// 字节序
	Byteorder string `json:"byteorder"`
	// 采集指令
	Command string `json:"cmd"`
	// 测点数据类型
	Datatype string `json:"datatype"`
	// 扩展参数
	Extend string `json:"ext"`
	// 寄存器地址
	Register string `json:"val_key"`
	// 偏移
	Offset string `json:"offset"`
	// 缩放因子
	Scale string `json:"scale"`
	// 命令发送间隔权重
	CmdIntervalWeight string `json:"cmd_interval"`
}

// Parse 解析
func (p *ProtocolDefinition) Parse(data string) error {
	if p == nil {
		return fmt.Errorf("ProtocolDefinition is nil")
	}
	return json.Unmarshal([]byte(data), p)
}
