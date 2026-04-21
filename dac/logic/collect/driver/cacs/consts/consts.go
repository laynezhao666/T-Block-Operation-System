// Package consts 定义 CACS 协议相关的常量、命令码和错误码。
package consts

import (
	"fmt"
	"time"
)

const (
	// KIDDoorState 门状态标识符
	KIDDoorState string = "DoorState"

	// KPassiveReceivePort CACS被动监听端口号
	KPassiveReceivePort = 31235

	// KMethodOpen 开门方式常量
	KMethodOpen uint8 = 0

	// KWaitOpenTimeout 等待开门超时时间（秒）
	KWaitOpenTimeout int32 = 10

	// KHeader CACS协议包头标识
	KHeader uint32 = 0xcc567aee

	// KFirstRecvLen 首次接收数据长度（字节）
	KFirstRecvLen int = 12

	// KRequestPacketFixedLength 请求包固定长度
	KRequestPacketFixedLength int = 14
	// KResponsePacketFixedLength 响应包固定长度
	KResponsePacketFixedLength int = 18

	// --- CACS 协议命令码定义 ---

	// 注册命令
	KCommandRequestRegister  uint32 = 0xfe01 // 注册请求
	KCommandResponseRegister uint32 = 0xfe02 // 注册响应

	// 控制器参数命令
	KCommandRequestDownloadControllerParams  uint32 = 0xfe03 // 下载控制器参数请求
	KCommandResponseDownloadControllerParams uint32 = 0xfe04 // 下载控制器参数响应
	KCommandRequestGetControllerParams       uint32 = 0xfe05 // 获取控制器参数请求
	KCommandResponseGetControllerParams      uint32 = 0xfe06 // 获取控制器参数响应

	// 门参数命令
	KCommandRequestDownloadDoorParams  uint32 = 0xfe07 // 下载门参数请求
	KCommandResponseDownloadDoorParams uint32 = 0xfe08 // 下载门参数响应
	KCommandRequestGetDoorParams       uint32 = 0xfe09 // 获取门参数请求
	KCommandResponseGetDoorParams      uint32 = 0xfe0a // 获取门参数响应

	// 卡管理命令
	KCommandRequestDownloadCards  uint32 = 0xfe0b // 下载卡信息请求
	KCommandResponseDownloadCards uint32 = 0xfe0c // 下载卡信息响应
	KCommandRequestGetCards       uint32 = 0xfe0d // 获取卡信息请求
	KCommandResponseGetCards      uint32 = 0xfe0e // 获取卡信息响应
	KCommandRequestDeleteCards    uint32 = 0xfe0f // 删除卡信息请求
	KCommandResponseDeleteCards   uint32 = 0xfe10 // 删除卡信息响应

	// 时间组命令
	KCommandRequestAddTimeGroups     uint32 = 0xfe11 // 添加时间组请求
	KCommandResponseAddTimeGroups    uint32 = 0xfe12 // 添加时间组响应
	KCommandRequestGetTimeGroups     uint32 = 0xfe13 // 获取时间组请求
	KCommandResponseGetTimeGroups    uint32 = 0xfe14 // 获取时间组响应
	KCommandRequestDeleteTimeGroups  uint32 = 0xfe15 // 删除时间组请求
	KCommandResponseDeleteTimeGroups uint32 = 0xfe16 // 删除时间组响应

	// 上报与控制命令
	KCommandUploadDoorStatus         uint32 = 0xfe1d // 上报门状态
	KCommandRequestUploadEventAlarm  uint32 = 0xfe1e // 上报事件告警请求
	KCommandResponseUploadEventAlarm uint32 = 0xfe1f // 上报事件告警响应
	KCommandRequestRemoteControl     uint32 = 0xfe20 // 远程控制请求
	KCommandResponseRemoteControl    uint32 = 0xfe21 // 远程控制响应
	KCommandRequestSetTime           uint32 = 0xfe22 // 设置时间请求
	KCommandResponseSetTime          uint32 = 0xfe23 // 设置时间响应
	KCommandUploadControllerStatus   uint32 = 0xfe24 // 上报控制器状态

	// 点位上报间隔命令
	KCommandRequestSetPointUploadIntervalPeriod  uint32 = 0xfe25 // 设置上报间隔请求
	KCommandResponseSetPointUploadIntervalPeriod uint32 = 0xfe26 // 设置上报间隔响应
	KCommandRequestGetPointUploadIntervalPeriod  uint32 = 0xfe27 // 获取上报间隔请求
	KCommandResponseGetPointUploadIntervalPeriod uint32 = 0xfe28 // 获取上报间隔响应

	// 门状态查询与卡信息批量读取命令
	KCommandRequestDoorStatus    uint32 = 0xfe37 // 查询门状态请求
	KCommandResponseDoorStatus   uint32 = 0xfe38 // 查询门状态响应
	KCommandRequestGetCardsInfo  uint32 = 0xfe39 // 批量获取卡信息请求
	KCommandResponseGetCardsInfo uint32 = 0xfe3a // 批量获取卡信息响应

	// --- 返回码定义 ---

	KRtnNormal                        uint32 = 0x00 // 正常
	KRtnReadFlashError                uint32 = 0x01 // 读取FLASH出错
	KRtnWriteFlashError               uint32 = 0x02 // 写入FLASH出错
	KRtnFlashOutOfMemoryError         uint32 = 0x03 // FLASH空间已满
	KRtnParamLenError                 uint32 = 0x04 // 参数长度有误
	KRtnInternalServerError           uint32 = 0x05 // 控制器内部操作失败
	KRtnControllerTypeError           uint32 = 0x06 // 控制器类型错误
	KRtnControllerSeqNotMatchError    uint32 = 0x07 // 控制器序列号不匹配
	KRtnDownloadControllerParamsError uint32 = 0x08 // 下载控制器参数错误
	KRtnDownloadDoorParamsError       uint32 = 0x09 // 下载门参数错误
	KRtnGetDoorParamsError            uint32 = 0x0a // 获取门参数错误
	KRtnDownloadCardInfoError         uint32 = 0x0b // 下载卡信息错误
	KRtnUserCardReachLimit            uint32 = 0x0c // 用户卡号已满
	KRtnCardIsAllocatedError          uint32 = 0x0d // 卡已分配给其他用户
	KRtnGetTimeGroupParamsError       uint32 = 0x13 // 读取时间组参数错误
	KRtnTimeGroupNotFound             uint32 = 0x14 // 时间组不存在

	// --- 协议字段长度常量 ---

	KMacLen int = 6 // MAC地址长度

	KLogInterval int = 10 // 日志打印间隔

	// 数据包偏移量常量
	KOffsetDataInResponsePacket int = 4  // 响应包数据偏移
	KOffsetControllerSeq        int = 8  // 控制器序列号偏移
	KOffsetControllerMacAddress int = 16 // 控制器MAC地址偏移
	KOffsetControllerName       int = 22 // 控制器名称偏移
	KOffsetDoorNumber           int = 0  // 门编号偏移
	KOffsetDoorSensorInput      int = 10 // 门磁输入偏移
	KOffsetDoorPowerOutput      int = 12 // 门电源输出偏移

	// 超时与字段长度常量
	KDefaultTimeoutMs       = 5 * time.Second // 默认超时时间
	KControllerNameLen  int = 20              // 控制器名称长度
	KMACAddrLen             = 6               // MAC地址字节长度
	KAuthTypeLen            = 2               // 授权类型长度
	KTimeGroupPeriodNum     = 6               // 时间组时段数量
	KTimeGroupNum           = 8               // 时间组数量
	KRegisterReqLen     int = 43              // 注册请求长度

	// RRPC相关常量
	KRRPCPrefix = "rrpc" // RRPC键前缀
	KExtendConn = "conn" // 扩展连接标识

	// 门数量限制
	KSupportedDoorNum = 2 // 支持的门数量

	// --- 业务错误码 ---

	KNormal         = 0  // 正常
	KMarshalError   = -1 // 序列化错误
	KRequestError   = -2 // 请求发送错误
	KRecvRespError  = -3 // 接收响应错误
	KUnMarshalError = -4 // 反序列化错误

	DefaultLimit = 100 // 数据库查询默认分页大小

	// --- 事件类型常量 ---

	KEventSwipingCard            uint8 = 0 // 刷卡事件
	KEventInputPassword          uint8 = 1 // 密码输入事件
	KEventManualButtonExitRecord uint8 = 2 // 手动按钮出门记录
	KEventDoorOpen               uint8 = 3 // 门磁打开事件
	KEventDoorClose              uint8 = 4 // 门磁合上事件
	KEventDoorInterLock          uint8 = 5 // 门互锁事件

	// --- 告警类型常量 ---

	KAlarmDoorOpenTimeout              uint8 = 128 // 门开超时告警
	KAlarmViolentInvasion              uint8 = 129 // 暴力入侵告警
	KAlarmCoerceCard                   uint8 = 130 // 胁迫卡告警
	KAlarmCoercePassword               uint8 = 131 // 胁迫密码告警
	KAlarmAuxiliaryDI                  uint8 = 132 // 辅助DI告警
	KAlarmAntiPassback                 uint8 = 135 // 反潜回告警
	KAlarmFire                         uint8 = 136 // 火警
	KAlarmFingerPringControllerOffline uint8 = 137 // 指纹控制器掉线

	KAlarmOn  uint32 = 1 // 当前告警产生
	KAlarmOff uint32 = 0 // 当前告警取消
)

// RtnInfoMap 返回码到中文描述的映射表
var (
	RtnInfoMap = map[uint32]string{
		KRtnNormal:                  "正常",
		KRtnReadFlashError:          "读取FLASH出错",
		KRtnWriteFlashError:         "写入FLASK出错",
		KRtnFlashOutOfMemoryError:   "FLASH空间已满",
		KRtnParamLenError:           "参数长度有误",
		KRtnInternalServerError:     "控制器内部操作失败",
		KRtnDownloadDoorParamsError: "门参数下载数据值有误",
		KRtnDownloadCardInfoError:   "卡号信息下载数据值有误或卡号已存在",
		KRtnUserCardReachLimit:      "用户卡号已满",
		KRtnCardIsAllocatedError:    "该卡已经分配给其他用户",
		KRtnGetTimeGroupParamsError: "读取星期时间表数据值有误",
		KRtnTimeGroupNotFound:       "读取的星期时间表不存在",
	}

	// RequestInfoMap 请求错误码到中文描述的映射表
	RequestInfoMap = map[int]string{
		KMarshalError:   "请求参数Marshal失败",
		KRequestError:   "请求发送失败",
		KRecvRespError:  "接收响应失败",
		KUnMarshalError: "响应结果UnMarshal失败",
	}

	// EventAlarmInfoMap 事件告警类型到中文描述的映射表
	EventAlarmInfoMap = map[uint8]string{
		KEventSwipingCard:                  "刷卡",
		KEventInputPassword:                "密码输入",
		KEventManualButtonExitRecord:       "手动按钮出门记录",
		KEventDoorOpen:                     "门磁打开",
		KEventDoorClose:                    "门磁合上",
		KEventDoorInterLock:                "门互锁",
		KAlarmDoorOpenTimeout:              "门开超时",
		KAlarmViolentInvasion:              "暴力入侵",
		KAlarmCoerceCard:                   "胁迫卡告警",
		KAlarmCoercePassword:               "胁迫密码告警",
		KAlarmAuxiliaryDI:                  "辅助DI告警",
		KAlarmAntiPassback:                 "反潜回告警",
		KAlarmFire:                         "火警",
		KAlarmFingerPringControllerOffline: "指纹控制器掉线",
	}
)

// GetRRPCKey 生成带 channelID 的 RRPC Key，确保不同门控器的响应不会混淆。
func GetRRPCKey(channelID string, command uint32) string {
	return fmt.Sprintf("%s-%s-%x", KRRPCPrefix, channelID, command)
}

// GetRRPCDoorStatus 获取门状态查询的RRPC Key。
func GetRRPCDoorStatus(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDoorStatus)
}

// GetRRPCRemoteControl 获取远程控制的RRPC Key。
func GetRRPCRemoteControl(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestRemoteControl)
}

// GetRRPCDownloadControllerParams 获取下载控制器参数的RRPC Key。
func GetRRPCDownloadControllerParams(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDownloadControllerParams)
}

// GetRRPCGetControllerParams 获取读取控制器参数的RRPC Key。
func GetRRPCGetControllerParams(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestGetControllerParams)
}

// GetRRPCDownloadDoorParams 获取下载门参数的RRPC Key。
func GetRRPCDownloadDoorParams(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDownloadDoorParams)
}

// GetRRPCGetDoorParams 获取读取门参数的RRPC Key。
func GetRRPCGetDoorParams(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestGetDoorParams)
}

// GetRRPCDownloadCards 获取下载卡信息的RRPC Key。
func GetRRPCDownloadCards(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDownloadCards)
}

// GetRRPCGetCards 获取读取卡信息的RRPC Key。
func GetRRPCGetCards(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestGetCards)
}

// GetRRPCDeleteCards 获取删除卡信息的RRPC Key。
func GetRRPCDeleteCards(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDeleteCards)
}

// GetRRPCAddTimeGroups 获取添加时间组的RRPC Key。
func GetRRPCAddTimeGroups(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestAddTimeGroups)
}

// GetRRPCGetTimeGroups 获取读取时间组的RRPC Key。
func GetRRPCGetTimeGroups(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestGetTimeGroups)
}

// GetRRPCDeleteTimeGroups 获取删除时间组的RRPC Key。
func GetRRPCDeleteTimeGroups(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestDeleteTimeGroups)
}

// GetRRPCSetTime 获取设置时间的RRPC Key。
func GetRRPCSetTime(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestSetTime)
}

// GetRRPCGetCardsInfo 获取批量读取卡信息的RRPC Key。
func GetRRPCGetCardsInfo(channelID string) string {
	return GetRRPCKey(channelID, KCommandRequestGetCardsInfo)
}
