package utils

import (
	"fmt"
	"strings"
)

const (
	MODBUS           = "modbus"
	SYSDIO           = "sysdio"
	DIANZONG         = "dianzong"
	SNMP             = "snmp"
	INTEGRATEDLOCKER = "integratedlocker"
)
// ProtocolDefinitionMap 协议定义
type ProtocolDefinitionMap = map[string]interface{}
// MapObject 对象map
type MapObject = map[string]interface{}
// MeasurePoint 测点
type MeasurePoint = MapObject
// VersionInfo 版本信息
type VersionInfo struct {
	Version string `json:"ver" xlsx:"0"`
	Date    string `json:"date" xlsx:"1"`
	Desc    string `json:"desc" xlsx:"2"`
	User    string `json:"user" xlsx:"3"`
}
// NewMeasurePoint 新建测点
func NewMeasurePoint() MeasurePoint {
	return make(MeasurePoint)
}
// NewMapObject 新建对象map
func NewMapObject() MapObject {
	return make(MapObject)
}
// GetStringValue 获取字符串
func GetStringValue(data MapObject, key string) (string, error) {
	valueData, has := data[key]
	if !has {
		return "", fmt.Errorf("%+v 不存在字段 %v", data, key)
	}
	value, ok := valueData.(string)
	if !ok {
		return "", fmt.Errorf(`"%+v" 类型断言 .(string) 失败`, valueData)
	}
	return value, nil
}
// GetMapValue 获取对象map
func GetMapValue(data MapObject, key string) (MapObject, error) {
	valueData, has := data[key]
	if !has {
		return nil, fmt.Errorf("%+v 不存在字段 %v", data, key)
	}
	value, ok := valueData.(MapObject)
	if !ok {
		return nil, fmt.Errorf(`"%+v" 类型断言 .(string) 失败`, valueData)
	}
	return value, nil
}

// StringToMap 将格式为 "key=value,key=value" 的字符串转换为 map[string]string
func StringToMap(point CollectTemplatePointModel) (map[string]string, error) {
	if point.ValueType == "" {
		return nil, nil
	}
	result := make(map[string]string)

	switch point.ValueType {
	case "布尔型":
		// 使用逗号分割每个键值对
		pairs := strings.Split(point.ValueDesc, ",")
		for _, pair := range pairs {
			// 使用等号分割键和值
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid key-value pair: %s", pair)
			}

			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			if key == "" || value == "" {
				return nil, fmt.Errorf("empty key or value in pair: %s", pair)
			}

			// 构建新的键，例如 "val0", "val1" 等
			newKey := fmt.Sprintf("val%s", key)

			result[newKey] = value
		}
	case "浮点型":
		result["scale"] = point.Scale
		result["unit"] = point.ValueUnit
		result["valdesc"] = point.ValueDesc
		result["offset"] = point.Offset
	// }
	case "整数型":
		result["scale"] = point.Scale
		result["unit"] = point.ValueUnit
		result["valdesc"] = point.ValueDesc
		result["offset"] = point.Offset
	case "枚举型":
		result["valdesc"] = point.ValueDesc
	}
	return result, nil
}
// RwTrans 读写转换
func RwTrans(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	rw, ok := RwTransMap[input]
	if !ok {
		return "", fmt.Errorf("读写类型设置错误")
	}
	return rw, nil
}

var RwTransMap = map[string]string{
	"只写": "W",
	"只读": "R",
	"读写": "RW",
	"W":  "只写",
	"R":  "只读",
	"RW": "读写",
}

// 特殊转换逻辑，浮点型和整数型都转换成A
// 布尔型转换成D
var dataTypeMap = map[string]string{
	"浮点型": "A",
	"布尔型": "D",
	"整数型": "A",
	"枚举型": "E",
	"A":   "浮点型",
	"D":   "布尔型",
	// "D":   "整数型",
	"E": "枚举型",
}

// ConvertDataType 转译
func ConvertDataType(dataType string) string {
	if dataType == "" {
		return ""
	}
	newDataType, ok := dataTypeMap[dataType]
	if !ok {
		return dataType
	}
	return newDataType
}
