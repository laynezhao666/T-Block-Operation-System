package iec103_siemens

import model2 "agent/entity/consts"

const (
	IEC103_HeadLen    = 28             // 完整的一个心跳报文长度
	IEC103_FixHeadLen = 8              // 在头部里的固定字节长度，就是90EB+4个字节表示长度+90EB
	IEC103_LenHeadVal = 20             // 填在心跳头里的长度，最少为20，就是HeadLen减去8
	IEC103_AsduOffset = IEC103_HeadLen // ASDU在报文里的偏移量
	IEC103_CotOffset  = 30             // 在报文里的COT偏移量
	IEC103_AsduFixLen = 8              // 在报文头后有8个字节携带固定的信息，包括ASDU类型等
	IEC103_MinItemLen = 6              // 在报文头后ASDU8个固定字节后的值item最小长度为6
)

const (
	COT_Active    = 0x01 // 自发突发
	COT_Cycle     = 0x02 // 循环上报
	COT_ClockSync = 0x08 // 时钟同步
	COT_Call      = 0x09 // 呼叫上报
	COT_QueryEnd  = 0x0A // 召唤查询结束
	COT_Read      = 0x2A // 通用分类读命令，用于电度召唤
	COT_Write     = 0x28 // 通用分类写
	COT_WriteFail = 0x29 // 通用分类写否定确认
	COT_Write_Ack = 0x2C // 通用分类写命令确认
)

// 信息序号常量
const (
	INF_WriteWithConfirm = 0xF9 // 带确认的写条目（选择）
	INF_WriteWithExecute = 0xFA // 带执行的写条目（执行）
)

// 控制值常量
const (
	ControlValue_Off = 0x01 // 分
	ControlValue_On  = 0x02 // 合
)

const (
	ASDU_Type_CtrlSvc   = 0x15 // 通用分类服务控制方向
	ASDU_Type_ClockSync = 0x06 // 时钟同步
)

const (
	Spontaneous_communication = "spontaneous_communication" // 自发通信
	Cyclic_measurement        = "cyclic_measurement"        // 循环上报
	Call_data                 = "call_data"                 // 召唤所有数据
	Call_energy               = "call_energy"               // 召唤电度数据
	Fault_data                = "fault_data"                // 失败载波数据
	Query_End                 = "query_end"                 // 查询结束
	Write_Rsp                 = "write_rsp"                 // 写命令响应
	Unknown                   = "unknown"
)

// Header 28字节报文头
type Header struct {
	Start1        uint16
	Length        uint32
	Start2        uint16
	SourceStation uint16
	SourceDevice  uint16
	TargetStation uint16
	TargetDevice  uint16
	DataNumber    uint16
	DeviceType    uint16
	NetworkStatus uint16
	FirstRoute    uint16
	LastRoute     uint16
	EndMark       uint16
}

// ASDU 应用服务数据单元
type ASDU struct {
	TypeID     uint8
	VSQ        uint8
	COT        uint8
	CommonAddr uint8
	FunType    uint8
	InfNum     uint8
	Data       []byte
}

// ActiveReport 主动上报数据
type ActiveReport struct {
	Data      []byte
	Type      string
	Timestamp int64
}

// DataPoint 数据点
type DataPoint struct {
	Addr      uint32         `json:"address"`   // 地址组号条目号
	Value     interface{}    `json:"value"`     // 值
	Qua       model2.Quality `json:"quality"`   // 品质描述
	Ms        int64          `json:"ms"`        // 上游数据上报的时间戳，单位毫秒
	Cts       int64          `json:"cts"`       // 本地采集的数据时间戳，单位毫秒
	ExpiredMs int64          `json:"expiredMs"` // 过期时间，遥信和遥测的过期时间不一样,单位毫秒
}

// ControlData 控制命令数据
type ControlData struct {
	SelectData  []byte
	ExecuteData []byte
}

// CombineAddr 将组号和条目号组合成地址
// 组合方式：Addr = uint32(groupNum)<<8 | uint32(entryNum)
func CombineAddr(groupNum, entryNum byte) uint32 {
	return uint32(groupNum)<<8 | uint32(entryNum)
}

// ParseAddrToGroupEntry 将地址反解析为组号和条目号
// 反解析方式：groupNum = (addr >> 8) & 0xFF, entryNum = addr & 0xFF
func ParseAddrToGroupEntry(addr uint32) (groupNum, entryNum byte) {
	groupNum = byte((addr >> 8) & 0xFF)
	entryNum = byte(addr & 0xFF)
	return
}
