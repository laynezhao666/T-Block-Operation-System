package model

import (
	"agent/entity/consts"
	"agent/utils"
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
	// 精度
	Precision string `json:"precision"`
}

// MatchDirectCalc 是否匹配直接计算的规则,返回是/否以及匹配的测点名和对应的参数名
func (e *ExpressionDefinition) MatchDirectCalc(pointNoSet map[string]struct{}) (bool, string, string) {
	if len(e.Expr) == 0 || len(e.Mapping) == 0 || len(pointNoSet) == 0 {
		return false, "", ""
	}
	// 这里只处理单测点表达式
	kvs := utils.ParseKvString(e.Mapping, consts.SepExpr)
	if len(kvs) != 1 {
		return false, "", ""
	}
	for k, pointNo := range kvs {
		// 需要引用的测点为本模板的非计算测点
		if _, ok := pointNoSet[pointNo]; !ok {
			return false, "", ""
		}
		return true, pointNo, k
	}
	return false, "", ""
}

const (
	// CmdExpression 为ProtocolDefinition里Command标记用表达式计算
	CmdExpression string = "_expression"
)

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
