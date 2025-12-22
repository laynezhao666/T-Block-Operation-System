package point

import (
	"fmt"
	"strings"

	"alarm-compute/entity/epoint"
)

// AlarmPointValueValidate 测点数据有效性
func (m *PointManager) AlarmPointValueValidate(val float64) (bool, float64) {
	// 自定义的测点没有值
	if val == float64(epoint.AnalyzePointValueEmpty) {
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
func (m *PointManager) IsMinutePointData(key string) (minute bool, err error) {
	if len(key) == 0 {
		err = fmt.Errorf("key is empty")
		return
	}

	sep := strings.Split(key, epoint.PointKeySeperator)
	if len(sep) != epoint.PointSepNum {
		err = fmt.Errorf("split num is not %v, sep: %v", epoint.PointSepNum, sep)
		return
	}

	if sep[epoint.PointSepNum-1] == epoint.PointKeyMinuteFlag {
		minute = true
		return
	}

	return
}
