package cond

// ListCollectorDeviceCond 获取采集设备列表查询参数
type ListCollectorDeviceCond struct {
	DeviceGid          []string // Gid列表
	DeviceNumber       []string // 设备编号列表
	DeviceSn           []string // 设备Sn列表
	ParentDeviceNumber []string // 父级设备列表
	DeviceTypeEn       []string // 设备类型列表
	CollectorType      []int32  // 采集器类型列表
	TemplateName       []string // 模版名称列表
	MozuId             []int32  // 模组ID列表
	Page               int32    // 页
	Size               int32    //页大小
}
