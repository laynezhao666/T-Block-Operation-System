// Package driver 定义门禁协议的数据模型和通用类型。
package driver

import "dac/entity/model/db"

// TimeGroup 时间组信息
type TimeGroup struct {
	GroupNo  int        `json:"group_no" gorm:"column:group_no"`                           // 时间组编号
	Week     WeekDay    `json:"week,omitempty" gorm:"column:week;serializer:json"`         // 生效星期
	TimeZone []TimeZone `json:"timezone,omitempty" gorm:"column:timeZone;serializer:json"` // 时间段列表
}

// WeekDay 星期列表类型
type WeekDay = []int

// TimeZone 时间段（起止时间）
type TimeZone struct {
	Begin string `json:"begin"` // 开始时间
	End   string `json:"end"`   // 结束时间
}

// Card 门禁卡信息
type Card struct {
	CardNo      string `json:"card_no" gorm:"column:card_no"`                     // 卡号
	CardFlag    int    `json:"card_flag" gorm:"column:card_flag"`                 // 卡状态标志
	DoorNos     []int  `json:"door,omitempty" gorm:"column:door;serializer:json"` // 授权门列表
	TimeGroupNo int    `json:"group_no" gorm:"column:group_no"`                   // 时间组编号
	UserName    string `json:"user_name" gorm:"column:user_name"`                 // 用户名
	Password    string `json:"password" gorm:"column:password"`                   // 密码
}

// CardWithStaffInfo 带人员信息的门禁卡
type CardWithStaffInfo struct {
	Card
	UserID      db.IDType `json:"user_id"`      // 用户ID
	FaceImage   string    `json:"face_image"`   // 人脸图片
	FingerPrint string    `json:"finger_print"` // 指纹数据
}

// UserID 用户ID结构
type UserID struct {
	UserID db.IDType `json:"user_id"` // 用户ID
}

// CardWithIntNo 卡号为整数类型的门禁卡（兼容旧版协议）
type CardWithIntNo struct {
	CardNo      int64  `json:"card_no"`        // 卡号（整数）
	CardFlag    int    `json:"card_flag"`      // 卡状态标志
	DoorNos     []int  `json:"door,omitempty"` // 授权门列表
	TimeGroupNo int    `json:"group_no"`       // 时间组编号
	UserName    string `json:"user_name"`      // 用户名
	Password    string `json:"password"`       // 密码
}

// Door 门基本信息
type Door struct {
	No   int    `json:"door_no"`   // 门编号
	Name string `json:"door_name"` // 门名称
}

// DoorParam 门参数
type DoorParam struct {
	Door
	Password       string `json:"password"`         // 门密码
	KeepOpenTime   int    `json:"keep_time"`        // 门开保持时间（秒）
	OpenTimeout    int    `json:"open_time"`        // 门开超时时间（秒）
	LockErrorCount int    `json:"lock_err_cnt"`     // 卡封锁错误次数
	SlotInterval   int    `json:"slot_interval"`    // 非法卡刷卡间隔（秒）
	LockTime       int    `json:"lock_time"`        // 非法卡封锁时间（秒）
	OpenMode       int    `json:"open_mode"`        // 开门模式
	FireSignalMode int    `json:"fire_signal_mode"` // 火警信号模式
}

// ErrorCode 错误码结构
type ErrorCode struct {
	ErrCode    int    `json:"err_code"` // 错误码
	ErrMessage string `json:"err_msg"`  // 错误信息
}

// Error 实现error接口，返回错误信息
func (e *ErrorCode) Error() string {
	return e.ErrMessage
}

// Code 返回错误码
func (e *ErrorCode) Code() int {
	return e.ErrCode
}
