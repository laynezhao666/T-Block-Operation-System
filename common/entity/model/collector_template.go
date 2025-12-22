package model

import "time"

// CollectorTemplate 采集模版
type CollectorTemplate struct {
	Id              int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	TemplateName    string    `gorm:"column:template_name;type:varchar(127);comment:模版名称;not null;" json:"template_name"`                            // 模版名称
	MozuId          int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	DeviceTypeEn    string    `gorm:"column:device_type_en;type:varchar(127);comment:设备类型英文;not null;" json:"device_type_en"`                        // 设备类型英文
	DeviceTypeZh    string    `gorm:"column:device_type_zh;type:varchar(127);comment:设备类型中文;" json:"device_type_zh"`                                 // 设备类型中文
	Manufacturer    string    `gorm:"column:manufacturer;type:varchar(127);comment:设备制造商;not null;" json:"manufacturer"`                             // 设备制造商
	DeviceModelEn   string    `gorm:"column:device_model_en;type:varchar(127);comment:设备型号;not null;" json:"device_model_en"`                        // 设备型号
	ProtocolType    string    `gorm:"column:protocol_type;type:varchar(32);comment:协议类型;not null;" json:"protocol_type"`                             // 协议类型
	ProtocolVersion string    `gorm:"column:protocol_version;type:varchar(32);comment:协议版本;not null;" json:"protocol_version"`                       // 协议版本
	ProtocolExtend  string    `gorm:"column:protocol_extend;type:varchar(1024);comment:协议扩展信息;not null;" json:"protocol_extend"`                     // 协议扩展信息
	CreateAt        time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt        time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 采集模版表名称
func (CollectorTemplate) TableName() string {
	return "t_collector_template"
}

// CalcUniqueKey 采集模版唯一标识
func (t CollectorTemplate) CalcUniqueKey() string {
	return t.TemplateName
}
