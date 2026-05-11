package ping

import (
	"agent/logic/collector/device/model"
)

// PingValueParser ping值解析器
// ping驱动的值解析比较简单，返回的结果就是质量值（0表示正常，非0表示异常）
type PingValueParser struct {
	DataAddr  string
	DataType  string
	ByteOrder string
	Extend    string
}

// NewPingValueParser 创建ping值解析器
func NewPingValueParser(params *model.ValParseParams) *PingValueParser {
	if params == nil {
		return &PingValueParser{}
	}
	return &PingValueParser{
		DataAddr:  params.DataAddr,
		DataType:  params.DataType,
		ByteOrder: params.ByteOrder,
		Extend:    params.Extend,
	}
}
