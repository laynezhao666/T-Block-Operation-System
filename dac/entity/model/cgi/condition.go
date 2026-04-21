// Package cgi 定义CGI接口的请求条件模型。
package cgi

// TimeCondition 时间范围查询条件
type TimeCondition struct {
	BeginTime int64 `json:"begin_time"` // 开始时间戳
	EndTime   int64 `json:"end_time"`   // 结束时间戳
}

// OffsetCondition 分页查询条件
type OffsetCondition struct {
	Offset int `json:"offset"` // 偏移量
	Limit  int `json:"limit"`  // 每页数量
}

// QueryCondition 组合查询条件，包含分页和时间范围
type QueryCondition struct {
	OffsetCondition
	TimeCondition
}
