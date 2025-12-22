// Package dbmodel 告警策略表实体定义
package dbmodel

import (
	"fmt"
)

// AlarmStrategy 告警策略对应的数据库实体
type AlarmStrategy struct {
	Id                int64  `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                      // 主键ID
	DeviceGid         string `gorm:"column:device_gid;type:varchar(127);comment:设备GID;not null;" json:"device_gid"`              // 设备GID
	Rid               int64  `gorm:"column:rid;type:bigint(20);comment:策略ID;not null;default:0;" json:"rid"`                     // 策略ID
	RidVersion        string `gorm:"column:rid_version;type:bigint(20);comment:策略版本;not null;default:0;" json:"rid_version"`     // 策略版本
	RidType           int64  `gorm:"column:rid_type;type:tinyint(4);comment:策略类型,0:实时,1:延时;not null;default:0;" json:"rid_type"` // 策略类型,0:实时,1:延时
	MozuId            int64  `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`              // 所属模组ID
	AlarmName         string `gorm:"column:alarm_name;type:varchar(127);comment:告警名称;not null;" json:"alarm_name"`               // 告警名称
	AlarmExpression   string `gorm:"column:alarm_expression;type:text;comment:告警表达式;not null;" json:"alarm_expression"`          // 告警表达式
	RestoreExpression string `gorm:"column:restore_expression;type:text;comment:恢复表达式;not null;" json:"restore_expression"`      // 恢复表达式
	ExpressionMap     string `gorm:"column:expression_map;type:text;comment:表达式映射;not null;" json:"expression_map"`              // 表达式映射
	AlarmLevel        string `gorm:"column:alarm_level;type:varchar(8);comment:告警级别;not null;" json:"alarm_level"`               // 告警级别
	ContentTemplate   string `gorm:"column:content_template;type:text;comment:告警内容模版;not null;" json:"content_template"`         // 告警内容模版
	ComputeCost       int64  `gorm:"column:compute_cost;type:int(11);comment:计算复杂度;not null;default:0;" json:"compute_cost"`     // 计算复杂度
}

// TableName 告警策略配置表名
func (c *AlarmStrategy) TableName() string {
	return "t_alarm_strategy"
}

// CalcCombineKey 告警策略唯一标识
func (c *AlarmStrategy) CalcCombineKey() string {
	return fmt.Sprintf("%d;%s;%s", c.Rid, c.DeviceGid, c.RidVersion)
}
