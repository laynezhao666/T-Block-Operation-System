// Package cgi 定义门禁CGI接口的请求和响应数据模型。
package cgi

import (
	"dac/entity/model/db"
)

// Event 门禁进出事件记录
type Event struct {
	ControllerID   db.IDType `json:"controller_id"`   // 控制器ID
	ControllerName string    `json:"controller_name"` // 控制器名称
	Index          int       `json:"index"`           // 事件索引
	Time           string    `json:"time"`            // 事件时间
	CardNumber     string    `json:"card_number"`     // 卡号
	Username       string    `json:"username"`        // 用户名
	DoorNumber     int       `json:"door_number"`     // 门编号
	DoorName       string    `json:"door_name"`       // 门名称
	Company        string    `json:"company"`         // 公司
	Direction      string    `json:"direction"`       // 进出方向
	Type           int       `json:"type"`            // 事件类型
	Description    string    `json:"desc"`            // 事件描述
}
