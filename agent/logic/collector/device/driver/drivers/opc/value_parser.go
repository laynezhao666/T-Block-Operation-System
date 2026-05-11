package opc

import (
	"agent/entity/definition/datatype"
	"agent/logic/collector/device/model"
	"agent/utils"
)

// ValueParser 解析器
type ValueParser struct {
	Addr     string
	Extend   string
	DataType datatype.DataType
}

// NewOpcuaValueParser 新建解析器
func NewOpcuaValueParser(params *model.ValParseParams) *ValueParser {
	if params == nil {
		return nil
	}

	return &ValueParser{
		Addr:     params.DataAddr,
		Extend:   params.Extend,
		DataType: utils.GetDataType(params.DataType, nil, nil),
	}
}
