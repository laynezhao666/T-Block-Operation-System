package model

// RTAlarm 测试实时告警信息
type RTAlarm struct {
	// 是否告警
	Status bool `json:"status"`
	// 告警等级
	Level int `json:"level"`
	// 告警时间
	Tms int64 `json:"tms"`
}

// NewRTAlarm 创建实时告警信息
func NewRTAlarm() RTAlarm {
	return RTAlarm{
		Status: false,
		Level:  0,
		Tms:    0,
	}
}
