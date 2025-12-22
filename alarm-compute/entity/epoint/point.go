package epoint

import (
	"fmt"
	"strings"
)

const (
	// PointKeySeperator PointKeySeperator
	PointKeySeperator string = "###"
	// PointKeyMinuteFlag PointKeyMinuteFlag
	PointKeyMinuteFlag string = "60"
	// PointKeySecondFlag PointKeySecondFlag
	PointKeySecondFlag string = "1"

	// PointSepNum PointSepNum
	PointSepNum int = 3
)

const (
	// DevicePointeperator DevicePointeperator
	DevicePointeperator string = "."

	// DevicePointSepNum DevicePointSepNum
	DevicePointSepNum int = 2
)

const (
	// AnalyzePointValueEmpty 测点值空字符替代
	AnalyzePointValueEmpty = -1999999
	// HBasePointValueEmpty hbase测点数据为空数据
	HBasePointValueEmpty = "--"
)

// AlarmPointValueValidate 测点数据有效性
func AlarmPointValueValidate(val float64) (bool, float64) {
	// 自定义的测点没有值
	if val == float64(AnalyzePointValueEmpty) {
		return false, val
	}
	// 北向不支持的测点
	if val == float64(-9999) {
		return false, val
	}
	// 本地监控启动初始值
	if val == float64(-99999) {
		return false, val
	}
	// -99998:设备通讯中断,-99997:测点北向中未定义,其他:保留区间
	if val >= float64(-99998) && val <= float64(-99990) {
		return false, val
	}
	return true, val
}

// IsMinutePointData 是否分钟级的测点数据
func IsMinutePointData(key string) (minute bool, err error) {
	if len(key) == 0 {
		err = fmt.Errorf("key is empty")
		return
	}

	sep := strings.Split(key, PointKeySeperator)
	if len(sep) != PointSepNum {
		err = fmt.Errorf("split num is not %v, sep: %v", PointSepNum, sep)
		return
	}

	if sep[PointSepNum-1] == PointKeyMinuteFlag {
		minute = true
		return
	}

	return
}
