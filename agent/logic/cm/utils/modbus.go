package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	boolDataType = "bool"

	defaultMaxRegisterNumber = 80
	validMaxRegisterNumber   = 125
	defaultMaxRegisterGap    = 10
)

const (
	ValueDefKey      = "valdef"
	AlarmDefKey      = "almdef"
	ProtocolDescKey  = "protdef"
	ExpressionDefKey = "expdef"
	DeltaDefKey      = "deltadef"
)

var (
	MaxRegisterNumber = defaultMaxRegisterNumber
	MaxCoilNumber     = MaxRegisterNumber << 4
	MaxRegisterGap    = defaultMaxRegisterGap
	MaxCoilGap        = MaxRegisterGap << 4
)

// IntervalData 指定功能码下的区间元数据
type IntervalData struct {
	intervals                Intervals   // 每个测点对应的寄存器区间
	intervalIndex            int         // 临时值，表示当前访问的区间
	intervalIndex2pointIndex map[int]int // 每个区间对应的测点索引
}

// verifyCmd 校验采集指令
func verifyCmd(cmd string) (bool, error) {
	cmd = strings.ToUpper(cmd)
	for _, c := range cmd {
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false, fmt.Errorf("非法指令: '%c' in \"%v\"", c, cmd)
	}

	needGenerateCmd := false
	switch len(cmd) {
	case 2:
		needGenerateCmd = true
	case 10:
		// 自动生成的指令与手动填写的指令不合并
		needGenerateCmd = false
	default:
		return false, fmt.Errorf("非法指令: \"%v\"", cmd)
	}
	return needGenerateCmd, nil
}

// getIntervalLen 获取该数据类型对应区间长度
func getIntervalLen(datatype string) (int, error) {
	datatype = strings.ToLower(datatype)
	switch datatype {
	case "int16", "uint16":
		return 1, nil
	case "int32", "uint32", "int", "uint", "float":
		return 2, nil
	case "int64", "uint64", "double":
		return 4, nil
	}

	if !strings.HasPrefix(datatype, boolDataType) {
		return 0, fmt.Errorf("数据类型错误: \"%v\"", datatype)
	}

	// example: bool
	if len(datatype) == len(boolDataType) {
		return 1, nil
	}

	// example: bool3
	if !strings.Contains(datatype, ":") {
		pos := datatype[len(boolDataType):]
		v, err := strconv.ParseInt(pos, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("解析比特位置错误: %w", err)
		}
		if v < 0 || v > 0xf {
			return 0, fmt.Errorf("比特位置非法: %v", v)
		}
		return 1, nil
	}

	// example: bool2:4
	s := datatype[len(boolDataType):]
	ss := strings.Split(s, ":")
	if len(ss) == 2 {
		b, err := strconv.ParseInt(ss[0], 0, 32)
		if err != nil {
			return 0, err
		}
		e, err := strconv.ParseInt(ss[1], 0, 32)
		if err != nil {
			return 0, err
		}
		if 0 <= b && b < e && e <= 16 {
			// 取多个 bit 时，只能是取寄存器
			bitLen := int(e) - int(b)
			return (bitLen + 7) >> 3, nil
		}
	}

	return 0, fmt.Errorf("错误的数据类型: %v", datatype)
}

func setCommand(point MapObject, cmd string) error {
	data, err := GetMapValue(point, ProtocolDescKey)
	if err != nil {
		return err
	}
	data[CommandAndFuncCodeField] = cmd
	return nil
}

func fillCommand(
	funcCode string, intervalData *IntervalData, result *MergeResult, points []MeasurePoint,
) error {
	for oldIntervalIndex, newIntervalIndex := range result.IndexMap {
		in := result.Intervals[newIntervalIndex]
		pointIndex, ok := intervalData.intervalIndex2pointIndex[oldIntervalIndex]
		if !ok {
			return fmt.Errorf("功能码: %v, 未找到 %v 对应的映射", funcCode, oldIntervalIndex)
		}
		cmd := fmt.Sprintf("%v%04X%04X", funcCode, in.Begin, in.End-in.Begin)
		if len(cmd) != 10 {
			return fmt.Errorf("生成了错误指令: %v, 测点: %+v", cmd, points[pointIndex])
		}
		if err := setCommand(points[pointIndex], cmd); err != nil {
			return err
		}
	}
	return nil
}

// ShouldTryGenerateCommands 是否需要自动生成指令
func ShouldTryGenerateCommands(protocol string) bool {
	return strings.HasPrefix(protocol, MODBUS)
}

// GenerateCommands 根据寄存器与功能码自动生成采集指令
func GenerateCommands(points []MeasurePoint) error {
	cmdIntervals := make(map[string]*IntervalData)
	for i, point := range points {
		protocol, err := GetMapValue(point, ProtocolDescKey)
		if err != nil {
			return fmt.Errorf("point %+v error: %w", point, err)
		}

		cmd, err := GetStringValue(protocol, CommandAndFuncCodeField)
		if err != nil {
			return fmt.Errorf("point %+v error: %w", point, err)
		}
		needGenerate, err := verifyCmd(cmd)
		if err != nil {
			return fmt.Errorf("%+v 中的采集指令错误: %w", point, err)
		}
		if !needGenerate {
			continue
		}

		intervalData, ok := cmdIntervals[cmd]
		if !ok {
			// 如果该功能码下不存在区间数据，则初始化对应数据
			intervalData = &IntervalData{
				intervals:                nil,
				intervalIndex:            0,
				intervalIndex2pointIndex: make(map[int]int),
			}
			cmdIntervals[cmd] = intervalData
		}

		dataType, err := GetStringValue(protocol, DataTypeField)
		if err != nil {
			return fmt.Errorf("point %+v error: %w", point, err)
		}
		l, err := getIntervalLen(dataType)
		if err != nil {
			return fmt.Errorf("%+v 中的数据类型错误: %w", point, err)
		}

		regStr, err := GetStringValue(protocol, RegAddrField)
		if err != nil {
			return fmt.Errorf("point %+v error: %w", point, err)
		}
		reg, err := strconv.ParseInt(regStr, 0, 32)
		if err != nil {
			return fmt.Errorf("point %+v error: %w", point, err)
		}

		// 该测点对应的寄存器区间
		in := Interval{Begin: int(reg), End: int(reg) + l}

		intervalData.intervalIndex2pointIndex[intervalData.intervalIndex] = i
		intervalData.intervals = append(intervalData.intervals, in)
		intervalData.intervalIndex++
	}

	return fillCommands(points, cmdIntervals)
}

// fillCommands 根据寄存器区间进行合并，生成采集指令
func fillCommands(points []MeasurePoint, cmdIntervals map[string]*IntervalData) error {
	var r *MergeResult
	var err error
	var maxRange, maxGap int
	// 处理每个功能码及对应区间
	for funcCode, intervalData := range cmdIntervals {
		switch funcCode {
		case "01", "02":
			maxRange = MaxCoilNumber
			maxGap = MaxCoilGap
		case "03", "04":
			maxRange = MaxRegisterNumber
			maxGap = MaxRegisterGap
		default:
			return fmt.Errorf("不支持的功能码: %v", funcCode)
		}
		if r, err = GenerateIntervals(intervalData.intervals, maxRange, maxGap); err != nil {
			return err
		}
		if err = fillCommand(funcCode, intervalData, r, points); err != nil {
			return err
		}
	}

	return nil
}
