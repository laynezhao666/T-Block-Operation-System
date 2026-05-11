package xbm

import (
	"errors"
	"strconv"
	"strings"

	"agent/logic/collector/device/model"
)

// XBM协议命令码定义
const (
	// 读取存储数据命令
	CmdAcquireVoltage     byte = 0x50 // 读取存储电压
	CmdAcquireResistance  byte = 0x51 // 读取存储电阻
	CmdAcquireTemp1       byte = 0x52 // 读取存储温度1（R-sensor）
	CmdAcquireLoopCurrent byte = 0x53 // 读取存储环流
	CmdAcquireTemp2       byte = 0x54 // 读取存储温度2（Transformer）

	// 测量并获取数据命令
	CmdMeasureVoltage     byte = 0x70 // 测量并获取电压
	CmdMeasureResistance  byte = 0x71 // 测量并获取电阻
	CmdMeasureTemp1       byte = 0x72 // 测量并获取温度1（R-sensor）
	CmdMeasureLoopCurrent byte = 0x73 // 测量并获取环流
	CmdMeasureTemp2       byte = 0x74 // 测量并获取温度2（Transformer）
	CmdMeasureResForCal   byte = 0x75 // 测量校准电阻

	// 异常响应
	CmdAbnormalResponse byte = 0xF0
)

// XBMValueParser XBM变压器值解析器
type XBMValueParser struct {
	// SensorAddr 传感器地址 (2字节，小端序)
	SensorAddr uint16
	// Extend 扩展参数
	Extend string
}

// NewXBMValParser 创建XBM值解析器
// 地址格式: "sensor_addr"
func NewXBMValParser(params *model.ValParseParams) *XBMValueParser {
	if params == nil {
		return nil
	}

	dataAddr := strings.TrimSpace(params.DataAddr)
	if dataAddr == "" {
		return nil
	}

	sensorAddr, err := parseIntValue(dataAddr)
	if err != nil {
		return nil
	}

	return &XBMValueParser{
		SensorAddr: uint16(sensorAddr),
		Extend:     params.Extend,
	}
}

// parseIntValue 解析整数值，支持十进制和十六进制
func parseIntValue(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseInt(s[2:], 16, 64)
	}
	return strconv.ParseInt(s, 10, 64)
}

// GetPointValParser 获取测点值解析器
func GetPointValParser(p *model.PointInfo) (*XBMValueParser, error) {
	if p == nil {
		return nil, errors.New("point is nil")
	}

	valueParser, ok := p.Attr.ValParser.(*XBMValueParser)
	if !ok || valueParser == nil {
		return nil, errors.New("xbm value parser is not configured")
	}
	return valueParser, nil
}

// GetValueFormula 根据命令码获取值的转换公式描述
func GetValueFormula(cmd byte) (divisor float64, unit string) {
	switch cmd {
	case CmdAcquireVoltage, CmdMeasureVoltage:
		return 100.0, "V" // Voltage * 100
	case CmdAcquireResistance, CmdMeasureResistance, CmdMeasureResForCal:
		return 100.0, "mOhm" // Resistance * 100
	case CmdAcquireTemp1, CmdMeasureTemp1, CmdAcquireTemp2, CmdMeasureTemp2:
		return 10.0, "C" // Temperature * 10
	case CmdAcquireLoopCurrent, CmdMeasureLoopCurrent:
		return 10.0, "A" // Current * 10
	default:
		return 1.0, ""
	}
}
