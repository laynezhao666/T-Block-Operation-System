package dto

// CondDeviceEntityGetList 设备实体-GetList查询条件
type CondDeviceEntityGetList struct {
	DeviceGid          []string // 设备GID
	DeviceNumber       []string // 设备编号
	ParentDeviceNumber []string // 父设备编号
	ApplicationTypeEn  []string // 应用类型英文
	ApplicationTypeZh  []string // 应用类型中文
	Page               int      // 页
	Size               int      // 页大小
}

// DeviceTreeNode 设备树节点
type DeviceTreeNode struct {
	DeviceGid               string
	DeviceNumber            string
	DeviceNumberShow        string
	DeviceNo                string
	DeviceName              string
	EnableStatus            int32
	DeviceTypeEn            string
	DeviceTypeZh            string
	ApplicationTypeEn       string
	ApplicationTypeZh       string
	BelongApplicationTypeEn string
	Children                []*DeviceTreeNode
}
