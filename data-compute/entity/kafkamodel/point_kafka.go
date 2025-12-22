// Package kafkamodel kafka消息发送相关实体
package kafkamodel

// KafkaMsgPoint 测点信息
type KafkaMsgPoint struct {
	I string `json:"i"` // 测点名称
	V string `json:"v"` // 测点值
	Q string `json:"q"` // 质量？
	T string `json:"t"` // 时间戳
}

// KafkaMsgKey 测点kafka的key
type KafkaMsgKey struct {
	MID   string `json:"mID"`   // 模组ID
	DID   string `json:"dID"`   // 设备ID
	WID   string `json:"wID"`   // WorkerID
	Seq   int    `json:"seq"`   // 序号
	T     int64  `json:"t"`     // 推送时间S
	D     int32  `json:"d"`     // 测点周期
	BKey  string `json:"bKey"`  // 业务key
	PubMs int64  `json:"pubMs"` // 推送毫秒时间
	Type  int32  `json:"type"`  // 测点类型
}

// KafkaMsgValue 测点kafka的value
type KafkaMsgValue struct {
	Interval      int64            `json:"interval"`
	BoxID         string           `json:"box_id"` // TBox ID
	Points        []*KafkaMsgPoint `json:"points"` // 测点数据组
	VirtualPoints []*KafkaMsgPoint `json:"virtual_points"`
}
