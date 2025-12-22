package modbus

import (
	"errors"
	"fmt"
	"agent/utils"
	byteorder2 "agent/utils/byteorder"
	"strconv"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition/datatype"
	"agent/logic/collector/device/model"
)

// ModbusValueParser Modbus解析器
type ModbusValueParser struct {
	Addr      uint16
	Extend    string
	ByteOrder byteorder2.ByteOrderExtend
	DataType  datatype.DataType
	BitBegin  uint8
	BitEnd    uint8
}

// NewModbusValParser 创建解析器
func NewModbusValParser(params *model.ValParseParams) *ModbusValueParser {
	var byteOrder byteorder2.ByteOrderExtend = byteorder2.BigEndianExtend
	switch params.ByteOrder {
	case consts.ValueBigEndian:
		byteOrder = byteorder2.BigEndianExtend
	case consts.ValueLittleEndian:
		byteOrder = byteorder2.LittleEndianExtend
	case consts.ValueBigSwap:
		byteOrder = byteorder2.BigEndianSwapExtend
	case consts.ValueLittleSwap:
		byteOrder = byteorder2.LittleEndianSwapExtend
	}

	addr, err := parseAddress(params.DataAddr)
	if err != nil {
		return nil
	}

	parser := ModbusValueParser{
		Addr:      uint16(addr),
		Extend:    params.Extend,
		ByteOrder: byteOrder,
	}
	parser.DataType = utils.GetDataType(params.DataType, &parser.BitBegin, &parser.BitEnd)
	return &parser
}

// reg支持16进制设置
func parseAddress(dataAddr string) (int64, error) {
	if strings.HasPrefix(strings.ToLower(dataAddr), "0x") {
		return strconv.ParseInt(dataAddr[2:], 16, 64)
	}
	return strconv.ParseInt(dataAddr, 10, 64)
}

// GetPointValParser 获取解析器
func GetPointValParser(p *model.PointInfo) (*ModbusValueParser, error) {
	if p == nil {
		return nil, errors.New("point is nil")
	}

	valueParser, ok := p.Attr.ValParser.(*ModbusValueParser)
	if !ok || valueParser == nil {
		log.Warnf("modbus value parser is not configured, point_id=%v", p.Attr.ID)
		return nil, fmt.Errorf("modbus value parser is not configured, point_id=%s", p.Attr.ID)
	}
	return valueParser, nil
}
