// Package xbrother 定义XBrother门禁协议的请求和响应数据结构。
package xbrother

import "reflect"

// GetFieldSizeSum 计算结构体所有字段的大小总和（不含内存对齐填充）
func GetFieldSizeSum(s interface{}) int {
	v := reflect.ValueOf(s)
	size := 0

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		size += int(field.Type.Size())
	}

	return size
}

// CommonResp 通用响应
type CommonResp struct {
	Rtn uint8 // 返回码
}

// SetControllerAddrReq 设置控制器地址请求
type SetControllerAddrReq struct {
	Addr [6]uint8 // MAC地址
}

// SetControllerParamsReq 设置控制器参数请求
type SetControllerParamsReq struct {
	InterLockType  uint8  // 互锁类型: 0无互锁 1:1-2门互锁 2:3-4门互锁
	FireAlarmTime  uint16 // 火警报警时间
	AlarmTime      uint16 // 报警时间
	CoercePassword uint16 // 胁迫密码
}

// OpenDoorReq 开门请求
type OpenDoorReq struct {
	DoorNo uint8 // 门编号
}

// OpenDoorPermenentlyReq 常开门请求
type OpenDoorPermenentlyReq struct {
	DoorNo uint8 // 门编号
}

// CloseDoorReq 关门请求
type CloseDoorReq struct {
	DoorNo uint8 // 门编号
}

// LockDoorReq 锁门请求
type LockDoorReq struct {
	LockStatus uint8 // 锁状态
}

// SetTimeReq 设置时间请求
type SetTimeReq struct {
	Second uint8 // 秒
	Minute uint8 // 分
	Hour   uint8 // 时
	Week   uint8 // 星期
	Day    uint8 // 日
	Month  uint8 // 月
	Year   uint8 // 年
}

// SetDoorParamsReq 设置门参数请求
type SetDoorParamsReq struct {
	OpenDoorTime        uint16 // 开门时间（低8位在第二字节 高8位在第6字节）
	OpenDoorTimeout     uint8  // 开门超时时间
	BidirectionalDetect uint8  // 双向检测
	LongTimeOpenAlarm   uint8  // 长时间开门报警
	AlarmType           uint8  // 报警类型
	AlarmTime           uint16 // 报警时间
}

// AlarmUploadReq 告警上报请求
type AlarmUploadReq struct {
	Second  uint8 // 秒
	Minute  uint8 // 分
	Hour    uint8 // 时
	Day     uint8 // 日
	Month   uint8 // 月
	Year    uint8 // 年
	Type    uint8 // 告警类型
	Door    uint8 // 门编号
	HasNext uint8 // 是否有后续数据
	Seq     uint8 // 序列号
}

// AlarmUploadResp 告警上报响应
type AlarmUploadResp struct {
	Seq uint8 // 确认序列号
}

// EventUploadReq 事件上报请求
type EventUploadReq struct {
	CardNo  uint32 // 卡号
	Second  uint8  // 秒
	Minute  uint8  // 分
	Hour    uint8  // 时
	Day     uint8  // 日
	Month   uint8  // 月
	Year    uint8  // 年
	Type    uint8  // 事件类型
	Door    uint8  // 门编号
	HasNext uint8  // 是否有后续数据
	Seq     uint8  // 序列号
}

// EventUploadResp 事件上报响应
type EventUploadResp struct {
	Seq uint8 // 确认序列号
}

// ControllerStatusUploadReq 控制器状态上报请求
type ControllerStatusUploadReq struct {
	Reserve1       uint8  `json:"reserve1"` // 未用
	Year           uint8  `json:"year"`
	Month          uint8  `json:"month"`
	Day            uint8  `json:"day"`
	Hour           uint8  `json:"hour"`
	Minute         uint8  `json:"minute"`
	Second         uint8  `json:"second"`
	DoorStatus     uint8  `json:"door_status"`     // 从低到高前4位分别代表1，2，3，4门的状态 1表示关 0表示开
	BatchSize      uint8  `json:"batch_size"`      // 未用
	Reserve2       uint8  `json:"reserve2"`        // 未用
	FunctionType   uint8  `json:"function_type"`   // 未用
	ControllerType uint8  `json:"controller_type"` // 未用
	LockStatus     uint8  `json:"lock_status"`     // 未用
	Reserve3       uint32 `json:"reserve3"`        // 未用
	AuxiliaryRelay uint8  `json:"auxiliary_relay"` // 未用
	Version        uint8  `json:"version"`
	Reserve4       uint16 `json:"reserve4"` // 未用
}

// ClearDoorTimeGroupsReq 清除门时间组请求
type ClearDoorTimeGroupsReq struct {
	//Door uint8
}

// AddTimeGroupReq 添加时间组请求
type AddTimeGroupReq struct {
	TimeZone      uint8 // 时区编号
	StartHour     uint8 // 开始小时
	StartMinute   uint8 // 开始分钟
	EndHour       uint8 // 结束小时
	EndMinute     uint8 // 结束分钟
	WeekDay       uint8 // 星期几
	OpenDoorType  uint8 // 开门类型
	DeadlineYear  uint8 // 截止年
	DeadlineMonth uint8 // 截止月
	DeadlineDay   uint8 // 截止日
}

// CleanReq 清除数据请求
type CleanReq struct{}

// AlarmSettingReq 告警设置请求
type AlarmSettingReq struct {
	DisableAlarm    uint8 // 1关闭报警 0输出报警
	KeepEnableAlarm uint8 // 保持启用报警
}

// ClearCardsReq 清除所有卡请求
type ClearCardsReq struct{}

// DeleteCardReq 删除卡请求
type DeleteCardReq struct {
	CardIndex uint16 // 卡索引
	CardId    uint32 // 卡号（未用）
}

// AddCardReq 添加卡请求
type AddCardReq struct {
	CardIndex       uint16 // 卡索引
	CardId          uint32 // 卡号（未用）
	Password        uint16 // 密码
	AccessTimeGroup uint32 // 时间组位掩码（从低到高16位对应16个时间组）
	Reserved        uint32 // 保留字段
	Status          uint8  // 状态（最低位1有效0无效）
}
