package model

import (
	"encoding/json"
	"etrpc-go/log"
	"fmt"
	"time"
)

// AlarmStrategy 告警策略对应的数据库实体
type AlarmStrategy struct {
	Id                   int64     `gorm:"column:id;type:bigint(20);comment:主键ID;primaryKey;not null;" json:"id"`                                         // 主键ID
	DeviceGid            string    `gorm:"column:device_gid;type:varchar(127);comment:设备GID;not null;" json:"device_gid"`                                 // 设备GID
	Rid                  int64     `gorm:"column:rid;type:bigint(20);comment:策略ID;not null;default:0;" json:"rid"`                                        // 策略ID
	RidVersion           string    `gorm:"column:rid_version;type:varchar(127);comment:策略版本;not null;default:'';" json:"rid_version"`                        // 策略版本
	RidType              int32     `gorm:"column:rid_type;type:tinyint(4);comment:策略类型,0:实时,1:延时;not null;default:0;" json:"rid_type"`                    // 策略类型,0:实时,1:延时
	MozuId               int32     `gorm:"column:mozu_id;type:int(11);comment:所属模组ID;not null;default:0;" json:"mozu_id"`                                 // 所属模组ID
	AlarmName            string    `gorm:"column:alarm_name;type:varchar(127);comment:告警名称;not null;" json:"alarm_name"`                                  // 告警名称
	AlarmExpression      string    `gorm:"column:alarm_expression;type:text;comment:告警表达式;not null;" json:"alarm_expression"`                             // 告警表达式
	AlarmExpressionStr   string    `gorm:"column:alarm_expression_str;type:text;comment:告警表达式(中文);not null;" json:"alarm_expression_str"`                 // 告警表达式(中文)
	RestoreExpression    string    `gorm:"column:restore_expression;type:text;comment:恢复表达式;not null;" json:"restore_expression"`                         // 恢复表达式
	RestoreExpressionStr string    `gorm:"column:restore_expression_str;type:text;comment:恢复表达式(中文);not null;" json:"restore_expression_str"`             // 恢复表达式(中文)
	ExpressionMap        string    `gorm:"column:expression_map;type:text;comment:表达式映射;not null;" json:"expression_map"`                                 // 表达式映射
	AlarmLevel           string    `gorm:"column:alarm_level;type:varchar(8);comment:告警级别;not null;" json:"alarm_level"`                                  // 告警级别
	ContentTemplate      string    `gorm:"column:content_template;type:text;comment:告警内容模版;not null;" json:"content_template"`                            // 告警内容模版
	Owner                string    `gorm:"column:owner;type:varchar(32);comment:告警负责人;not null;" json:"owner"`                                            // 告警负责人
	ComputeCost          int32     `gorm:"column:compute_cost;type:int(11);comment:计算复杂度;not null;default:0;" json:"compute_cost"`                        // 计算复杂度
	CreateAt             time.Time `gorm:"<-:false;column:create_at;type:datetime;comment:记录创建时间;not null;default:CURRENT_TIMESTAMP;" json:"create_at"`   // 记录创建时间
	UpdateAt             time.Time `gorm:"<-:false;column:update_at;type:datetime;comment:记录上次更新时间;not null;default:CURRENT_TIMESTAMP;" json:"update_at"` // 记录上次更新时间
}

// TableName 告警策略表名称
func (t AlarmStrategy) TableName() string {
	return "t_alarm_strategy"
}

// CalcUniqueKey 告警策略唯一标识
func (t AlarmStrategy) CalcUniqueKey() string {
	return fmt.Sprintf("%s|%d|%s", t.DeviceGid, t.Rid, t.RidVersion)
}

// ExpressionMapping 表达式映射实体
type ExpressionMapping struct {
	Fire struct {
		ExprMap map[string]string `json:"expr_map"`
		Engine  string            `json:"engine,omitempty"`
	} `json:"fire"`
	Restore struct {
		ExprMap map[string]string `json:"expr_map"`
		Engine  string            `json:"engine,omitempty"`
	} `json:"restore"`
}

// AlarmStrategyLevelStat 告警策略级别统计实体
type AlarmStrategyLevelStat struct {
	Rid        int64  `gorm:"column:rid"`
	AlarmLevel string `gorm:"column:alarm_level"`
	Cnt        int32  `gorm:"column:cnt"`
}

// VariableGidMap (告警/恢复)变量映射
type VariableGidMap struct {
	ExprMap map[string]string `json:"expr_map"`
	Engine  string            `json:"engine"`
}

// ExprMap (告警/恢复)表达式映射
type ExprMap struct {
	Fire    VariableGidMap `json:"fire"`
	Restore VariableGidMap `json:"restore"`
}

// GetExprMap 获取策略测点映射
func (a *AlarmStrategy) GetExprMap() *ExprMap {
	if a == nil || len(a.ExpressionMap) <= 0 {
		return &ExprMap{}
	}
	var exp = &ExprMap{}
	err := json.Unmarshal([]byte(a.ExpressionMap), exp)
	if err != nil {
		log.Errorf("GetExprMap Unmarshal err: %v, exp:%v", err, a.ExpressionMap)
		return &ExprMap{}
	}
	return exp
}
