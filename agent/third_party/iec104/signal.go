package iec104

// Signal 104信号
type Signal struct {
	TypeID  uint        `json:"type_id"` // 类型id，1:单点遥信，9:单点遥测等
	Address uint32      `json:"address"` // 地址
	Value   interface{} `json:"value"`   // 值
	Quality byte        `json:"quality"` // 品质描述
	Ts      int64       `json:"ts"`      // 上游数据上报的时间戳，单位毫秒
	Cts     int64       `json:"cts"`     // 本地采集的数据时间戳，单位毫秒
}
