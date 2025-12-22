package model

import "time"

// MozuInfo 模组信息
type MozuInfo struct {
	Id               int32     `gorm:"column:id;type:int(11);comment:主键ID;primaryKey;not null;" json:"id"`                                            // 主键ID
	MozuId           int32     `gorm:"column:mozu_id;type:int(11);comment:模组ID;not null;" json:"mozu_id"`                                             // 模组ID
	MozuName         string    `gorm:"column:mozu_name;type:varchar(32);comment:模组名称;not null;" json:"mozu_name"`                                     // 模组名称
	MozuCode         string    `gorm:"column:mozu_code;type:varchar(32);comment:模组编码;not null;" json:"mozu_code"`                                     // 模组编码
	MozuType         int32     `gorm:"column:mozu_type;type:int(11);comment:模组类型;not null;default:0;" json:"mozu_type"`                               // 模组类型
	BelongBuilding   string    `gorm:"column:belong_building;type:varchar(32);comment:所属楼栋;not null;" json:"belong_building"`                         // 所属楼栋
	BelongCampus     string    `gorm:"column:belong_campus;type:varchar(32);comment:所属园区;not null;" json:"belong_campus"`                             // 所属园区
	BelongCampusCode string    `gorm:"column:belong_campus_code;type:varchar(32);comment:所属园区编码;not null;" json:"belong_campus_code"`                 // 所属园区编码
	PublishVersion   string    `gorm:"column:publish_version;type:varchar(32);comment:下发版本;not null;" json:"publish_version"`                         // 下发版本
	AlarmVersion     string    `gorm:"column:alarm_version;type:varchar(32);comment:下发版本;not null;" json:"alarm_version"`                             // 下发版本
	CreateAt         time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt         time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 模组信息表名称
func (MozuInfo) TableName() string {
	return "t_mozu_info"
}
