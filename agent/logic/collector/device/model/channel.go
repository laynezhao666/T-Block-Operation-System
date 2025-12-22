package model

// Channel 设备通道
type Channel struct {
	Name    string
	Params  string
	Address string
}

// ChannelInfo 描述程序如何访问设备
// 包括通道地址、打开需要的参数等
type ChannelInfo struct {
	// 名称，如: /dev/ttymxc1, 192.168.1.250:161
	Name string
	// 参数，如: 9600:8:N:1
	Params string
	// 地址，与具体协议相关
	Address string
	// 协议版本号
	ProtocolVer string
	// 与设备接口单次通讯超时
	TimeoutMs int
	// 并发协程数
	ParallelCount int
	// 请求包中允许的最大测点数
	PacketMaxPointCount int
	// 扩展参数
	ExtendKV     map[string]string
	DriverExtend string
}
