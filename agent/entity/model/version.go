package model

// ConfigChangeEvent 设备变更结构体
type ConfigChangeEvent struct {
	CollectorChanged []string // 采集配置变更设备
	StdChanged       []string // 标准点配置变更设备
}
type FileVersion struct {
	Timestamp   int64
	Sequence    int64
	FullVersion string
}
