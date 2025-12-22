package model

// ChannelData 通道数据
type ChannelData struct {
	ChannelID     string `json:"chid"`
	ChannelParams string `json:"chparams"`
	// 设备地址
	Address string `json:"addr"`
	// 协议版本
	ProtocolVersion string `json:"prot_ver"`
	CmdInterval     int    `json:"cmd_interval"`
	WaitTimeMs      int    `json:"wait_time"`
	TimeoutMs       int    `json:"timeout"`
	// 并发协程数
	ParallelCount       int `json:"parallel_count"`
	PacketMaxPointCount int `json:"packet_max_point_count"`
	// 扩展参数
	Extend string `json:"extend"`
	Chtype string `json:"chtype"`
}
