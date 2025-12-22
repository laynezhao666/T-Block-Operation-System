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
}
