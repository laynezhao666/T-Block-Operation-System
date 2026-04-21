// Package cacs 定义CACS门禁协议的请求和响应数据结构。
package cacs

import (
	"dac/logic/collect/driver/cacs/consts"
	"reflect"
)

// GetFieldSizeSum 计算结构体所有字段的实际大小总和（不含内存对齐填充）
func GetFieldSizeSum(s interface{}) int {
	return getFieldSizeSumRecursive(reflect.ValueOf(s))
}

// getFieldSizeSumRecursive 递归计算结构体字段的实际大小（不含内存对齐填充）
func getFieldSizeSumRecursive(v reflect.Value) int {
	size := 0

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		switch field.Kind() {
		case reflect.Struct:
			// 递归处理嵌套结构体
			size += getFieldSizeSumRecursive(field)
		case reflect.Array:
			// 数组：元素大小 × 数组长度
			elemKind := fieldType.Type.Elem().Kind()
			if elemKind == reflect.Struct {
				// 结构体数组，需要递归计算每个元素
				for j := 0; j < field.Len(); j++ {
					size += getFieldSizeSumRecursive(field.Index(j))
				}
			} else {
				// 基本类型数组
				elemSize := int(fieldType.Type.Elem().Size())
				size += elemSize * field.Len()
			}
		default:
			// 基本类型直接使用大小
			size += int(fieldType.Type.Size())
		}
	}

	return size
}

// ControllerRegisterReq 控制器注册请求
type ControllerRegisterReq struct {
	Id            uint32
	Version       uint32
	Seq           uint32
	EventAlarmSeq uint32
	MAC           [consts.KMACAddrLen]byte
	Name          [consts.KControllerNameLen]byte
	Mode          uint8
}

// ControllerRegisterResp 控制器注册响应
type ControllerRegisterResp struct {
	Id uint32
}

// DoorControlReq 门控制请求
type DoorControlReq struct {
	ControlMode uint8
	Id          uint32
}

// DoorControlResp 门控制响应
type DoorControlResp struct {
}

// DoorStateReq 门状态查询请求
type DoorStateReq struct {
	Id uint32
}

// DoorStateResp 门状态查询响应
type DoorStateResp struct {
	Id                 uint32
	AuxiliaryDI        uint8
	ReservedDI1        uint8
	ReservedDI2        uint8
	AlarmDO            uint8
	ReservedDO1        uint8
	ReservedDO2        uint8
	DoorSensorStatus   uint8
	ButtonStatus       uint8
	ElectricLockStatus uint8
	OpenDoorStatus     uint8
	DoorMode           uint8
}

// DownloadControllerParamsReq 下载控制器参数请求
type DownloadControllerParamsReq struct {
	Mode uint8
	Name string
}

// DownloadControllerParamsResp 下载控制器参数响应
type DownloadControllerParamsResp struct {
}

// GetControllerParamsReq 获取控制器参数请求
type GetControllerParamsReq struct {
}

// GetControllerParamsResp 获取控制器参数响应
type GetControllerParamsResp struct {
	Mode uint8
	Name [20]byte
}

// DownloadDoorParamsReq 下载门参数请求
type DownloadDoorParamsReq struct {
	Id                         uint32
	AuxiliaryDI                uint8
	ReservedDI1                uint8
	ReservedDI2                uint8
	AlarmDO                    uint8
	ReservedDO1                uint8
	ReservedDO2                uint8
	DoorSensorStatus           uint8
	ElectricLockStatus         uint8
	ButtonStatus               uint8
	DoorMode                   uint8
	MultiCardsOpenDoorNum      uint8
	AntiPassback               uint8
	EntryAreaNum               uint16
	ExitAreaNum                uint16
	OpenDoorKeepTime           uint32
	OpenDoorTimeoutTime        uint32
	AlarmTime                  uint32
	CoercivePassword           uint32
	Password                   uint32
	InterLockId1               uint32
	InterLockId2               uint32
	InterLockId3               uint32
	MultiCardsOpenDoorInterval uint32
	CardPasswordInterval       uint32
	LockType                   uint32
}

// DownloadDoorParamsResp 下载门参数响应
type DownloadDoorParamsResp struct{}

// GetDoorParamsReq 获取门参数请求
type GetDoorParamsReq struct {
	Id uint32
}

// GetDoorParamsResp 获取门参数响应
type GetDoorParamsResp struct {
	Id                         uint32
	AuxiliaryDI                uint8
	ReservedDI1                uint8
	ReservedDI2                uint8
	AlarmDO                    uint8
	ReservedDO1                uint8
	ReservedDO2                uint8
	DoorSensorStatus           uint8
	ElectricLockStatus         uint8
	ButtonStatus               uint8
	DoorMode                   uint8
	MultiCardsOpenDoorNum      uint8
	AntiPassback               uint8
	EntryAreaNum               uint16
	ExitAreaNum                uint16
	OpenDoorKeepTime           uint32
	OpenDoorTimeoutTime        uint32
	AlarmTime                  uint32
	CoercivePassword           uint32
	Password                   uint32
	InterLockId1               uint32
	InterLockId2               uint32
	InterLockId3               uint32
	MultiCardsOpenDoorInterval uint32
	CardPasswordInterval       uint32
	LockType                   uint32 // 协议实际返回数据部分共64字节，比文档多3字节，怀疑是锁类型实际为4字节
}

// DoorAuth 门授权信息
type DoorAuth struct {
	PermitPeriod uint8
	AuthType     [consts.KAuthTypeLen]uint8
}

// CardInfo 卡信息
type CardInfo struct {
	Id          uint32
	UserId      uint32
	Password    uint32
	StartYear   uint16
	StartMonth  uint8
	StartDay    uint8
	StartHour   uint8
	StartMinute uint8
	StartSecond uint8
	Reserved    uint8
	EndYear     uint16
	EndMonth    uint8
	EndDay      uint8
	EndHour     uint8
	EndMinute   uint8
	EndSecond   uint8
	CardType    uint8
	AuthDoor1   DoorAuth
	AuthDoor2   DoorAuth
	AuthDoor3   DoorAuth
	AuthDoor4   DoorAuth
	AreaId      uint16
}

// DownloadCardsReq 下载卡请求
type DownloadCardsReq struct {
	Num   uint8
	Cards []CardInfo
}

// DownloadCardsResp 下载卡响应
type DownloadCardsResp struct {
	SuccessNum uint8
	FailCardId uint32
}

// GetCardsReq 获取卡请求（Type=0按用户编号，Type=1按卡号）
type GetCardsReq struct {
	Type uint8
	Id   uint32
}

// GetCardsResp 获取卡响应
type GetCardsResp struct {
	CardNum uint8
	Card1   CardInfo
	Card2   CardInfo
	Card3   CardInfo
	Card4   CardInfo
}

// DeleteCardsReq 删除卡请求。
// Type=0: 按卡号删除
// Type=1: 按用户编号删除
// Type=2: 删除所有卡
type DeleteCardsReq struct {
	Type uint8
	Id   uint32
}

// DeleteCardsResp 删除卡响应
type DeleteCardsResp struct {
}

// TimeGroup 时间组时段
type TimeGroup struct {
	StartHour   uint8
	StartMinute uint8
	EndHour     uint8
	EndMinute   uint8
}

// AddTimeGroupsReq 添加时间组请求
type AddTimeGroupsReq struct {
	Id         uint8
	WhatDay    uint8
	TimeGroups [consts.KTimeGroupPeriodNum]TimeGroup
}

// AddTimeGroupsResp 添加时间组响应
type AddTimeGroupsResp struct {
}

// GetTimeGroupsReq 获取时间组请求
type GetTimeGroupsReq struct {
	Id      uint8
	WhatDay uint8
}

// GetTimeGroupsResp 获取时间组响应
type GetTimeGroupsResp struct {
	Id         uint8
	WhatDay    uint8
	TimeGroups [consts.KTimeGroupPeriodNum]TimeGroup
}

// DeleteTimeGroupsReq 删除时间组请求。
// Type=0: 删除一天，Id和WhatDay为0-7
// Type=1: 删除一张时间表，Id为0-7，WhatDay为0xFF
// Type=2: 删除所有时间表，Id为0xFF，WhatDay为0xFF
type DeleteTimeGroupsReq struct {
	Type    uint8
	Id      uint8
	WhatDay uint8
}

// DeleteTimeGroupsResp 删除时间组响应
type DeleteTimeGroupsResp struct{}

// UploadDoorStatus 门状态上报数据
type UploadDoorStatus struct {
	Id                 uint32
	AuxiliaryDI        uint8
	ReservedDI1        uint8
	ReservedDI2        uint8
	AlarmDO            uint8
	ReservedDO1        uint8
	ReservedDO2        uint8
	DoorSensorStatus   uint8
	ButtonStatus       uint8
	ElectricLockStatus uint8
	OpenDoorStatus     uint8
	DoorMode           uint8
}

// EventAlarmItem 事件告警条目
type EventAlarmItem struct {
	Type         uint8
	Year         uint16
	Month        uint8
	Day          uint8
	Hour         uint8
	Minute       uint8
	Second       uint8
	DoorId       uint32
	CardId       uint32
	Extras       uint32
	CardReaderId uint8
}

// UploadEventAlarmReq 事件告警上报请求
type UploadEventAlarmReq struct {
	Num   uint16
	Seq   uint32
	Items []EventAlarmItem
}

// UploadEventAlarmResp 事件告警上报响应
type UploadEventAlarmResp struct {
	SuccessNum    uint16
	EventAlarmSeq uint32
}

// UploadControllerStatus 控制器状态上报数据
type UploadControllerStatus struct {
	Year            uint16
	Month           uint8
	Day             uint8
	Hour            uint8
	Minute          uint8
	Second          uint8
	FireAlarmStatus uint8
}

// SetTimeReq 设置时间请求
type SetTimeReq struct {
	Year   uint16
	Month  uint8
	Day    uint8
	Hour   uint8
	Minute uint8
	Second uint8
}

// SetTimeResp 设置时间响应
type SetTimeResp struct {
}

// GetCardsInfoReq 获取卡信息请求（按索引分页）
type GetCardsInfoReq struct {
	Index uint32
}

// GetCardsInfoResp 获取卡信息响应
type GetCardsInfoResp struct {
	NextIndex uint32
	Num       uint32
	Cards     []CardInfo
}
