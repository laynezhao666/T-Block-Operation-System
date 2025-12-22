package model

// MessageKey 测点消息队列的key
type MessageKey struct {
	MozuId    string `json:"mID"`   // 模组ID
	DeviceID  string `json:"dID"`   // 设备ID
	WorkerID  string `json:"wID"`   // worker id, 生成uuid
	Seq       uint64 `json:"seq"`   // 序列号
	Timestamp int64  `json:"t"`     // 采集时间戳，单位秒
	Interval  int32  `json:"d"`     // 测点类型，60:周期性,1:变化测点
	PubMs     int64  `json:"pubMs"` // 投递kafka的毫秒时间戳，单位毫秒
}

// MessageValue 测点消息队列的value
type MessageValue struct {
	Interval      int32   `json:"interval"`       // 上报周期
	BoxID         string  `json:"box_id"`         // TBox ID
	Points        []Point `json:"points"`         // 测点数据组
	VirtualPoints []Point `json:"virtual_points"` // 虚拟测点,暂未使用
}

// Point 测点信息
type Point struct {
	Name      string `json:"i"` // 测点名称
	Value     string `json:"v"` // 测点值
	Quality   string `json:"q"` // 质量
	Timestamp string `json:"t"` // 时间戳，单位秒
}
