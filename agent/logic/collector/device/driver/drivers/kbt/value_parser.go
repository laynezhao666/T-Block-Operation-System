package kbt

import (
	"fmt"
	"strings"
)

// ValueParser KBT1000协议解析器
// LineNum: 线路编号 (1-64)
// Extend: 扩展参数
// DataType: 数据类型
// LineName: 线路名称
// Status: 线路状态 (true=接地故障, false=正常)
type ValueParser struct {
	LineNum   uint32
	Extend    string
	DataType  string
	LineName  string
	Status    bool
}

// ReadFrom 从KBT1000协议帧中读取线路状态值
// KBT1000协议中，线路状态通过位图方式表示，每个位对应一条线路的接地故障状态
func (vp *ValueParser) ReadFrom(payload []byte) (any, error) {
	if vp == nil {
		return nil, fmt.Errorf("nil ValueParser")
	}

	// KBT1000协议中，线路编号范围是1-64
	if vp.LineNum < 1 || vp.LineNum > 64 {
		return nil, fmt.Errorf("KBT line number out of range: %d, must be 1-64", vp.LineNum)
	}

	// 根据DataType处理不同的值类型
	switch strings.ToUpper(strings.TrimSpace(vp.DataType)) {
	case "BOOL", "BIT":
		// 布尔值：0=正常，1=接地故障
		// 这里返回布尔值，但实际应用中可能需要转换为整数
		return vp.Status, nil
	case "INT", "INTEGER":
		// 整数值：0=正常，1=接地故障
		if vp.Status {
			return 1, nil
		}
		return 0, nil
	case "STRING", "STR":
		// 字符串值："正常"或"接地故障"
		if vp.Status {
			return "接地故障", nil
		}
		return "正常", nil
	default:
		// 默认返回整数值
		if vp.Status {
			return 1, nil
		}
		return 0, nil
	}
}

// GetLineName 获取线路名称
func (vp *ValueParser) GetLineName() string {
	if vp.LineName != "" {
		return vp.LineName
	}

	// 默认线路名称映射
	lineNames := map[uint32]string{
		1: "母线1", 2: "母线2", 3: "母线3", 4: "母线4",
		5: "线路1", 6: "线路2", 7: "线路3", 8: "线路4",
		9: "线路5", 10: "线路6", 11: "线路7", 12: "线路8",
		13: "线路9", 14: "线路10", 15: "线路11", 16: "线路12",
		17: "线路13", 18: "线路14", 19: "线路15", 20: "线路16",
		21: "线路17", 22: "线路18", 23: "线路19", 24: "线路20",
		25: "线路21", 26: "线路22", 27: "线路23", 28: "线路24",
		29: "线路25", 30: "线路26", 31: "线路27", 32: "线路28",
		33: "线路29", 34: "线路30", 35: "线路31", 36: "线路32",
		37: "线路33", 38: "线路34", 39: "线路35", 40: "线路36",
		41: "线路37", 42: "线路38", 43: "线路39", 44: "线路40",
		45: "线路41", 46: "线路42", 47: "线路43", 48: "线路44",
		49: "线路45", 50: "线路46", 51: "线路47", 52: "线路48",
		53: "线路49", 54: "线路50", 55: "线路51", 56: "线路52",
		57: "线路53", 58: "线路54", 59: "线路55", 60: "线路56",
		61: "线路57", 62: "线路58", 63: "线路59", 64: "线路60",
	}

	if name, exists := lineNames[vp.LineNum]; exists {
		return name
	}

	return fmt.Sprintf("线路%d", vp.LineNum)
}

// SetLineStatus 设置线路状态
func (vp *ValueParser) SetLineStatus(status bool) {
	vp.Status = status
	vp.LineName = vp.GetLineName()
}

// GetLineStatus 获取线路状态描述
func (vp *ValueParser) GetLineStatus() string {
	if vp.Status {
		return "接地故障"
	}
	return "正常"
}