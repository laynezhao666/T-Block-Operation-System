package rt

import (
	"dac/entity/model/db"
)

type DoorController struct {
	db.DoorController
	Doors []db.Door `json:"doors"`
}

type FetchInterval struct {
	FetchEventInterval     float64 `json:"fetch_event_interval"`
	FetchLoopEventInterval float64 `json:"fetch_loop_event_interval"`
	FetchAlarmInterval     float64 `json:"fetch_alarm_interval"`
	FetchLoopAlarmInterval float64 `json:"fetch_loop_alarm_interval"`
}

type DoorControllerItemWithEnable struct {
	Name            string `xlsx:"0"`
	Vendor          string `xlsx:"1"`
	Model           string `xlsx:"2"`
	SN              string `xlsx:"3"`
	Room            string `xlsx:"4"`
	Block           string `xlsx:"5"`
	No              string `xlsx:"6"`
	ChannelID       string `xlsx:"7"`
	Timeout         string `xlsx:"8"`
	ProtocolName    string `xlsx:"9"`
	ProtocolVersion string `xlsx:"10"`
	Account         string `xlsx:"11"`
	Password        string `xlsx:"12"`
	Enable          string `xlsx:"13"`
	CommandInterval string `xlsx:"14"`
	DoorNum         string `xlsx:"15"`
	URLMode         string `xlsx:"16"`
	Extend          string `xlsx:"17"`
}

type ChannelLink struct {
	ChType       string `json:"chtype"`         // 通道类型，如 "socket"
	ChID         string `json:"chid"`           // 通道ID，如 "169.49.69.162:8080"
	ChParams     string `json:"chparams"`       // 通道参数
	Addr         string `json:"addr"`           // 地址
	WaitTime     string `json:"wait_time"`      // 等待时间
	CmdInterval  string `json:"cmd_interval"`   // 命令间隔
	Timeout      string `json:"timeout"`        // 超时时间，如 "3000"
	MaxFailCount string `json:"max_fail_count"` // 最大失败次数
	MaxFailTime  string `json:"max_fail_time"`  // 最大失败时间
}

type Extend struct {
	ProtocolName     string `json:"protocol_name"`
	Protocol_version string `json:"protocol_version"`
}
