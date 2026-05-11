// Package dlt DL/T 645-2007 多功能电能表通信协议驱动
package dlt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"agent/entity/consts"
	"agent/entity/definition/datatype"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/byteorder"
)

// DLTValueParser DL/T 645-2007 值解析器
type DLTValueParser struct {
	// 数据地址/偏移量（在响应数据中的偏移）
	Offset int
	// 数据类型
	DataType datatype.DataType
	// 字节序
	ByteOrder byteorder.ByteOrderExtend
	// 小数位数
	Decimals int
	// 扩展参数
	Extend string
}

// NewDLTValParser 创建DLT值解析器
func NewDLTValParser(params *model.ValParseParams) *DLTValueParser {
	parser := &DLTValueParser{
		Offset:    0,
		DataType:  datatype.FloatType,
		ByteOrder: byteorder.LittleEndianExtend, // DL/T 645默认低字节在前
		Decimals:  -1,                           // -1表示未配置，不处理精度
		Extend:    params.Extend,
	}

	// 解析数据地址（偏移量）
	if params.DataAddr != "" {
		if offset, err := parseOffset(params.DataAddr); err == nil {
			parser.Offset = offset
		}
	}

	// 解析数据类型
	var bitBegin, bitEnd uint8
	parser.DataType = utils.GetDataType(params.DataType, &bitBegin, &bitEnd)

	// 解析字节序
	switch params.ByteOrder {
	case consts.ValueBigEndian:
		parser.ByteOrder = byteorder.BigEndianExtend
	case consts.ValueLittleEndian:
		parser.ByteOrder = byteorder.LittleEndianExtend
	case consts.ValueBigSwap:
		parser.ByteOrder = byteorder.BigEndianSwapExtend
	case consts.ValueLittleSwap:
		parser.ByteOrder = byteorder.LittleEndianSwapExtend
	}

	// 解析扩展参数中的小数位数
	if params.Extend != "" {
		parser.parseExtend(params.Extend)
	}

	return parser
}

// parseExtend 解析扩展参数
func (p *DLTValueParser) parseExtend(extend string) {
	// 支持格式: "decimals=2" 或 "dec=4"
	parts := strings.Split(extend, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.TrimSpace(kv[1])

		switch key {
		case "decimals", "dec":
			if dec, err := strconv.Atoi(value); err == nil && dec >= 0 {
				p.Decimals = dec
			}
		}
	}
}

// ParseValue 解析测点值
// data: 响应数据域（已减去0x33）
// dataId: 数据标识
func (p *DLTValueParser) ParseValue(data []byte, dataId uint32) (interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short")
	}

	// 跳过前4字节的数据标识
	// 响应数据格式: DI0 DI1 DI2 DI3 + 数据
	if len(data) <= 4+p.Offset {
		return nil, fmt.Errorf("data length insufficient for offset %d", p.Offset)
	}

	valueData := data[4+p.Offset:]

	// 根据数据标识类型解析
	return p.parseByDataId(valueData, dataId)
}

// parseByDataId 根据数据标识解析数据
// DL/T 645-2007 数据标识格式: DI3 DI2 DI1 DI0
// dataId: 0xDI3DI2DI1DI0 (如 0x02010100 表示A相电压)
func (p *DLTValueParser) parseByDataId(data []byte, dataId uint32) (interface{}, error) {
	// 根据DL/T 645-2007附录A的数据格式定义
	// 不同的数据标识有不同的数据格式

	di0 := byte(dataId & 0xFF)
	di1 := byte((dataId >> 8) & 0xFF)
	di2 := byte((dataId >> 16) & 0xFF)
	di3 := byte((dataId >> 24) & 0xFF)

	// 根据DI3判断数据大类
	switch di3 {
	case 0x00:
		// 电能量类数据
		// DI3=00: 组合有功/无功电能、正向/反向有功/无功电能等
		// 格式: XXXXXX.XX (4字节BCD码，单位kWh/kvarh)
		return p.parseEnergy(data)

	case 0x01:
		// 最大需量类数据
		// DI3=01: 正向/反向有功/无功最大需量
		// 格式: XX.XXXX (3字节BCD码) + 时间
		return p.parseDemand(data)

	case 0x02:
		// 变量类数据（瞬时量）
		// DI3=02: 电压、电流、有功功率、无功功率、功率因数、相角、频率等
		return p.parseInstantByDI2DI1(data, di2, di1)

	case 0x03:
		// 事件记录类数据
		return p.parseBCD(data, p.getDecimals())

	case 0x04:
		// 参变量类数据
		return p.parseParamData(data, di2, di1, di0)

	default:
		// 其他数据类型，默认BCD解析
		return p.parseBCD(data, p.getDecimals())
	}
}

// parseInstantByDI2DI1 根据DI2和DI1解析瞬时量数据
// DI3=02时的变量类数据解析
func (p *DLTValueParser) parseInstantByDI2DI1(data []byte, di2, di1 byte) (float64, error) {
	switch di2 {
	case 0x01:
		// DI2=01: 电压
		switch di1 {
		case 0x01, 0x02, 0x03:
			// A/B/C相电压 XXX.X V (2字节BCD)
			if len(data) < 2 {
				return 0, errors.New("voltage data too short")
			}
			return p.parseBCD(data[:2], 1)
		default:
			return p.parseBCD(data, p.getDecimals())
		}

	case 0x02:
		// DI2=02: 电流
		switch di1 {
		case 0x01, 0x02, 0x03:
			// A/B/C相电流 XXX.XXX A (3字节BCD)
			if len(data) < 3 {
				return 0, errors.New("current data too short")
			}
			return p.parseBCD(data[:3], 3)
		default:
			return p.parseBCD(data, p.getDecimals())
		}

	case 0x03:
		// DI2=03: 瞬时有功功率
		// XX.XXXX kW (3字节BCD，带符号)
		if len(data) < 3 {
			return 0, errors.New("active power data too short")
		}
		return p.parseBCDSigned(data[:3], 4)

	case 0x04:
		// DI2=04: 瞬时无功功率
		// XX.XXXX kvar (3字节BCD，带符号)
		if len(data) < 3 {
			return 0, errors.New("reactive power data too short")
		}
		return p.parseBCDSigned(data[:3], 4)

	case 0x05:
		// DI2=05: 瞬时视在功率
		// XX.XXXX kVA (3字节BCD)
		if len(data) < 3 {
			return 0, errors.New("apparent power data too short")
		}
		return p.parseBCD(data[:3], 4)

	case 0x06:
		// DI2=06: 功率因数
		// X.XXX (2字节BCD，带符号)
		if len(data) < 2 {
			return 0, errors.New("power factor data too short")
		}
		return p.parseBCDSigned(data[:2], 3)

	case 0x07:
		// DI2=07: 相角
		// XXX.X ° (2字节BCD，带符号)
		if len(data) < 2 {
			return 0, errors.New("phase angle data too short")
		}
		return p.parseBCDSigned(data[:2], 1)

	case 0x08:
		// DI2=08: 电压波形失真度
		// XX.XX % (2字节BCD)
		if len(data) < 2 {
			return 0, errors.New("voltage THD data too short")
		}
		return p.parseBCD(data[:2], 2)

	case 0x09:
		// DI2=09: 电流波形失真度
		// XX.XX % (2字节BCD)
		if len(data) < 2 {
			return 0, errors.New("current THD data too short")
		}
		return p.parseBCD(data[:2], 2)

	case 0x0A:
		// DI2=0A: 电压相位角
		// XXX.X ° (2字节BCD)
		if len(data) < 2 {
			return 0, errors.New("voltage phase angle data too short")
		}
		return p.parseBCD(data[:2], 1)

	case 0x0B:
		// DI2=0B: 电流相位角
		// XXX.X ° (2字节BCD)
		if len(data) < 2 {
			return 0, errors.New("current phase angle data too short")
		}
		return p.parseBCD(data[:2], 1)

	case 0x80:
		// DI2=80: 频率 XX.XX Hz (2字节BCD)
		if len(data) < 2 {
			return 0, errors.New("frequency data too short")
		}
		return p.parseBCD(data[:2], 2)

	default:
		return p.parseBCD(data, p.getDecimals())
	}
}

// parseParamData 解析参变量类数据
// DI3=04时的参变量数据
func (p *DLTValueParser) parseParamData(data []byte, di2, di1, di0 byte) (float64, error) {
	// 根据具体参变量类型解析
	// 大部分参变量为BCD格式，根据具体需求扩展
	return p.parseBCD(data, p.getDecimals())
}

// getDecimals 获取小数位数，未配置时返回0（不处理精度）
func (p *DLTValueParser) getDecimals() int {
	if p.Decimals < 0 {
		return 0
	}
	return p.Decimals
}

// parseEnergy 解析电能量数据
// 格式: XXXXXX.XX (4字节BCD码，单位kWh)
func (p *DLTValueParser) parseEnergy(data []byte) (float64, error) {
	if len(data) < 4 {
		return 0, errors.New("energy data too short")
	}
	return p.parseBCD(data[:4], 2)
}

// parseDemand 解析需量数据
// 格式: XX.XXXX (3字节BCD码)
func (p *DLTValueParser) parseDemand(data []byte) (float64, error) {
	if len(data) < 3 {
		return 0, errors.New("demand data too short")
	}
	return p.parseBCD(data[:3], 4)
}

// parseEnergy 解析电能量数据
// decimals: 小数位数
func (p *DLTValueParser) parseBCD(data []byte, decimals int) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("empty data")
	}

	// BCD码转换为整数值
	// 低字节在前（小端）
	var value int64 = 0
	multiplier := int64(1)

	for i := 0; i < len(data); i++ {
		b := data[i]
		low := int64(b & 0x0F)
		high := int64((b >> 4) & 0x0F)

		// 检查BCD码是否有效
		if low > 9 || high > 9 {
			return 0, fmt.Errorf("invalid BCD byte: %02X", b)
		}

		value += low * multiplier
		multiplier *= 10
		value += high * multiplier
		multiplier *= 10
	}

	// 应用小数位数
	result := float64(value)
	for i := 0; i < decimals; i++ {
		result /= 10.0
	}

	return result, nil
}

// parseBCDSigned 解析带符号的BCD码数据
// 最高字节的最高位为符号位
func (p *DLTValueParser) parseBCDSigned(data []byte, decimals int) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("empty data")
	}

	// 获取符号位（最高字节的最高位）
	lastByte := data[len(data)-1]
	negative := (lastByte & 0x80) != 0

	// 清除符号位
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	dataCopy[len(dataCopy)-1] &= 0x7F

	value, err := p.parseBCD(dataCopy, decimals)
	if err != nil {
		return 0, err
	}

	if negative {
		value = -value
	}

	return value, nil
}

// GetPointValParser 获取测点值解析器
func GetPointValParser(p *model.PointInfo) (*DLTValueParser, error) {
	if p == nil {
		return nil, errors.New("point is nil")
	}

	valueParser, ok := p.Attr.ValParser.(*DLTValueParser)
	if !ok || valueParser == nil {
		return nil, errors.New("dlt value parser is not configured")
	}
	return valueParser, nil
}

// parseOffset 解析偏移量
func parseOffset(dataAddr string) (int, error) {
	dataAddr = strings.TrimSpace(dataAddr)
	if strings.HasPrefix(strings.ToLower(dataAddr), "0x") {
		val, err := strconv.ParseInt(dataAddr[2:], 16, 32)
		return int(val), err
	}
	val, err := strconv.ParseInt(dataAddr, 10, 32)
	return int(val), err
}
