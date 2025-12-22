package device

import (
	"encoding/json"
	"strconv"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/logic/collector/device/model"
)

// AnalogValueDesc 模拟量值描述
type AnalogValueDesc struct {
	ScaleEnable bool
	Scale       definition.FloatType
	Offset      definition.FloatType
}

// Parse 解析 valDef 对应的模拟量值描述
func (a *AnalogValueDesc) Parse(point *model.TemplateInstancePointInfo) bool {
	if a == nil {
		return false
	}
	// 模拟量值定义可以为空，为空则均使用默认值
	if point == nil {
		return true
	}

	var tempObj struct {
		Scale  string `json:"scale"`
		Offset string `json:"offset"`
	}

	valMap, ok := point.ValueDef.(map[string]interface{})
	if !ok {
		log.Errorf("point ValueDef struct convert err:%v", point.ValueDef)
		return false
	}

	if off, ok := valMap["offset"]; ok {
		tempObj.Offset = off.(string)
	}
	if scale, ok := valMap["scale"]; ok {
		tempObj.Scale = scale.(string)
	}

	if tempObj.Scale != "" {
		scale, err := strconv.ParseFloat(tempObj.Scale, definition.FloatTypeSize)
		if err != nil {
			log.Errorf("scale value parse err:%v", err)
			return true //  可忽略该错误
		}
		a.Scale = definition.FloatType(scale)

		if tempObj.Scale == "1" && (tempObj.Offset == "" || tempObj.Offset == "0") {
			a.ScaleEnable = false
		} else {
			a.ScaleEnable = true
		}
	}

	if tempObj.Offset != "" && tempObj.Offset != "0" {
		offset, err := strconv.ParseFloat(tempObj.Offset, definition.FloatTypeSize)
		if err != nil {
			log.Errorf("offset value parse err:%v", err)
			return true //  可忽略该错误
		}
		a.Offset = definition.FloatType(offset)

		// 仅有偏移量有值时，Scale默认为1
		a.ScaleEnable = true
		if tempObj.Scale == "" {
			a.Scale = 1
		}
	}

	return true
}

// DigitalValDesc 状态量值描述，k 值对应描述为 v，k仅为0或1
type DigitalValDesc map[string]string

// Parse 解析 valDef 对应的数字量值描述，valDef 形式为 {"val1":"desc1","val2":"desc2"}
func (d *DigitalValDesc) Parse(valDef interface{}) bool {
	// 数字量值定义不能为空
	if d == nil || valDef == nil {
		return false
	}

	// 尝试将 valDef 转换为 map[string]interface{}
	valMap, ok := valDef.(map[string]interface{})
	if !ok {
		return false
	}

	jsonData, err := json.Marshal(valMap)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(jsonData, d); err != nil {
		log.Errorf("DigitalValDesc Parse Error unmarshalling:%v", err)
		return false
	}

	return true
}

// EnumValDesc 枚举量值描述，k 值对应描述为 v
type EnumValDesc map[string]string

// Parse 解析 valDef 对应的数字量值描述，valDef 形式为 {"val1":"desc1","val2":"desc2"}
func (d *EnumValDesc) Parse(valDef interface{}) bool {
	// 数字量值定义不能为空
	if d == nil || valDef == nil {
		return false
	}

	// 尝试将 valDef 转换为 map[string]interface{}
	valMap, ok := valDef.(map[string]interface{})
	if !ok {
		return false
	}

	jsonData, err := json.Marshal(valMap)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(jsonData, d); err != nil {
		log.Errorf("EnumValDesc Parse Error unmarshalling:%v", err)
		return false
	}

	return true
}
