package simulator

import (
	"agent/entity/definition/datatype"
	"agent/logic/collector/device/driver/drivers/simulator/generator"
)

// ValueParser 采集测点数据解析器
type ValueParser struct {
	DataType  datatype.DataType
	Generator *generator.Generator
}
