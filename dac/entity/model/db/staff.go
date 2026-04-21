// Package db 定义门禁系统的数据库模型和表结构。
package db

// StaffType 人员类型
type StaffType int

// 人员相关常量
const (
	DefaultStaffID   = IDType(-1)    // 默认人员ID（未绑定）
	StaffTypeDeleted = StaffType(-1) // 已删除人员类型
)

// StaffBase 人员基础信息
type StaffBase struct {
	ID   IDType `gorm:"primaryKey;index;unique;autoIncrement" json:"id"` // 主键
	Name string `json:"name"`                                            // 姓名
}

// Staff 人员完整信息
type Staff struct {
	StaffBase
	Password    string    `json:"password"`                      // 密码
	Fingerprint string    `json:"fingerprint"`                   // 指纹
	Picture     string    `json:"picture"`                       // 照片
	Sex         string    `json:"sex"`                           // 性别
	Phone       string    `json:"phone"`                         // 电话
	Email       string    `json:"email"`                         // 邮箱
	Company     string    `json:"company"`                       // 公司
	PaperType   string    `json:"paper_type"`                    // 证件类型
	Paper       string    `json:"paper"`                         // 证件号
	Comment     string    `json:"comment"`                       // 备注
	Type        StaffType `json:"type" gorm:"column:type"`       // 人员类型
	MozuID      string    `json:"mozu_id" gorm:"column:mozu_id"` // 模组ID
}

// TableName 返回人员表名
func (Staff) TableName() string {
	return "t_dac_staff"
}

// CardAndStaffBase 卡和人员的简要信息
type CardAndStaffBase struct {
	CardNo string    `json:"card_no"` // 卡号
	Staff  StaffBase `json:"staff"`   // 人员基础信息
}
