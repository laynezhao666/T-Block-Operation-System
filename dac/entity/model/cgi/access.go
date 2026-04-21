package cgi

import (
	"dac/entity/model/db"
	"dac/entity/model/driver"
)

// DeleteStaffModeType 删除员工模式类型
type DeleteStaffModeType int

// 删除员工模式常量
const (
	DeleteStaffOnly           = DeleteStaffModeType(0) // 仅删除员工
	DeleteStaffAndDisableCard = DeleteStaffModeType(1) // 删除员工并禁用卡
	DeleteStaffAndDeleteCard  = DeleteStaffModeType(2) // 删除员工并删除卡
)

// StaffAssignedType 员工分配状态类型
type StaffAssignedType int

// 员工分配状态常量
const (
	StaffUnassigned = StaffAssignedType(0) // 未分配
	StaffAssigned   = StaffAssignedType(1) // 已分配
)

// TimeGroup 时间组信息，包含驱动层时间组和组名称
type TimeGroup struct {
	driver.TimeGroup
	GroupName string `json:"group_name"`
}

// Staffs 员工列表响应
type Staffs struct {
	Total int64          `json:"total"`
	List  []StaffAndCard `json:"list"`
}

// StaffAndCard 员工及其关联卡号信息
type StaffAndCard struct {
	db.Staff
	CardNo []string `json:"card_no"`
}

// AccessGroup 权限组及其关联门信息
type AccessGroup struct {
	db.AccessGroupBaseInfo
	Doors []db.DoorBaseInfo `json:"doors"`
}

// Staff 员工信息，包含分配状态
type Staff struct {
	Enable StaffAssignedType `json:"enable"`
	db.Staff
}

// Card 卡信息，包含关联的员工和权限组
type Card struct {
	db.Card
	Staff        Staff         `json:"staff"`
	AccessGroups []AccessGroup `json:"access_groups"`
}

// Cards 卡列表响应
type Cards struct {
	Total int64 `json:"total"`
}

// AccessGroupAndRelationInfos 权限组及关联信息列表响应
type AccessGroupAndRelationInfos struct {
	Total int64                        `json:"total"`
	List  []AccessGroupAndRelationInfo `json:"list"`
}

// AccessGroupAndRelationInfo 权限组详细信息，包含门、时间组和卡
type AccessGroupAndRelationInfo struct {
	db.AccessGroup
	Door      []db.DoorBaseInfo `json:"doors"`
	TimeGroup struct {
		GroupNo   int    `json:"group_no"`
		GroupName string `json:"group_name"`
	} `json:"time_group"`
	Card []db.CardAndStaffBase `json:"cards"`
}

// IDAndName 通用的ID和名称结构体
type IDAndName struct {
	ID   db.IDType `json:"id"`
	Name string    `json:"name"`
}
