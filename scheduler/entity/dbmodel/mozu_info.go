// Package dbmodel 模组信息表实体定义
package dbmodel

// MozuInfo 模组信息
type MozuInfo struct {
	MozuId         int32  `gorm:"column:mozu_id;type:int(11);comment:模组ID;not null;" json:"mozu_id"`                     // 模组ID
	PublishVersion string `gorm:"column:publish_version;type:varchar(32);comment:下发版本;not null;" json:"publish_version"` // 下发版本
	AlarmVersion   string `gorm:"column:alarm_version;type:varchar(32);comment:下发版本;not null;" json:"alarm_version"`     // 下发版本
}

// TableName 模组信息表名称
func (MozuInfo) TableName() string {
	return "t_mozu_info"
}
