package model

// HeartbeatMessage 数据上报client向数据接收方上报的心跳消息
type HeartbeatMessage struct {
	Box       BoxInfo `json:"box"`
	Timestamp int64   `json:"timestamp"`
}

// BoxInfo t-box相关信息
type BoxInfo struct {
	ClientId string `json:"client_id"`
	Ip       string `json:"ip"`
}
