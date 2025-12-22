package dto

// CondMozuInfoGetList 模组信息查询条件
type CondMozuInfoGetList struct {
	MozuId   []int32  // 模组ID
	MozuName []string // 模组名称
	MozuCode []string // 模组编码
}
