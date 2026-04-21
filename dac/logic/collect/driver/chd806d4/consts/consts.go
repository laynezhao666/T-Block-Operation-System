// Package consts 定义CHD806D4门禁协议的常量和RRPC键生成函数。
package consts

import "fmt"
import "time"

// CHD 协议常量定义

// ============ 数据包标识符 ============
const (
	SOI = 0x7E // 起始符 '~'
	EOI = 0x0D // 结束符 '\r'
)

// ============ 数据包类型 CID2/RTN/UP-REP ============
const (
	// CID2 - 命令数据包类型
	CID2AccessAuth   = 0x48 // 访问权限确认命令
	CID2SetParameter = 0x49 // 设置参数命令
	CID2ReadInfo     = 0x4A // 读取信息命令

	// UP-REP - 主动上报类型（0x80 + 重复次数）
	UPRepTrigger  = 0x80 // 触发包（事件上报请求）
	UPRepHeartBit = 0x80 // 心跳包（重复次数=0）
)

// ============ 返回码 RTN ============
const (
	RTNSuccess              = 0x00 // 正常执行
	RTNReserved             = 0x01 // 保留
	RTNChecksumError        = 0x02 // 累加和检查错误
	RTNLengthError          = 0x03 // 参数长度检查错误
	RTNInvalidCID           = 0x04 // 无效的数据包类型
	RTNInvalidCommand       = 0x05 // 不能识别的命令格式
	RTNInvalidData          = 0x06 // 无效的数据
	RTNNoPermission         = 0x07 // 无访问权限
	RTNPasswordError        = 0xE0 // 密码确认不正确
	RTNPasswordModifyFailed = 0xE1 // 密码修改不成功
	RTNStorageFull          = 0xE2 // 存储空间已满
	RTNModifyFailed         = 0xE3 // 参数修改失败
	RTNStorageEmpty         = 0xE4 // 存储空间已空
	RTNNoInfo               = 0xE5 // 无相应信息项
	RTNUserIDDuplicate      = 0xE6 // 用户ID重复
	RTNCardNoDuplicate      = 0xE7 // 卡号重复
	RTNUserInfoDuplicate    = 0xE8 // 用户信息全部重复
)

// ============ 命令组 COMMAND GROUP ============
const (
	GroupAuth = 0x01 // 权限认证组
	GroupSet  = 0x02 // 设置操作组
	GroupRead = 0x03 // 读取信息组
)

// ============ 命令类型 COMMAND TYPE ============
const (
	// 权限认证组
	TypeAuthVerify = 0x80 // 密码校验
	TypeAuthCancel = 0x81 // 取消权限
	TypeAuthModify = 0x82 // 修改密码

	// 用户卡操作组 (协议4.3)
	TypeCardAuthUser  = 0x83 // 授权一个用户卡（至门控制器）(4.3.1)
	TypeCardCancel    = 0x84 // 取消用户卡权限 (4.3.2)
	TypeCardDeleteAll = 0x85 // 全部删除用户（从门控制器内）(4.3.3) - 假设是0x85

	// 读取用户信息 (协议5.4)
	TypeCardGetCount    = 0x8B // 读取已授权的用户数量 (5.4.1)
	TypeCardGetByPos    = 0x8C // 按存储位置读取用户信息 (5.4.2)
	TypeCardGetByUserID = 0x8D // 按用户ID查询 (5.4.3)
	TypeCardGetByCardNo = 0x8E // 按卡号查询 (5.4.4)

	// 门状态操作组
	TypeDoorStateGet            = 0x8F // 获取门状态（远程监控）
	TypeDoorRemoteOpen          = 0x8A // 远程开门（不带操作员信息）
	TypeDoorRemoteOpenWithOp    = 0x8B // 远程开门（带系统操作员信息）
	TypeDoorNormallyCloseWithOp = 0x90 // 远程常闭门与解除（带系统操作员信息）
	TypeDoorNormallyOpenWithOp  = 0x91 // 远程常开门与解除（带系统操作员信息）

	// 时间操作组
	TypeReadTime = 0x80 // 读取实时钟（CID：0X40，CommandGroup：0X03）
	TypeTimeSync = 0x80 // 日期时间同步（CID：0X49，CommandGroup：0X02）

	// 时间组操作组
	TypeTimeGroupSet        = 0x81 // 设置时段表（门号+表序号+16字节时段数据）
	TypeTimeGroupWeekSet    = 0x82 // 设置星期时段索引（门号+0+56字节索引）
	TypeTimeGroupHolidaySet = 0x82 // 设置特殊日期时段索引（门号+1+月日+8字节索引）
	TypeTimeGroupHolidayDel = 0x82 // 删除特殊日期时段索引（门号+2+月日）
	TypeTimeGroupHolidayClr = 0x82 // 清空全部特殊日期（门号+3+00 00）

	TypeTimeGroupGet          = 0x89 // 读取时段表（门号+表序号）
	TypeTimeGroupWeekGet      = 0x8A // 读取星期时段索引（门号+0+星期）
	TypeTimeGroupHolidayGet   = 0x8A // 按序列读取特殊日期时段索引（门号+序号+0）
	TypeTimeGroupHolidayByDay = 0x97 // 按日期读取特殊日期时段表（门号+月+日）

	// 门参数操作
	TypeDoorParamGet = 0x90 // 读取门工作参数
	TypeDoorParamSet = 0x86 // 设置门工作参数

	// 事件读取操作组
	TypeRecordGetParam      = 0x81 // 读取记录参数（SAVEP, LOADP, MAXLEN）
	TypeRecordGetSeq        = 0x82 // 顺序读取一条历史记录
	TypeRecordGetSeqWithPos = 0x83 // 顺序读取一条历史记录（带存储位置）
	TypeRecordGetRandom     = 0x84 // 随机读取记录（指定位置）
	TypeRecordGetLatest     = 0x85 // 查询最新事件记录
	TypeRecordStartAck      = 0x86 // 启动"应答读取"模式
	TypeRecordAckNext       = 0x87 // 应答并读取下一条
	TypeRecordStopAck       = 0x88 // 停止"应答读取"工作模式

)

// ============ RRPC Key 定义 (用于请求-响应匹配)============
//
// 统一格式：chd806d4:{channelID}:{cmdType}:{seqNo}
//
// 说明：
// 1. 所有命令都使用相同的格式，包括权限认证
// 2. cmdType 用于区分不同的命令类型（如 auth:verify, card:add 等）
// 3. seqNo 用于区分同一命令的不同请求实例
// 4. 响应包中只有 SeqNo，因此需要通过 requestMap 存储 SeqNo -> RRPC Key 的映射

var (
	UnknownEventAlarmDesc = "未知"
)

// GetRRPCAuthVerify 密码校验
func GetRRPCAuthVerify(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:auth:verify:%d", channelID, seqNo)
}

// GetRRPCAuthCancel 取消权限
func GetRRPCAuthCancel(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:auth:cancel:%d", channelID, seqNo)
}

// GetRRPCAuthModify 修改密码
func GetRRPCAuthModify(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:auth:modify:%d", channelID, seqNo)
}

// GetRRPCCardAuthUser 授权用户卡
func GetRRPCCardAuthUser(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:authuser:%d", channelID, seqNo)
}

// GetRRPCCardAdd 添加卡
func GetRRPCCardAdd(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:add:%d", channelID, seqNo)
}

// GetRRPCCardDelete 删除卡
func GetRRPCCardDelete(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:delete:%d", channelID, seqNo)
}

// GetRRPCCardUpdate 修改卡
func GetRRPCCardUpdate(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:update:%d", channelID, seqNo)
}

// GetRRPCCardQuery 查询卡
func GetRRPCCardQuery(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:query:%d", channelID, seqNo)
}

// GetRRPCCardGetAll 获取所有卡
func GetRRPCCardGetAll(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:getall:%d", channelID, seqNo)
}

// GetRRPCCardGetCount 读取已授权用户数量 (5.4.1 - 0x8B)
func GetRRPCCardGetCount(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:getcount:%d", channelID, seqNo)
}

// GetRRPCCardGetByPos 按存储位置读取用户 (5.4.2 - 0x8C)
func GetRRPCCardGetByPos(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:getbyoffset:%d", channelID, seqNo)
}

// GetRRPCCardGetByUserID 按用户ID查询 (5.4.3 - 0x8D)
func GetRRPCCardGetByUserID(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:getbyuserid:%d", channelID, seqNo)
}

// GetRRPCCardGetByCardNo 按卡号查询 (5.4.4 - 0x8E)
func GetRRPCCardGetByCardNo(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:card:getbycardno:%d", channelID, seqNo)
}

// GetRRPCDoorRemoteOpen 远程开门（不带操作员信息 0x8A）
func GetRRPCDoorRemoteOpen(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:remoteopen:%d", channelID, seqNo)
}

// GetRRPCDoorRemoteOpenWithOp 远程开门（带操作员信息 0x8B）
func GetRRPCDoorRemoteOpenWithOp(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:remoteopenwithop:%d", channelID, seqNo)
}

// GetRRPCDoorNormallyCloseWithOp 远程常闭门（带操作员信息 0x90）
func GetRRPCDoorNormallyCloseWithOp(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:normallyclosewithop:%d", channelID, seqNo)
}

// GetRRPCDoorNormallyOpenWithOp 远程常开门（带操作员信息 0x91）
func GetRRPCDoorNormallyOpenWithOp(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:normallyopenwithop:%d", channelID, seqNo)
}

// GetRRPCDoorGetState 获取门状态
func GetRRPCDoorGetState(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:getstate:%d", channelID, seqNo)
}

// GetRRPCEventGetAll 获取所有事件
func GetRRPCEventGetAll(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:event:getall:%d", channelID, seqNo)
}

// GetRRPCReadTime 读取实时钟
func GetRRPCReadTime(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:read:time:%d", channelID, seqNo)
}

// GetRRPCTimeGet 获取时间
func GetRRPCTimeGet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:time:get:%d", channelID, seqNo)
}

// GetRRPCTimeSet 设置时间
func GetRRPCTimeSet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:time:set:%d", channelID, seqNo)
}

// GetRRPCTimeGroupSet 设置时段表
func GetRRPCTimeGroupSet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:set:%d", channelID, seqNo)
}

// GetRRPCTimeGroupGet 读取时段表
func GetRRPCTimeGroupGet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:get:%d", channelID, seqNo)
}

// GetRRPCTimeGroupWeekSet 设置星期时段索引
func GetRRPCTimeGroupWeekSet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:weekset:%d", channelID, seqNo)
}

// GetRRPCTimeGroupWeekGet 读取星期时段索引
func GetRRPCTimeGroupWeekGet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:weekget:%d", channelID, seqNo)
}

// GetRRPCTimeGroupHolidaySet 设置特殊日期时段索引
func GetRRPCTimeGroupHolidaySet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:holidayset:%d", channelID, seqNo)
}

// GetRRPCTimeGroupHolidayGet 读取特殊日期时段索引
func GetRRPCTimeGroupHolidayGet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:holidayget:%d", channelID, seqNo)
}

// GetRRPCTimeGroupHolidayDel 删除特殊日期时段索引
func GetRRPCTimeGroupHolidayDel(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:holidaydel:%d", channelID, seqNo)
}

// GetRRPCTimeGroupHolidayClr 清空全部特殊日期
func GetRRPCTimeGroupHolidayClr(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:timegroup:holidayclr:%d", channelID, seqNo)
}

// GetRRPCDoorParamGet 获取门参数
func GetRRPCDoorParamGet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:param:get:%d", channelID, seqNo)
}

// GetRRPCDoorParamSet 设置门参数
func GetRRPCDoorParamSet(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:door:param:set:%d", channelID, seqNo)
}

// GetRRPCRecordGetParam 读取记录参数
func GetRRPCRecordGetParam(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:getparam:%d", channelID, seqNo)
}

// GetRRPCRecordGetSeq 顺序读取记录
func GetRRPCRecordGetSeq(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:getseq:%d", channelID, seqNo)
}

// GetRRPCRecordGetSeqWithPos 顺序读取记录（带位置）
func GetRRPCRecordGetSeqWithPos(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:getseqwithpos:%d", channelID, seqNo)
}

// GetRRPCRecordGetRandom 随机读取记录
func GetRRPCRecordGetRandom(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:getrandom:%d", channelID, seqNo)
}

// GetRRPCRecordGetLatest 获取最新事件
func GetRRPCRecordGetLatest(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:getlatest:%d", channelID, seqNo)
}

// GetRRPCRecordStartAck 启动应答读取模式
func GetRRPCRecordStartAck(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:startack:%d", channelID, seqNo)
}

// GetRRPCRecordAckNext 应答读取下一条
func GetRRPCRecordAckNext(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:acknext:%d", channelID, seqNo)
}

// GetRRPCRecordStopAck 停止应答读取模式
func GetRRPCRecordStopAck(channelID string, seqNo uint8) string {
	return fmt.Sprintf("chd806d4:%s:record:stopack:%d", channelID, seqNo)
}

// ============ 默认配置 ============
const (
	DefaultADR1    = 0x00 // 默认组内地址
	DefaultADR2    = 0x00 // 默认分组地址
	DefaultTimeout = 5000 // 默认超时时间（毫秒）
	DefaultPort    = 8002 // CHD 协议默认端口
)

// ============ 错误信息映射 ============
var RTNMessages = map[uint8]string{
	RTNSuccess:              "执行成功",
	RTNReserved:             "保留",
	RTNChecksumError:        "累加和检查错误",
	RTNLengthError:          "参数长度检查错误",
	RTNInvalidCID:           "无效的数据包类型",
	RTNInvalidCommand:       "不能识别的命令格式",
	RTNInvalidData:          "无效的数据",
	RTNNoPermission:         "无访问权限",
	RTNPasswordError:        "密码确认不正确",
	RTNPasswordModifyFailed: "密码修改不成功",
	RTNStorageFull:          "存储空间已满",
	RTNModifyFailed:         "参数修改失败",
	RTNStorageEmpty:         "存储空间已空",
	RTNNoInfo:               "无相应信息项",
	RTNUserIDDuplicate:      "用户ID重复",
	RTNCardNoDuplicate:      "卡号重复",
	RTNUserInfoDuplicate:    "用户信息全部重复",
}

// GetRTNMessage 获取返回码对应的错误信息
func GetRTNMessage(rtn uint8) string {
	if msg, ok := RTNMessages[rtn]; ok {
		return msg
	}
	return "未知错误"
}

// ============ Redis 常量定义 ============

const (
	// Redis Key 前缀
	RedisKeyCHDPrefix = "dac.chd806d4"

	// Redis 超时时间
	RedisDefaultTimeout = 60 * time.Second
	RedisNoTimeout      = 0

	// 门状态
	DriverDoorStatusOpen  = 1
	DriverDoorStatusClose = 0
)

// GetRRPCSetDoorStatus 获取设置门状态的 RRPC Key
func GetRRPCSetDoorStatus(channelID string) string {
	return fmt.Sprintf("chd806d4:%s:doorstatus:set", channelID)
}

// GetRRPCSetDriverEvent 获取设置事件的 RRPC Key
func GetRRPCSetDriverEvent(channelID string) string {
	return fmt.Sprintf("chd806d4:%s:event:set", channelID)
}

// GetRRPCSetDriverAlarm 获取设置告警的 RRPC Key
func GetRRPCSetDriverAlarm(channelID string) string {
	return fmt.Sprintf("chd806d4:%s:alarm:set", channelID)
}
