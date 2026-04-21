// Package cgi 定义CGI接口的请求和响应数据模型。
package cgi

import "dac/entity/model/db"

// RequestInfo 请求详情，包含控制器基本信息
type RequestInfo struct {
	ID             db.IDType `json:"id"`              // 请求ID
	ControllerID   db.IDType `json:"controller_id"`   // 控制器ID
	Method         string    `json:"method"`          // 请求方法
	Message        string    `json:"message"`         // 响应消息
	CreateTime     int64     `json:"create_time"`     // 创建时间
	AccessTime     int64     `json:"access_time"`     // 访问时间
	MozuID         string    `json:"mozu_id"`         // 模组ID
	ControllerName string    `json:"controller_name"` // 控制器名称
	ControllerIP   string    `json:"controller_ip"`   // 控制器IP
}

// AsynchronousInfo 异步请求信息，用于前端展示
type AsynchronousInfo struct {
	ID             db.IDType `json:"id"`              // 请求ID
	Method         string    `json:"method"`          // 请求方法
	CreateTime     string    `json:"create_time"`     // 创建时间
	AccessTime     string    `json:"access_time"`     // 访问时间
	ControllerName string    `json:"controller_name"` // 控制器名称
	State          string    `json:"state"`           // 请求状态
	Payload        string    `json:"payload"`         // 请求载荷
}

// Requests 异步请求列表响应
type Requests struct {
	Total int64              `json:"total"` // 总数
	List  []AsynchronousInfo `json:"list"`  // 请求列表
}
