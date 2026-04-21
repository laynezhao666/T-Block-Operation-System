// Package cgi 定义CGI接口的请求和响应数据模型。
package cgi

import (
	"dac/entity/model/db"
)

// Door CGI门信息，扩展了状态点位ID
type Door struct {
	db.Door
	StateID string `json:"state_id"` // 门状态点位ID
}

// DoorController CGI控制器信息，包含门列表和监测点位ID
type DoorController struct {
	db.DoorController
	Doors   []Door `json:"doors"`    // 门列表
	CommID  string `json:"comm_id"`  // 通讯状态点位ID
	FaultID string `json:"fault_id"` // 故障状态点位ID
}
