// Package cgi 定义CGI接口的请求和响应数据模型。
package cgi

// Rooms 机房列表响应模型
type Rooms struct {
	MozuID   int         `json:"mozu_id"`   // 模组ID
	MozuName string      `json:"mozu_name"` // 模组名称
	Rooms    interface{} `json:"rooms"`     // 机房列表
}
