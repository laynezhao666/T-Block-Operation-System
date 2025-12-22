package model

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

// PointType 测点类型枚举值
type PointType int

const (
	AnalogType  PointType = 0 // 模拟量，浮点型、整型
	DigitalType PointType = 1 // 状态量，布尔值
	EnumType    PointType = 2 // 枚举量

	AnalogTypeString  string = "A" // 模拟量类型对应标记
	DigitalTypeString string = "D" // 状态量类型对应标记（实际为bool型，会0、1状态，注意此处elvdb的原始命名错误，只能兼容）
	BoolTypeString    string = "B" // 布尔量类型对应标记
	EnumTypeString    string = "E" // 枚举量类型对应标记

	AccessWrite string = "W" // 可写权限
)

// ValParseParams 描述如何从设备响应报文中解析测点值
// 包括测点在响应报文中的地址、类型、字节序等信息
type ValParseParams struct {
	DataAddr  string
	DataType  string
	ByteOrder string
	Extend    string
}

// PointAttr 测点属性
type PointAttr struct {
	ID   definition.DataPointIDType
	Type PointType
	// 值描述
	ValDesc interface{}
	// 用于解析值的对象
	ValParser interface{}
}

// PointInfo 测点信息
type PointInfo struct {
	// 测点属性
	Attr PointAttr
	// 测点值
	RtVal model.RTValue
}

// ListPoints 测点列表
type ListPoints []*PointInfo

// TemplateInstancePointInfo 实例化的采集测点信息
type TemplateInstancePointInfo struct {
	// 访问权限
	Access string `json:"access"`
	// 分类
	Class string `json:"cls"`
	// 测点 ID
	ID definition.DataPointIDType `json:"id"`
	// 测点名称
	Name string `json:"name"`
	// 采集点表达式定义
	ExprDef ExpressionDefinition `json:"expdef"`
	// 协议定义
	ProtocolDef ProtocolDefinition `json:"protdef"`
	// 值定义
	ValueDef interface{} `json:"valdef"`
	// 测点类型
	ValueType string `json:"valtype"`
	// 模拟数据
	SimulatorDef interface{} `json:"simulator"`
	// 所属虚拟子设备
	SubDevice string `json:"sub_device"`
	// 是否北向定义
	IsNorthDef string `json:"is_north"`
	// 有效范围
	ValueRange string `json:"value_range"`
	// 变化死区
	ValueDeadZone string `json:"value_deadzone"`
}

// InstancePointsInfo 实例化的采集测点信息
type InstancePointsInfo []TemplateInstancePointInfo

// GetDataPoints 获取测点数据
func (pi InstancePointsInfo) GetDataPoints() model.DataPoints {
	l := len(pi)
	points := make(model.DataPoints, l)
	for i := range points {
		points[i].ID = pi[i].ID
	}
	return points
}

// StdInstancePointInfo 标准化测点
type StdInstancePointInfo struct {
	// 标准设备
	StdDevice string `json:"device_gid"`
	// 标准测点
	StdPoint string `json:"point_name_en"`
	// 标准测点中文名
	StdPointZh string `json:"point_name_zh"`
	// 变化阈值（绝对值）
	Threshold string `json:"threshold"`
	// 映射表达式
	Expr string `json:"expression"`
	// 映射
	Mapping string `json:"expression_map"`
	// 可读的映射
	MappingZh string `json:"expression_map_zh"`
	// 映射参数
	Param map[string]string
	// 测点值类型(数据类型：模拟量、状态量)
	ValueType string `json:"value_type"`
	// 是否启用(0:禁用, 1:启用)
	Enable int32 `json:"point_kpi"`
	// 测点值有效范围
	ValueValidRange string `json:"value_valid_range"`
	// 测点值单位
	ValueUnit string `json:"value_unit"`
	// 测点值精度
	ValuePrecision string `json:"value_precision"`
	// 值枚举映射
	ValueEnum string `json:"value_enum"`
	// 读写
	PointRw string `json:"point_rw"`
	// 等级
	PointLevel string `json:"point_level"`
}

// StdInstancePointsInfo 标准化测点信息
type StdInstancePointsInfo []StdInstancePointInfo

// GetDataPoints 获取测点数据
func (spi StdInstancePointsInfo) GetDataPoints() model.DataPoints {
	l := len(spi)
	points := make(model.DataPoints, l)
	for i := range points {
		points[i].ID = definition.DataPointIDType(spi[i].StdDevice + spi[i].StdPoint)
		points[i].PointType = definition.StdPointType
	}
	return points
}

// StdCommon 存放通用标准化信息
type StdCommon struct {
	Prefix string `json:"prefix"`
}
