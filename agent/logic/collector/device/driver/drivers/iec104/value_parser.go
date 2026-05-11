package iec104

import (
	"strconv"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition/datatype"
	"agent/logic/collector/device/model"
	"agent/utils"
)

// ValueParser 解析器
type ValueParser struct {
	Addr     uint32
	Extend   string
	DataType datatype.DataType
}

// NewIEC104ValueParser 新建解析器
func NewIEC104ValueParser(params *model.ValParseParams) *ValueParser {
	if params == nil {
		return nil
	}

	num, err := strconv.ParseInt(strings.TrimSpace(params.DataAddr), 0, 32)
	if err != nil {
		log.Warnf("iec104 new value parser error: %v, params: %v", err, *params)
		return nil
	}
	return &ValueParser{
		Addr:     uint32(num),
		Extend:   params.Extend,
		DataType: utils.GetDataType(params.DataType, nil, nil),
	}
}
