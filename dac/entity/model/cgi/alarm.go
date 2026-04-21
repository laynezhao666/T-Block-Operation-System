// Package cgi 定义CGI接口的请求和响应数据模型。
package cgi

import (
	"dac/entity/model/db"
)

// Alarm CGI告警信息模型
type Alarm struct {
	ControllerID   db.IDType `json:"controller_id"`   // 控制器ID
	ControllerName string    `json:"controller_name"` // 控制器名称
	Index          int       `json:"index"`           // 告警索引
	Time           string    `json:"time"`            // 告警时间
	DoorNumber     int       `json:"door_number"`     // 门编号
	DoorName       string    `json:"door_name"`       // 门名称
	Type           int       `json:"type"`            // 告警类型
	State          int       `json:"state"`           // 告警状态
	StateDesc      string    `json:"state_desc"`      // 状态描述
	Description    string    `json:"desc"`            // 告警描述
}
