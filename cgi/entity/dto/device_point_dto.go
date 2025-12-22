package dto

// CondGetDevicePointList 设备测点-GetList查询条件
type CondGetDevicePointList struct {
	DeviceGid    []string // 设备GID
	DeviceNumber []string // 设备编号
	PointNameEn  []string // 测点英文名称
	PointNameZh  []string // 测点中文名称
	PointKey     []string // 测点唯一标识
	PointRw      []string // 测点读写范围
	PointLevel   []string // 测点级别
	ValueType    []string // 测点值类型
	Page         int      // 页
	Size         int      // 页大小
}
