// Package consts 定义XBrother门禁协议的常量、命令码和RRPC地址。
package consts

import (
	"fmt"
	"time"
)

// 协议控制字符常量
const (
	STX uint8 = 0x02 // 起始符
	ETX uint8 = 0x03 // 结束符
	ACK uint8 = 0x06 // 确认
	DLE uint8 = 0x10 // 转义
	NAK uint8 = 0x15 // 否认
	SYN uint8 = 0x16 // 同步

	FirstRecvLen = 7 // 首次接收长度
	CSAndETXLen  = 2 // 校验和+结束符长度

	// 命令码常量
	CommandSetControllerParams    uint8 = 0x63 // 设置控制器参数
	CommandOpenDoor               uint8 = 0x2c // 开门
	CommandDoorOpenPermenently    uint8 = 0x2d // 常开门
	CommandCloseDoor              uint8 = 0x2e // 关门
	CommandSetTime                uint8 = 0x07 // 设置时间
	CommandSetDoorParams          uint8 = 0x61 // 设置门参数
	CommandClearTimeGroups        uint8 = 0x0f // 清除时间组
	CommandAddTimeGroup           uint8 = 0x0d // 添加时间组
	CommandDeleteCard             uint8 = 0x16 // 删除卡
	CommandClearCards             uint8 = 0x17 // 清除所有卡
	CommandClean                  uint8 = 0x04 // 清除数据
	CommandSetAlarm               uint8 = 0x18 // 设置告警
	CommandSetFireAlarm           uint8 = 0x19 // 设置火警
	CommandGetEvents              uint8 = 0x31 // 获取事件
	CommandGetAlarms              uint8 = 0x3a // 获取告警
	CommandAddCard                uint8 = 0x62 // 添加卡
	CommandAddCards               uint8 = 0x88 // 批量添加卡
	CommandGetStatus              uint8 = 0x40 // 获取状态
	CommandUploadEvents           uint8 = 0x53 // 上报事件
	CommandUploadAlarms           uint8 = 0x54 // 上报告警
	CommandUploadStatus           uint8 = 0x55 // 上报状态
	CommandUploadControllerStatus uint8 = 0x56 // 上报控制器状态
	CommandHeartBeat2             uint8 = 0x57 // 心跳
	CommandLockDoor               uint8 = 0x2f // 锁门

	RRPCPrefix = "dac-rrpc" // RRPC地址前缀

	PasswordLen        = 4    // 密码长度
	PasswordLowerBound = 1000 // 密码最小值

	// 卡状态常量
	ControllerCardStatusEnable  uint8 = 1 // 启用
	ControllerCardStatusDisable uint8 = 0 // 禁用

	// 告警参数常量
	DefaultAlarmTime     uint16 = 1             // 默认告警时间
	DoorAlarm            uint8  = uint8(1)      // 门告警
	OpenDoorTimeoutAlarm uint8  = uint8(1) << 1 // 开门超时告警
	InvalidCardAlarm     uint8  = uint8(1) << 2 // 无效卡告警
	InvalidTimeAlarm     uint8  = uint8(1) << 3 // 无效时间告警

	// 时间组截止日期默认值
	TimeGroupDeadlineYear  uint8 = 99 // 实际代表2099年
	TimeGroupDeadlineMonth uint8 = 1  // 截止月
	TimeGroupDeadlineDay   uint8 = 1  // 截止日

	// 星期位掩码
	WeekSunday    uint8 = 0x01 // 周日
	WeekMonday    uint8 = 0x02 // 周一
	WeekTuesday   uint8 = 0x04 // 周二
	WeekWednesday uint8 = 0x08 // 周三
	WeekThursday  uint8 = 0x10 // 周四
	WeekFriday    uint8 = 0x20 // 周五
	WeekSaturday  uint8 = 0x40 // 周六

	// 开门类型
	OpenDoorTypeOnlyCard uint8 = 0x01 // 仅刷卡
	OpenDoorTypeCardPass uint8 = 0x02 // 卡加密码
	OpenDoorTypePass     uint8 = 0x08 // 仅密码

	DefaultLimit = 10 // 默认分页大小

	// 锁门状态
	LockDoorStateLock   uint8 = 1 // 锁定
	LockDoorStateUnlock uint8 = 0 // 解锁

	// Redis相关常量
	RedisKeyXBrotherPredix string = "dac.xbrother"   // Redis Key前缀
	RedisNoTimeout                = 0                // 无超时
	RedisDefaultTimeout           = 60 * time.Second // 默认超时
	RedisRetryTimes        int    = 3                // 重试次数

	// 卡索引和方向常量
	CardIndexMax     uint32 = 30000 // 最大卡索引
	DirectionEnter          = 0     // 进入方向
	DirectionExit           = 1     // 离开方向
	DirectionUnKnown        = 2     // 未知方向

	ChanInitLength = 5 // 通道初始长度
)

// 事件和告警类型常量
const (
	TypeEffectiveCard   uint8 = 0 // 有效卡
	TypeInvalidCard     uint8 = 1 // 无效卡号
	TypeTimezoneError   uint8 = 2 // 无时区
	TypeInvalidTimezone uint8 = 3 // 无效时区
	TypePINError        uint8 = 4 // 密码错误

	TypeExitAccess            uint8 = 10 // 出门
	TypeEntryAccess           uint8 = 11 // 进门
	TypeCardAndPasswordAccess uint8 = 12 // 卡加密码开门
	TypeOpenByDoubleCards     uint8 = 13 // 双卡开门
	TypeOpenByPassword        uint8 = 14 // 密码开门
	TypeFreePass              uint8 = 15 // 无限制通行

	TypeEntryByCardAndPIN uint8 = 20 // 卡加密码进门
	TypeExitByCardAndPIN  uint8 = 21 // 卡加密码出门
	TypeEntryByPIN        uint8 = 24 // 密码进门
	TypeExitByPIN         uint8 = 25 // 密码出门
	TypeEntryByFree       uint8 = 26 // 自由进门
	TypeExitByFree        uint8 = 27 // 自由出门
	TypeRejectEntry       uint8 = 28 // 限制进门
	TypeRejectExit        uint8 = 29 // 限制出门

	TypeReaderAlarm     uint8 = 39 // 读卡器防拆报警
	TypeAlarmA          uint8 = 40 // 火警
	TypeAlarmB          uint8 = 41 // 报警输出
	TypeDoorAlarm       uint8 = 56 // 非法开门报警
	TypeDoorOpenTooLong uint8 = 57 // 门开太久报警
	TypeDoorUnlock      uint8 = 58 // 门没锁报警

	TypeFireAlarm       uint8 = 72 // 触发火警输入
	TypeDoorInterUnlock uint8 = 73 // 被互锁禁止开门
	TypeControllerStart uint8 = 75 // 控制器启动
	TypeInvalidDoor     uint8 = 76 // 无效门

	// 驱动层卡状态
	DriverCardStatusEnable  int = 0 // 启用
	DriverCardStatusDisable int = 1 // 禁用
	DriverCardStatusDelete  int = 2 // 已删除

	// 时间组数量限制
	DaysOneWeek             = 7  // 每周天数
	OneDoorMaxTimeGroupNum  = 16 // 单门最大时间组数
	TwoDoorMaxTimeGroupNum  = 16 // 双门最大时间组数
	FourDoorMaxTimeGroupNum = 8  // 四门最大时间组数

	// 控制器门状态
	ControllerDoorStatusOpen  = 0 // 控制器侧开门
	ControllerDoorStatusClose = 1 // 控制器侧关门
	DriverDoorStatusOpen      = 1 // 驱动侧开门
	DriverDoorStatusClose     = 0 // 驱动侧关门

	// 门参数默认值
	DefaultDoorParamLongTimeOpenAlarm   uint8 = 1 // 长时间开门告警
	DefaultDoorParamBidirectionalDetect uint8 = 1 // 双向检测

	DefaultStartCardIndex = 0                     // 起始卡索引
	DurationSleepTime     = 10 * time.Millisecond // 命令间隔
	ConnCheckTime         = 30 * time.Second      // 连接检查间隔
	DLMAndReOpenCheckTime = 5 * time.Second       // DLM重连检查间隔
)

// RRPC地址变量（按命令码生成）
var (
	rrpcSetControllerParams = fmt.Sprintf("%s-%x", RRPCPrefix, CommandSetControllerParams)
	rrpcOpenDoor            = fmt.Sprintf("%s-%x", RRPCPrefix, CommandOpenDoor)
	rrpcDoorOpenPermanently = fmt.Sprintf("%s-%x", RRPCPrefix, CommandDoorOpenPermenently)
	rrpcCloseDoor           = fmt.Sprintf("%s-%x", RRPCPrefix, CommandCloseDoor)
	rrpcLockDoor            = fmt.Sprintf("%s-%x", RRPCPrefix, CommandLockDoor)
	rrpcSetTime             = fmt.Sprintf("%s-%x", RRPCPrefix, CommandSetTime)
	rrpcSetDoorParams       = fmt.Sprintf("%s-%x", RRPCPrefix, CommandSetDoorParams)
	rrpcClearDoorTimeGroups = fmt.Sprintf("%s-%x", RRPCPrefix, CommandClearTimeGroups)
	rrpcAddTimeGroup        = fmt.Sprintf("%s-%x", RRPCPrefix, CommandAddTimeGroup)
	rrpcClearCards          = fmt.Sprintf("%s-%x", RRPCPrefix, CommandClearCards)
	rrpcDeleteCard          = fmt.Sprintf("%s-%x", RRPCPrefix, CommandDeleteCard)
	rrpcAddCard             = fmt.Sprintf("%s-%x", RRPCPrefix, CommandAddCard)
	rrpcClean               = fmt.Sprintf("%s-%x", RRPCPrefix, CommandClean)
	rrpcSetAlarm            = fmt.Sprintf("%s-%x", RRPCPrefix, CommandSetAlarm)
	rrpcSetFireAlarm        = fmt.Sprintf("%s-%x", RRPCPrefix, CommandSetFireAlarm)
	rrpcSetDriverEvent      = fmt.Sprintf("%s-%s", RRPCPrefix, "SetDriverEvent")
	rrpcSetDriverAlarm      = fmt.Sprintf("%s-%s", RRPCPrefix, "SetDriverAlarm")
	rrpcSetDoorStatus       = fmt.Sprintf("%s-%s", RRPCPrefix, "SetDoorStatus")

	EventAlarmDescMap = map[uint8]string{
		TypeEffectiveCard:         "单卡识别",
		TypeInvalidCard:           "无效卡号",
		TypeTimezoneError:         "没有时区，即没有开放时间",
		TypeInvalidTimezone:       "无效时区，即该时间没有权限",
		TypePINError:              "密码错误",
		TypeExitAccess:            "出门",
		TypeEntryAccess:           "进门",
		TypeCardAndPasswordAccess: "卡加密码开门",
		TypeOpenByDoubleCards:     "双卡开门",
		TypeOpenByPassword:        "密码开门",
		TypeFreePass:              "无限制通行",
		TypeEntryByCardAndPIN:     "卡加密码进门",
		TypeExitByCardAndPIN:      "卡加密码出门",
		TypeEntryByPIN:            "密码进门",
		TypeExitByPIN:             "密码出门",
		TypeEntryByFree:           "自由进门",
		TypeExitByFree:            "自由出门",
		TypeRejectEntry:           "限制进门",
		TypeRejectExit:            "限制出门",
		TypeReaderAlarm:           "读卡器防拆报警",
		TypeAlarmA:                "火警",
		TypeAlarmB:                "报警输出",
		TypeDoorAlarm:             "非法开门报警",
		TypeDoorOpenTooLong:       "门开太久报警",
		TypeDoorUnlock:            "门没锁报警",
		TypeFireAlarm:             "触发火警输入",
		TypeDoorInterUnlock:       "被互锁禁止开门",
		TypeControllerStart:       "控制器启动",
		TypeInvalidDoor:           "无效门",
	}

	UnknownEventAlarmDesc = "未知"
)

// GetRRPCSetControllerParams 获取设置控制器参数的RRPC地址
func GetRRPCSetControllerParams(channelID string) string {
	return channelID + "-" + rrpcSetControllerParams
}

// GetRRPCOpenDoor 获取开门的RRPC地址
func GetRRPCOpenDoor(channelID string) string { return channelID + "-" + rrpcOpenDoor }

// GetRRPCDoorOpenPermanently 获取常开门的RRPC地址
func GetRRPCDoorOpenPermanently(channelID string) string {
	return channelID + "-" + rrpcDoorOpenPermanently
}

// GetRRPCCloseDoor 获取关门的RRPC地址
func GetRRPCCloseDoor(channelID string) string { return channelID + "-" + rrpcCloseDoor }

// GetRRPCLockDoor 获取锁门的RRPC地址
func GetRRPCLockDoor(channelID string) string { return channelID + "-" + rrpcLockDoor }

// GetRRPCSetTime 获取设置时间的RRPC地址
func GetRRPCSetTime(channelID string) string { return channelID + "-" + rrpcSetTime }

// GetRRPCSetDoorParams 获取设置门参数的RRPC地址
func GetRRPCSetDoorParams(channelID string) string { return channelID + "-" + rrpcSetDoorParams }

// GetRRPCClearDoorTimeGroups 获取清除门时间组的RRPC地址
func GetRRPCClearDoorTimeGroups(channelID string) string {
	return channelID + "-" + rrpcClearDoorTimeGroups
}

// GetRRPCAddTimeGroup 获取添加时间组的RRPC地址
func GetRRPCAddTimeGroup(channelID string) string { return channelID + "-" + rrpcAddTimeGroup }

// GetRRPCClearCards 获取清除所有卡的RRPC地址
func GetRRPCClearCards(channelID string) string { return channelID + "-" + rrpcClearCards }

// GetRRPCDeleteCard 获取删除卡的RRPC地址
func GetRRPCDeleteCard(channelID string) string { return channelID + "-" + rrpcDeleteCard }

// GetRRPCAddCard 获取添加卡的RRPC地址
func GetRRPCAddCard(channelID string) string { return channelID + "-" + rrpcAddCard }

// GetRRPCClean 获取清除数据的RRPC地址
func GetRRPCClean(channelID string) string { return channelID + "-" + rrpcClean }

// GetRRPCSetAlarm 获取设置告警的RRPC地址
func GetRRPCSetAlarm(channelID string) string { return channelID + "-" + rrpcSetAlarm }

// GetRRPCSetFireAlarm 获取设置火警的RRPC地址
func GetRRPCSetFireAlarm(channelID string) string { return channelID + "-" + rrpcSetFireAlarm }

// GetRRPCSetDriverEvent 获取设置驱动事件的RRPC地址
func GetRRPCSetDriverEvent(channelID string) string { return channelID + "-" + rrpcSetDriverEvent }

// GetRRPCSetDriverAlarm 获取设置驱动告警的RRPC地址
func GetRRPCSetDriverAlarm(channelID string) string { return channelID + "-" + rrpcSetDriverAlarm }

// GetRRPCSetDoorStatus 获取设置门状态的RRPC地址
func GetRRPCSetDoorStatus(channelID string) string { return channelID + "-" + rrpcSetDoorStatus }
