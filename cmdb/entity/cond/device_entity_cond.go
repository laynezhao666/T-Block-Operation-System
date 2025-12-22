package cond

// ListDeviceEntityCond 获取标准设备列表查询参数
type ListDeviceEntityCond struct {
	DeviceGid               []string // Gid列表
	DeviceNumber            []string // 设备编号列表
	ParentDeviceNumber      []string // 父级设备列表
	EnableStatus            []int32  // 启用状态
	MozuId                  []int32  // 模组ID列表
	IdcArea                 []string //IDC区域
	DeviceTypeEn            []string // 设备种类英文
	DeviceTypeZh            []string // 设备种类中文
	ApplicationTypeEn       []string // 应用类型英文
	ApplicationTypeZh       []string // 应用类型中文
	BelongApplicationTypeEn []string // 所属应用类型
	Page                    int32    // 页
	Size                    int32    // 页大小
}
