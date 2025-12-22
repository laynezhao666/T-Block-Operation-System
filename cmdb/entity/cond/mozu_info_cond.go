package cond

// ListMozuInfoCond 模组信息查询条件
type ListMozuInfoCond struct {
	MozuId           []int32 // 模组ID
	MozuName         string  // 模组名称
	MozuCode         string  // 模组编号
	MozuType         []int32 // 模组类型
	BelongCampus     string  // 所属园区
	BelongCampusCode string  // 所属园区编码
	Page             int32   // 页
	Size             int32   //页大小
}
