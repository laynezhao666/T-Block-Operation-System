package model

import (
	"fmt"
	"time"
)

// CollectorTemplatePoint 采集模版测点
type CollectorTemplatePoint struct {
	Id           int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	TemplateName string    `gorm:"column:template_name;type:varchar(127);comment:模版名称;not null;" json:"template_name"`                            // 模版名称
	MozuId       int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	SubDevice    string    `gorm:"column:sub_device;type:varchar(127);comment:子设备名称;not null;" json:"sub_device"`                                 // 子设备名称
	PointNameEn  string    `gorm:"column:point_name_en;type:varchar(255);comment:测点名称英文;not null;" json:"point_name_en"`                          // 测点名称英文
	PointNameZh  string    `gorm:"column:point_name_zh;type:varchar(255);comment:测点名称中文;not null;" json:"point_name_zh"`                          // 测点名称中文
	PointType    string    `gorm:"column:point_type;type:varchar(16);comment:测点类型;not null;" json:"point_type"`                                   // 测点类型
	PointRw      string    `gorm:"column:point_rw;type:varchar(16);comment:测点读写分类;not null;" json:"point_rw"`                                     // 测点读写分类
	DeltaDef     string    `gorm:"column:delta_def;type:varchar(1024);comment:变化定义规则;not null;" json:"delta_def"`                                 // 变化定义规则
	VerifyDef    string    `gorm:"column:verify_def;type:varchar(1024);comment:校验规则;not null;" json:"verify_def"`                                 // 校验规则
	ExpDef       string    `gorm:"column:exp_def;type:text;comment:表达式定义规则;" json:"exp_def"`                                                      // 表达式定义规则
	ProtDef      string    `gorm:"column:prot_def;type:text;comment:协议定义规则;" json:"prot_def"`                                                     // 协议定义规则
	ValDef       string    `gorm:"column:val_def;type:varchar(1024);comment:值定义规则;not null;" json:"val_def"`                                      // 值定义规则
	Simulator    string    `gorm:"column:simulator;type:varchar(255);comment:模拟定义规则;not null;" json:"simulator"`                                  // 模拟定义规则
	CreateAt     time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt     time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 采集测点表名称
func (CollectorTemplatePoint) TableName() string {
	return "t_collector_template_point"
}

// CalcUniqueKey 计算采集测点唯一标识
func (t CollectorTemplatePoint) CalcUniqueKey() string {
	return fmt.Sprintf("%s|%s|%s", t.TemplateName, t.PointNameEn, t.SubDevice)
}
