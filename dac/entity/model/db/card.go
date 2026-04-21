// Package db 定义门禁系统的数据库模型和表结构。
package db

// CardType 卡类型（长期卡/临时卡）
type CardType int

// 卡类型常量
const (
	CardTypeLongTerm  = CardType(0) // 长期卡
	CardTypeTemporary = CardType(1) // 临时卡
)

// CardFlagType 卡状态标志（启用/禁用）
type CardFlagType int

// 卡状态标志常量
const (
	CardFlagEnable  = CardFlagType(0) // 启用
	CardFlagDisable = CardFlagType(1) // 禁用
)

// Card 卡信息
type Card struct {
	CardNo    string       ` json:"card_no" gorm:"primaryKey;column:card_no"` // 卡号
	CardFlag  CardFlagType `json:"card_flag"`                                 // 卡状态标志
	CardType  CardType     `json:"card_type"`                                 // 卡类型
	ValidTime int64        `json:"valid_time"`                                // 有效期时间戳
	StaffID   IDType       `json:"staff_id" gorm:"index;column:staff_id" `    // 关联人员ID
	MozuID    string       `json:"mozu_id" gorm:"primaryKey;column:mozu_id"`  // 模组ID
}

// TableName 返回卡信息表名
func (Card) TableName() string {
	return "t_dac_card"
}

// CardAccessRelation 门禁卡权限关联信息
type CardAccessRelation struct {
	ID            IDType `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	AccessGroupID IDType `gorm:"index" json:"access_group_id"`       // 权限组ID
	CardNo        string `gorm:"index" json:"card_no"`               // 卡号
	MozuID        string `gorm:"index;column:mozu_id"`               // 模组ID
}

// TableName 返回卡权限关联表名
func (CardAccessRelation) TableName() string {
	return "t_card_access_relation"
}
