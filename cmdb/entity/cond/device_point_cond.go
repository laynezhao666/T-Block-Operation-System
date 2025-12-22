package cond

// ListDevicePointCond 获取标准测点列表查询条件
type ListDevicePointCond struct {
	DeviceGid       []string // Gid列表
	DeviceNumber    []string // 设备编号列表
	BelongCollector []string // 归属采集器列表
	PointNameEn     []string // 测点中文
	PointNameZh     []string // 测点英文
	PointKey        []string // 测点标识
	PointType       []int32  // 测点类型
	PointRw         []string // 测点只读
	PointLevel      []string // 测点级别
	MozuId          []int32  // 模组ID列表
	Page            int32    // 页
	Size            int32    //页大小
}
