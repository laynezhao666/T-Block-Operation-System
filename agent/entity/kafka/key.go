package kafka

// KafkaKey kafka消息key
type KafkaKey struct {
	MozuID      string `json:"mID"`
	DeviceGiD   string `json:"dID"`
	WorkerID    string `json:"wID"`
	Seq         uint64 `json:"seq"`
	Timestamp   int64  `json:"t"`
	Interval    int    `json:"d"`
	BalancerKey string `json:"bKey"`  // 分区hash Key
	PubMs       int64  `json:"pubMs"` // 投递时的毫秒时间戳
	Type        int    `json:"type"`  // 1-采集; 2-标准; 3-虚拟; 4-告警测点
	CiID        string `json:"ciID"`  // 采集器id
	N           int    `json:"n"`     // 测点数
}
