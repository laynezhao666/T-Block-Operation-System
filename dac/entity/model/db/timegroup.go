// Package db 定义门禁系统的数据库模型和表结构。
package db

// TimeGroup 时间组数据库模型
type TimeGroup struct {
	ID         int    `gorm:"primaryKey;autoIncrement"`                     // 主键
	GroupNo    int    `gorm:"index;unique;column:group_no" json:"group_no"` // 时间组编号
	GroupName  string // 时间组名称
	Week       string // 星期的json数据结构
	TimeZone   string // 时段的json数据结构
	UpdateTime int64  // 更新时间
}

// TableName 返回时间组表名
func (TimeGroup) TableName() string {
	return "t_dac_time_group"
}
