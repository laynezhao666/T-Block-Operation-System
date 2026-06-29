package iec104

import "time"

/*
IEC104规约的超时和报文丢失重发的处理机制

为了能对TCP链接进行检查和维护，IEC 60870-5-104规定了几个超时时间，即t0、 t1、t2、t3，它们的取值范围为1~255s，准确度为1s。
t0规定了主站端和子站RTU端创建一次TCP链接的最大容许时间，主站端和子站 RTU端之间的TCP链接在实际运行中可能常常进行关闭和重建，
这发生在4种状况下：

	① 主站端和子站RTU端之间的I格式报文传送出现丢失、错序或者发送U格式报文得不到应答时，双方都可主动关闭TCP链接，而后进行重建;
	② 主站系统重新启动后将与各个子站从新创建TCP链接；
	③ 子站RTU合上电源或因为自恢复而从新启动后，将重建链接；
	④子站RTU收到主站端的RESET_PROCESS（复位远方终端）信号后，将关闭链接并从新初始化，而后重建链接。每次创建链接时，
		RTU都调用socket的listen()函数进行侦听，主站端调用socket的connect()函数进行连接，若是在t0时间内未能成功创建链接，
		可能网络发生了故障，主站端应该向运行人员给出警告信息。

t1规定发送方发送一个I格式报文或U格式报文后，必须在t1的时间内获得接收方的承认，不然发送方认为TCP链接出现问题并应重新创建连接。
t2规定接收方在接收到I格式报文后，若通过t2时间未再收到新的I格式报文，则必须向发送方发送S格式帧对已经接收到的I格式报文进行承认，

	显然t2必须小于t1。

t3规定调度端或子站RTU端每接收一帧I帧、S帧或者U帧将重新触发计时器t3，若在t3内未接收到任何报文，将向对方发送测试链路帧。
*/

const (
	defaultElectricityTotalCallInterval = 5 * time.Minute  // 厂商建议时间
	defaultTotalCallInterval            = 15 * time.Minute // 总召唤定时间周期
	defaultTimeoutT0                    = 30 * time.Second // T0, 建立TCP/IP连接的超时时间
	defaultTimeoutT1                    = 15 * time.Second // T1, 发送或测试APDU的超时
	defaultTimeoutT2                    = 10 * time.Second // T2, 无数据报文t2<t1时确认的超时
	defaultTimeoutT3                    = 20 * time.Second // T3, 长期空闲t3>t1状态下发送测试帧的超时
	defaultMaxReadTimeout               = time.Minute
)

const (
	SFrameFrequencyMin = 1  // S帧回复频率最小值
	SFrameFrequencyMax = 20 // S帧回复频率最大值
)

// IEC104协议约束
const (
	apciLen        = 6 // 应用规约控制信息长度
	asduLen        = 6 // 数据单位标识符长度
	seqAddrLen     = 3 // 连续性地址长度
	startAndNumLen = 2 // 起始符和数据长度报文长度
)

// APCI 解析偏移
const (
	APCIOffsetStart = iota
	APCIOffsetApduLen
	APCIOffsetCtr1
	APCIOffsetCtr2
	APCIOffsetCtr3
	APCIOffsetCtr4
)

// ASDU 信息偏移
const (
	ASDUOffsetTypeID = iota
	ASDUOffsetSequenceAndNum
	ASDUOffsetCause0
	ASDUOffsetCause1
	ASDUOffsetPublicAddress0
	ASDUOffsetPublicAddress1
)

// 数据类型
const (
	MSpNa1 = 1   // 不带游标的单点遥信，3个字节的地址，1个字节的值
	MDpNa1 = 3   // 不带时标的双点遥信，每个遥信占1个字节
	MMeNa1 = 9   // 带品质描述的测量值，每个遥测值占3个字节
	MMeNc1 = 13  // 带品质描述的浮点值，每个遥测值占5个字节
	MItNa1 = 15  // 电度总量,每个遥脉值占5个字节
	MSpTb1 = 30  // 带游标的单点遥信，3个字节的地址，1个字节的值，7个字节短时标
	MEiNA1 = 70  // 初始化结束
	CIcNa1 = 100 // 总召唤
	CCiNa1 = 101 // 电度总召唤
)

// 数据传送原因，《DLT 634.5101-2002》 7.2.3.1
const (
	COTPerCyc   = 1  // 周期、循环
	COTSpont    = 3  // 突发(自发)
	COTInit     = 4  // 初始化
	COTReq      = 5  // 请求或者被请求
	COTAct      = 6  // 激活
	COTActCon   = 7  // 激活确认
	COTDeAct    = 8  // 停止激活
	COTDeActCon = 9  // 停止激活确认
	COTActTerm  = 10 // 激活终止
	COTIntrgen  = 20 // 响应站召唤
	COTReqCoGen = 37 // 响应计数量站召唤
)

// StartFrame 起始符
const startFrame = 0x68

var (
	startDtAct = [4]byte{0x07, 0x00, 0x00, 0x00} // 启动激活帧
	startDtCon = [4]byte{0x0b, 0x00, 0x00, 0x00} // 启动确认帧
	testFrAct  = [4]byte{0x43, 0x00, 0x00, 0x00} // 测试激活帧
	testFrCon  = [4]byte{0x83, 0x00, 0x00, 0x00} // 测试确认帧
	stopDtAct  = [4]byte{0x13, 0x00, 0x00, 0x00} // 停止激活帧
	stopDtCon  = [4]byte{0x23, 0x00, 0x00, 0x00} // 停止确认帧

	totalCallAct            = []byte{0x64, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x14} // 总召
	electricitytotalCallAct = []byte{0x65, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x05} // 电度总召
)

// 帧标识
const (
	iFrame byte = 0
	sFrame byte = 1
	uFrame byte = 3
)
