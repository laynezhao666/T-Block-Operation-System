package pointeval

import "alarm-compute/entity/epoint"

// AlarmTaskRet AlarmTaskRet
type AlarmTaskRet struct {
	PointValueMap        map[string]float64     `json:"pointValueMap,omitempty"`
	HistoryPointValueMap epoint.HistoryValueMap `json:"historyPointValueMap,omitempty"`
	PointMap             map[string][]string    `json:"pointMap,omitempty"`
	StartRunAt           int64                  `json:"startRunAt,omitempty"`
	ExpMap               *PointTypeMap          `json:"expMap,omitempty"`
}
