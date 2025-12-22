package model

import (
	"time"

	"trpc.group/trpc-go/trpc-go"
)

// AlarmWorker 告警管理worker信息，记录占用的workerId等信息
type AlarmWorker struct {
	Id           int64     `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`                       // 主键ID
	SetId        int32     `gorm:"column:set_id;NOT NULL" json:"set_id"`                                 // 片区Id
	WorkerId     int32     `gorm:"column:worker_id;NOT NULL" json:"worker_id"`                           // 占用的workerId
	OccupyStatus int32     `gorm:"column:occupy_status;NOT NULL" json:"occupy_status"`                   // id占用状态 0: 未占用 1:已占用
	Uuid         string    `gorm:"column:uuid;NOT NULL" json:"uuid"`                                     // 占用workerId所使用的唯一标识
	PodIp        string    `gorm:"column:pod_ip;NOT NULL" json:"pod_ip"`                                 // podIp
	HeartBeat    time.Time `gorm:"column:heartbeat;default:0;NOT NULL" json:"heartbeat"`                 // 最近一次上报心跳时间
	CreateAt     time.Time `gorm:"column:create_at;default:CURRENT_TIMESTAMP;NOT NULL" json:"create_at"` // 创建时间
}

// NewAlarmWorker 创建告警管理worker信息
func NewAlarmWorker(uid string, setId, workerId int32) *AlarmWorker {
	return &AlarmWorker{
		SetId:        setId,
		WorkerId:     workerId,
		OccupyStatus: 1,
		Uuid:         uid,
		PodIp:        trpc.GlobalConfig().Global.LocalIP,
		HeartBeat:    time.Now(),
		CreateAt:     time.Now(),
	}
}
