package dto

// CondCollectorGetDeviceList 获取采集设备列表查询条件
type CondCollectorGetDeviceList struct {
	DeviceGid          []string // 设备GID
	DeviceNumber       []string // 设备编码
	ParentDeviceNumber []string // 父级设备编码
	CollectorType      []int32  // 采集类型,1:Tbox,2: Tbox下子设备，3：厂商采集器，4：厂商采集器子设备
	Page               int
	Size               int
}
