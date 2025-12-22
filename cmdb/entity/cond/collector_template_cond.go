package cond

// ListCollectorTemplateCond 获取采集模板列表
type ListCollectorTemplateCond struct {
	TemplateName []string // 模板名称
	ProtocolType []string // 协议类型
	MozuId       []int32  //模组ID
	Page         int      // 页数
	Size         int      // 每页大小
}
