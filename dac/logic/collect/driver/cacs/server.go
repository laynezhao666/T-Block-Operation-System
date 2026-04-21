// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/driver/cacs"
	"dac/entity/model/marshaller"
	"dac/entity/utils"
	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/cacs/consts"
	"dac/repo/dac"

	"dac/entity/utils/rrpc"
)

// DoorServer CACS门控器TCP通信服务端，管理与单个门控器的TCP连接。
// 负责数据收发、协议解析、事件告警处理和门状态管理。
type DoorServer struct {
	controllerID          db.IDType // 控制器ID，用于数据库存储
	channelID             string
	serverName            string
	p                     Packet
	timeoutMS             time.Duration
	conn                  net.Conn
	ctx                   context.Context    // Server 独立的 context
	cancel                context.CancelFunc // 用于取消 context
	firstRecvBuff         []byte
	lastRecvBuff          []byte
	isRegistered          bool
	controllerSeq         uint32
	controllerMACAddr     [consts.KMacLen]uint8
	controllerName        string
	eventAlarmSeq         uint32
	mod                   uint8
	fdMutex               sync.Mutex
	uploadDoorStatusMutex sync.Mutex
	uploadDoorStatus      cacs.UploadDoorStatus
	isUploadDoorStatusNew bool
	uploadEventAlarmMutex sync.RWMutex
	uploadEventAlarm      cacs.UploadEventAlarmReq
	events                []cacs.EventAlarmItem
	alarms                []cacs.EventAlarmItem
	currentAlarms         map[uint32]map[uint8]cacs.EventAlarmItem
	isUploadEventAlarmNew bool

	DoorStatusMap map[uint32]cacs.UploadDoorStatus
	isFireAlarm   bool
	FireAlarms    []cacs.UploadControllerStatus
}

// NewDoorServer 创建CACS门控器TCP服务端实例
func NewDoorServer(timeoutMS time.Duration, conn net.Conn,
	serverName string, channelID string,
	controllerID db.IDType,
) *DoorServer {
	return &DoorServer{
		controllerID:  controllerID, // 新增
		timeoutMS:     timeoutMS,
		conn:          conn,
		serverName:    serverName,
		channelID:     channelID,
		events:        make([]cacs.EventAlarmItem, 0),
		alarms:        make([]cacs.EventAlarmItem, 0),
		currentAlarms: make(map[uint32]map[uint8]cacs.EventAlarmItem),
		DoorStatusMap: make(map[uint32]cacs.UploadDoorStatus),
		FireAlarms:    make([]cacs.UploadControllerStatus, 0),
	}
}

// passiveListenServerLoop 被动监听循环，持续接收和处理门控器数据
func (s *DoorServer) passiveListenServerLoop() int {
	ret, err := s.recvData()
	if ret < 0 {
		if err != nil {
			config.Log.Infof("recvData err, try to close, err:%v", err)
		} else {
			config.Log.Infof("recvData err, try to close")
		}
		dtcp.FlushAndClose(s.conn, s.timeoutMS)
		return -1
	}
	ret = s.handle()
	if ret < 0 {
		config.Log.Infof("handle err, try to close")
		dtcp.FlushAndClose(s.conn, s.timeoutMS)
	}
	return ret
}

// handle 处理接收到的数据包，根据命令码分发到对应的处理函数
func (s *DoorServer) handle() int {
	switch s.p.RecvCommand() {
	case consts.KCommandRequestRegister:
		return s.handleRegister()
	case consts.KCommandResponseDoorStatus:
		return s.handleDoorStatus()
	case consts.KCommandResponseRemoteControl:
		return s.handleRemoteControl()
	case consts.KCommandResponseDownloadControllerParams:
		return s.handleDownloadControllerParams()
	case consts.KCommandResponseGetControllerParams:
		return s.handleGetControllerParams()
	case consts.KCommandResponseDownloadDoorParams:
		return s.handleDownloadDoorParams()
	case consts.KCommandResponseGetDoorParams:
		return s.handleGetDoorParams()
	case consts.KCommandResponseDownloadCards:
		return s.handleDownloadCards()
	case consts.KCommandResponseGetCards:
		return s.handleGetCards()
	case consts.KCommandResponseDeleteCards:
		return s.handleDeleteCards()
	case consts.KCommandResponseAddTimeGroups:
		return s.handleAddTimeGroups()
	case consts.KCommandResponseGetTimeGroups:
		return s.handleGetTimeGroups()
	case consts.KCommandResponseDeleteTimeGroups:
		return s.handleDeleteTimeGroups()
	case consts.KCommandUploadDoorStatus:
		return s.handleUploadDoorStatus()
	case consts.KCommandRequestUploadEventAlarm:
		return s.handleUploadEventAlarm()
	case consts.KCommandUploadControllerStatus:
		return s.handleUploadControllerStatus()
	case consts.KCommandResponseSetTime:
		return s.handleSetTime()
	case consts.KCommandResponseGetCardsInfo:
		return s.handleGetCardsInfo()
	default:
		return 0
	}
}

// recvData 从TCP连接接收数据，分两阶段读取（头部+数据体）
func (s *DoorServer) recvData() (int, error) {
	bytes, err := dtcp.ReadN(s.conn, consts.KFirstRecvLen, 0)
	if err != nil {
		return -1, fmt.Errorf("first recv error, err: %v", err)
	}

	s.firstRecvBuff = bytes

	// 记录接收的首包日志
	needLogging := config.C.IsLoggingPacket(s.channelID)
	if needLogging {
		config.Log.Infof("[CACS] recv first: %s", utils.ToHex(s.firstRecvBuff, " "))
	}

	var lastRecvLen uint32 = 0
	if !s.p.ParseFirstRecv(s.firstRecvBuff, &lastRecvLen) {
		config.Log.Infof("解析首次报文错误: recv: %s", utils.ToHex(s.firstRecvBuff, " "))
		return 0, nil
	}
	// CRC字段
	lastRecvLen += 2
	bytes, err = dtcp.ReadN(s.conn, int(lastRecvLen), s.timeoutMS)
	if err != nil {
		return -1, fmt.Errorf("second recv error, err: %v", err)
	}

	if !s.p.ParseLastRecv(bytes) {
		config.Log.Infof("解析末次报文错误: recv: %v", s.firstRecvBuff)
	}
	s.lastRecvBuff = s.p.recvData

	// 记录接收的末包日志
	if needLogging {
		config.Log.Infof("[CACS] recv last: %s", utils.ToHex(bytes, " "))
		// ========== 打印收到的数据包详细信息（仅在开启日志时打印）==========
		s.printReceivedPacket()
	}

	return 0, nil
}

// 命令码名称映射表（包含请求和响应）
var cmdNames = map[uint32]string{
	// 请求命令
	consts.KCommandRequestRegister:                 "注册请求",
	consts.KCommandRequestDoorStatus:               "门状态请求",
	consts.KCommandRequestRemoteControl:            "远程控制请求",
	consts.KCommandRequestDownloadControllerParams: "下载控制器参数请求",
	consts.KCommandRequestGetControllerParams:      "获取控制器参数请求",
	consts.KCommandRequestDownloadDoorParams:       "下载门参数请求",
	consts.KCommandRequestGetDoorParams:            "获取门参数请求",
	consts.KCommandRequestDownloadCards:            "下载卡请求",
	consts.KCommandRequestGetCards:                 "获取卡请求",
	consts.KCommandRequestDeleteCards:              "删除卡请求",
	consts.KCommandRequestAddTimeGroups:            "添加时间组请求",
	consts.KCommandRequestGetTimeGroups:            "获取时间组请求",
	consts.KCommandRequestDeleteTimeGroups:         "删除时间组请求",
	consts.KCommandRequestSetTime:                  "设置时间请求",
	consts.KCommandRequestGetCardsInfo:             "获取卡信息请求",
	// 响应命令
	consts.KCommandResponseDoorStatus:               "门状态响应",
	consts.KCommandResponseRemoteControl:            "远程控制响应",
	consts.KCommandResponseDownloadControllerParams: "下载控制器参数响应",
	consts.KCommandResponseGetControllerParams:      "获取控制器参数响应",
	consts.KCommandResponseDownloadDoorParams:       "下载门参数响应",
	consts.KCommandResponseGetDoorParams:            "获取门参数响应",
	consts.KCommandResponseDownloadCards:            "下载卡响应",
	consts.KCommandResponseGetCards:                 "获取卡响应",
	consts.KCommandResponseDeleteCards:              "删除卡响应",
	consts.KCommandResponseAddTimeGroups:            "添加时间组响应",
	consts.KCommandResponseGetTimeGroups:            "获取时间组响应",
	consts.KCommandResponseDeleteTimeGroups:         "删除时间组响应",
	consts.KCommandResponseSetTime:                  "设置时间响应",
	consts.KCommandResponseGetCardsInfo:             "获取卡信息响应",
	// 主动上报命令
	consts.KCommandUploadDoorStatus:        "门状态上报",
	consts.KCommandRequestUploadEventAlarm: "事件告警上报",
	consts.KCommandUploadControllerStatus:  "控制器状态上报",
}

// getCommandName 获取命令名称
func getCommandName(cmd uint32) string {
	if name, exists := cmdNames[cmd]; exists {
		return name
	}
	return "未知命令"
}

// printHexDump 打印十六进制数据（每行16字节，带ASCII显示）
func printHexDump(data []byte) {
	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}

		// 打印偏移量
		fmt.Printf("%04X: ", i)

		// 打印十六进制
		for j := i; j < end; j++ {
			fmt.Printf("%02X ", data[j])
		}

		// 补齐空格
		for j := end; j < i+16; j++ {
			fmt.Printf("   ")
		}

		// 打印ASCII
		fmt.Printf(" | ")
		for j := i; j < end; j++ {
			if data[j] >= 32 && data[j] <= 126 {
				fmt.Printf("%c", data[j])
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}
}

// printReceivedPacket 打印收到的数据包详细信息
// 注意：此函数仅在 IsLoggingPacket 返回 true 时由 recvData 调用，无需再次检查
func (s *DoorServer) printReceivedPacket() {
	cmd := s.p.RecvCommand()
	cmdName := getCommandName(cmd)

	config.Log.Infof("\n" + strings.Repeat("=", 20) + "📦 收到数据包" + strings.Repeat("=", 20))
	config.Log.Infof("门控器: %s (MAC: %s)", s.controllerName, utils.ToHex(s.controllerMACAddr[:], ":"))
	config.Log.Infof("命令码: 0x%04X (%s)", cmd, cmdName)
	config.Log.Infof("数据长度: %d 字节", len(s.lastRecvBuff))
	config.Log.Infof("时间: %s", time.Now().Format("2006-01-02 15:04:05.000"))

	// 打印RTN信息
	if s.p.HasRtn() {
		rtn := s.p.GetRtn()
		rtnDesc := consts.RtnInfoMap[rtn]
		if rtnDesc == "" {
			rtnDesc = "未知错误"
		}
		if rtn == consts.KRtnNormal {
			config.Log.Infof("RTN: 0x%08X (%s) ✅", rtn, rtnDesc)
		} else {
			config.Log.Infof("RTN: 0x%08X (%s) ❌", rtn, rtnDesc)
		}
	} else {
		config.Log.Infof("RTN: 无 (主动上报数据包)")
	}

	// 打印完整报文（包含包头）
	config.Log.Infof("\n【报文内容】(包头 + 长度 + 命令 + 业务数据，不含RTN和CRC):")
	fullPacket := append(s.firstRecvBuff, s.lastRecvBuff...)
	config.Log.Infof("总长度: %d 字节", len(fullPacket))
	printHexDump(fullPacket)
	config.Log.Infof("")

}

// handleRegister 处理设备注册请求，提取MAC地址并标记已注册
func (s *DoorServer) handleRegister() int {
	// 注册包已在 HandleConnWithMACVerify 中处理
	// 如果在这里收到注册包，说明是重复注册，直接忽略
	config.Log.Infof("收到重复注册包，已忽略，MAC: %s", utils.ToHex(s.controllerMACAddr[:], ":"))
	return 0
}

// Close 关闭TCP连接并取消上下文
func (s *DoorServer) Close() {
	s.fdMutex.Lock()
	defer s.fdMutex.Unlock()
	s.conn.Close()
}

// Request 向门控器发送请求数据包（加锁保证线程安全）
func (s *DoorServer) Request(cmd uint32, data []byte) int {
	if !s.isRegistered {
		return -1
	}

	var packet Packet
	packet.BuildRequestPacket(cmd, data)
	sendBuff := packet.SendData()

	// 记录发送的请求包日志（仅在开启日志时打印详细信息）
	if config.C.IsLoggingPacket(s.channelID) {
		config.Log.Infof("[CACS] send cmd=0x%04X: %s", cmd, utils.ToHex(sendBuff, " "))

		// 详细日志打印（测试环境）
		config.Log.Infof("")
		config.Log.Infof("====================📤 发送数据包====================")
		config.Log.Infof("门控器: %s (MAC: %s)", s.controllerName, utils.ToHex(s.controllerMACAddr[:], ":"))
		config.Log.Infof("命令码: 0x%04X (%s)", cmd, getCommandName(cmd))
		config.Log.Infof("数据长度: %d 字节", len(data))
		config.Log.Infof("时间: %s", time.Now().Format("2006-01-02 15:04:05.000"))
		config.Log.Infof("")
		config.Log.Infof("【完整报文】(包头 + 数据):")
		config.Log.Infof("总长度: %d 字节", len(sendBuff))
		printHexDump(sendBuff)
		config.Log.Infof("")
		config.Log.Infof("========================================")
	}

	var sendLen int = -1
	var err error
	func() {
		s.fdMutex.Lock()
		defer s.fdMutex.Unlock()
		sendLen, err = dtcp.WriteN(s.conn, sendBuff, s.timeoutMS)
	}()
	if sendLen != len(sendBuff) {
		if err != nil {
			config.Log.Infof("发送注册响应报文失败, err: %v,  bytesSend: %d, try to close", err, sendLen)
		} else {
			config.Log.Infof("发送注册响应报文失败 bytesSend: %d, try to close", sendLen)
		}
		dtcp.FlushAndClose(s.conn, s.timeoutMS)
		return -1
	}
	return 0

}

// handleDoorStatus 处理门状态查询响应
func (s *DoorServer) handleDoorStatus() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}

	// rtn如果不为0x00，直接返回数据包，让上层判断，不做长度校验
	return s.checkLenAndSetRRPC(consts.GetRRPCDoorStatus(s.channelID), cacs.GetFieldSizeSum(cacs.DoorStateResp{}))

}

// handleRemoteControl 处理远程控制门禁响应
func (s *DoorServer) handleRemoteControl() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.DoorControlResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(consts.GetRRPCRemoteControl(s.channelID), s.lastRecvBuff)
	return 0
}

// handleDownloadControllerParams 处理下载控制器参数响应
func (s *DoorServer) handleDownloadControllerParams() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.DownloadControllerParamsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCDownloadControllerParams(s.channelID),
		s.lastRecvBuff)
	return 0
}

// handleGetControllerParams 处理获取控制器参数响应
func (s *DoorServer) handleGetControllerParams() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.GetControllerParamsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCGetControllerParams(s.channelID),
		s.lastRecvBuff)
	return 0
}

// handleDownloadDoorParams 处理下载门参数响应
func (s *DoorServer) handleDownloadDoorParams() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.DownloadDoorParamsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCDownloadDoorParams(s.channelID),
		s.lastRecvBuff)
	return 0
}

// handleGetDoorParams 处理获取门参数响应
func (s *DoorServer) handleGetDoorParams() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	// 使用 checkLenAndSetRRPC：RTN!=0 时不校验长度，不断链，让上层业务处理错误
	return s.checkLenAndSetRRPC(consts.GetRRPCGetDoorParams(s.channelID), cacs.GetFieldSizeSum(cacs.GetDoorParamsResp{}))
}

// handleDownloadCards 处理下载卡数据响应
func (s *DoorServer) handleDownloadCards() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}

	return s.checkLenAndSetRRPC(consts.GetRRPCDownloadCards(s.channelID), cacs.GetFieldSizeSum(cacs.DownloadCardsResp{}))
}

// handleGetCards 处理获取卡数据响应
func (s *DoorServer) handleGetCards() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.GetCardsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCGetCards(s.channelID), s.lastRecvBuff)
	return 0
}

// handleDeleteCards 处理删除卡数据响应
func (s *DoorServer) handleDeleteCards() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.DeleteCardsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCDeleteCards(s.channelID), s.lastRecvBuff)
	return 0
}

// handleAddTimeGroups 处理添加时间组响应
func (s *DoorServer) handleAddTimeGroups() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.AddTimeGroupsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCAddTimeGroups(s.channelID), s.lastRecvBuff)
	return 0
}

// handleGetTimeGroups 处理获取时间组响应
func (s *DoorServer) handleGetTimeGroups() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}

	// rtn如果不为0x00，直接返回数据包，让上层判断，不做长度校验
	return s.checkLenAndSetRRPC(consts.GetRRPCGetTimeGroups(s.channelID), cacs.GetFieldSizeSum(cacs.GetTimeGroupsResp{}))
	//if len(s.lastRecvBuff) != cacs.GetFieldSizeSum(cacs.GetTimeGroupsResp{}) {
	//	config.Log.Infof("报文长度错误: %d, 期望: %d", len(s.lastRecvBuff), cacs.GetFieldSizeSum(cacs.GetTimeGroupsResp{}))
	//	return -1
	//}
	//rrpc.Manager().Set(consts.KRRPCGetTimeGroups, s.lastRecvBuff)
	//return 0
}

// handleDeleteTimeGroups 处理删除时间组响应
func (s *DoorServer) handleDeleteTimeGroups() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.DeleteTimeGroupsResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCDeleteTimeGroups(s.channelID),
		s.lastRecvBuff)
	return 0
}

// handleUploadDoorStatus 处理设备主动上报的门状态数据
func (s *DoorServer) handleUploadDoorStatus() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.UploadDoorStatus{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	req, err := marshaller.RequestUploadDoorStatusUnMarshal(s.lastRecvBuff)
	if err != nil {
		config.Log.Infof("RequestUploadDoorStatusUnMarshal error: %v", err)
		return -1
	}
	func() {
		s.uploadDoorStatusMutex.Lock()
		defer s.uploadDoorStatusMutex.Unlock()
		s.uploadDoorStatus = req
		s.isUploadDoorStatusNew = true
	}()
	return 0
}

// recvDoorStatus 后台协程持续接收门状态更新
func (s *DoorServer) recvDoorStatus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			func() {
				s.uploadDoorStatusMutex.Lock()
				defer s.uploadDoorStatusMutex.Unlock()
				if s.isUploadDoorStatusNew {
					s.isUploadDoorStatusNew = false
				}
			}()
			time.Sleep(time.Second)
		}
	}
}

// handleUploadEventAlarm 处理设备主动上报的事件和告警数据
func (s *DoorServer) handleUploadEventAlarm() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	req, err := marshaller.RequestUploadEventAlarmUnMarshal(s.lastRecvBuff)
	if err != nil {
		config.Log.Infof("RequestUploadEventAlarmUnMarshal err: %v", err)
		data := marshaller.ResponseUploadEventAlarm(cacs.UploadEventAlarmResp{
			SuccessNum:    0,
			EventAlarmSeq: 0xff,
		})
		s.p.BuildResponsePacket(consts.KCommandResponseUploadEventAlarm, consts.KRtnParamLenError, data)
		return -1
	}
	// 如果平台自身的序号与控制器上传的序号相同则说明控制器上传的事件告警为重复内容，平台将不予处理
	if req.Seq == s.eventAlarmSeq {
		return 0
	}
	func() {
		s.uploadEventAlarmMutex.Lock()
		defer s.uploadEventAlarmMutex.Unlock()
		s.uploadEventAlarm = req
		s.eventAlarmSeq = req.Seq
		s.isUploadEventAlarmNew = true
	}()

	data := marshaller.ResponseUploadEventAlarm(cacs.UploadEventAlarmResp{
		SuccessNum:    req.Num,
		EventAlarmSeq: s.eventAlarmSeq,
	})
	s.p.BuildResponsePacket(consts.KCommandResponseUploadEventAlarm, 0, data)
	sendBuff := s.p.SendData()
	var sendLen int
	func() {
		s.fdMutex.Lock()
		defer s.fdMutex.Unlock()
		sendLen, err = dtcp.WriteN(s.conn, sendBuff, s.timeoutMS)
	}()

	// 记录发送日志
	if config.C.IsLoggingPacket(s.channelID) {
		config.Log.Infof("[CACS] send: %s", utils.ToHex(sendBuff, " "))
	}

	if sendLen != len(sendBuff) {
		if err != nil {
			config.Log.Infof("发送注册响应报文失败, err: %v,  bytesSend: %d, try to close", err, sendLen)
		} else {
			config.Log.Infof("发送注册响应报文失败 bytesSend: %d, try to close", sendLen)
		}
		dtcp.FlushAndClose(s.conn, s.timeoutMS)
		return -1
	}
	return 0
}

// handleSetTime 处理设置时间响应
func (s *DoorServer) handleSetTime() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.SetTimeResp{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(
		consts.GetRRPCSetTime(s.channelID), s.lastRecvBuff)
	return 0
}

// handleUploadControllerStatus 处理设备主动上报的控制器状态（火警等）
func (s *DoorServer) handleUploadControllerStatus() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	expectLen := cacs.GetFieldSizeSum(cacs.UploadControllerStatus{})
	if len(s.lastRecvBuff) != expectLen {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	req, err := marshaller.ReqUploadCtrlStatusUnmarshal(s.lastRecvBuff)
	if err != nil {
		config.Log.Infof("RequestUploadDoorStatusUnMarshal error: %v", err)
		return -1
	}
	if req.FireAlarmStatus == 1 {
		s.isFireAlarm = false
	} else if req.FireAlarmStatus == 0 {
		s.isFireAlarm = true
		s.FireAlarms = append(s.FireAlarms, req)
	}

	return 0
}

// saveEventAlarm 后台协程持续保存事件和告警到数据库
func (s *DoorServer) saveEventAlarm(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.isUploadEventAlarmNew {
				var items []cacs.EventAlarmItem
				func() {
					s.uploadEventAlarmMutex.Lock()
					defer s.uploadEventAlarmMutex.Unlock()
					items = s.uploadEventAlarm.Items
					for i := range items {
						item := items[i]
						// item.Type位uint8, CACS协议规定当Type为(0-127)时，为事件类型，为(128-225)时，为告警类型
						if item.Type <= 127 {
							s.events = append(s.events, item)
							// 存入数据库
							if err := s.saveEventToDB(ctx, item, len(s.events)-1); err != nil {
								config.Log.Infof("save event to db error: %v", err)
							}
						} else {
							s.alarms = append(s.alarms, item)
							// 存入数据库
							if err := s.saveAlarmToDB(ctx, item, len(s.alarms)-1); err != nil {
								config.Log.Infof("save alarm to db error: %v", err)
							}
							// CurrentAlarm 目前只关注门常开告警
							if item.Type == consts.KAlarmDoorOpenTimeout {
								if item.Extras == consts.KAlarmOn {
									// 确保内层 map 已初始化
									if s.currentAlarms[item.DoorId] == nil {
										s.currentAlarms[item.DoorId] = make(map[uint8]cacs.EventAlarmItem)
									}
									s.currentAlarms[item.DoorId][item.Type] = item
								} else if item.Extras == consts.KAlarmOff {
									if s.currentAlarms[item.DoorId] != nil {
										delete(s.currentAlarms[item.DoorId], item.Type)
									}
								} else {
									// 如果Extras不为KAlarmOn或者KAlarmOff，与协议规定不同，忽略
									continue
								}
							}
						}
					}
					s.isUploadEventAlarmNew = false
				}()
			}
			// 没有新事件时，休眠避免空转
			time.Sleep(time.Second)
		}
	}
}

// saveEventToDB 将事件存入数据库
func (s *DoorServer) saveEventToDB(ctx context.Context, item cacs.EventAlarmItem, index int) error {
	// 数据库未初始化时跳过（测试模式）
	if !dac.IsInitialized() {
		return nil
	}

	timestamp, err := utils.StringToUnixTime(fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		item.Year, item.Month, item.Day, item.Hour, item.Minute, item.Second))
	if err != nil {
		return fmt.Errorf("parse timestamp error: %v", err)
	}

	desc := consts.EventAlarmInfoMap[item.Type]
	if desc == "" {
		desc = "未知事件"
	}

	dbEvent := db.DriverEvent{
		ControllerID: s.controllerID,
		ChannelID:    s.channelID,
		Index:        index,
		Timestamp:    timestamp,
		CardNumber:   fmt.Sprintf("%d", item.CardId),
		Username:     "", // CACS协议事件格式没有用户名
		DoorNumber:   db.DoorNumberType(item.DoorId),
		Direction:    db.DirectionType(item.CardReaderId),
		Type:         db.EventType(item.Type),
		Description:  desc,
	}

	// 检查事件是否已存在，避免重复插入
	if _, err := dac.GetRW().GetDriverEvent(ctx, s.controllerID, dbEvent); err == nil {
		return nil // 事件已存在，跳过
	}

	return dac.GetRW().SetDriverEvents(ctx, s.controllerID, []db.DriverEvent{dbEvent})
}

// saveAlarmToDB 将告警存入数据库
func (s *DoorServer) saveAlarmToDB(ctx context.Context, item cacs.EventAlarmItem, index int) error {
	// 数据库未初始化时跳过（测试模式）
	if !dac.IsInitialized() {
		return nil
	}

	timestamp, err := utils.StringToUnixTime(fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		item.Year, item.Month, item.Day, item.Hour, item.Minute, item.Second))
	if err != nil {
		return fmt.Errorf("parse timestamp error: %v", err)
	}

	desc := consts.EventAlarmInfoMap[item.Type]
	if desc == "" {
		desc = "未知告警"
	}

	dbAlarm := db.DriverAlarm{
		ControllerID: s.controllerID,
		ChannelID:    s.channelID,
		Index:        index,
		Timestamp:    timestamp,
		DoorNumber:   db.DoorNumberType(item.DoorId),
		Type:         db.AlarmType(item.Type),
		State:        db.AlarmStateType(item.Extras), // Extras表示告警状态：0x01告警产生，0x02告警恢复
		Description:  desc,
	}

	// 检查告警是否已存在，避免重复插入
	if _, err := dac.GetRW().GetDriverAlarm(ctx, s.controllerID, dbAlarm); err == nil {
		return nil // 告警已存在，跳过
	}

	return dac.GetRW().SetDriverAlarms(ctx, s.controllerID, []db.DriverAlarm{dbAlarm})
}

// handleGetCardsInfo 处理获取卡详细信息响应
func (s *DoorServer) handleGetCardsInfo() int {
	if !s.isRegistered {
		config.Log.Infof("错误：当前门禁控制器未注册")
		return -1
	}
	data := s.lastRecvBuff
	var resp cacs.GetCardsInfoResp
	if len(data) < int(unsafe.Sizeof(resp.NextIndex))+int(unsafe.Sizeof(resp.Num)) {
		return -1
	}
	buf := bytes.NewBuffer(data)
	resp.NextIndex = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.NextIndex))))
	resp.Num = dtcp.ReadUint32Little(buf.Next(int(unsafe.Sizeof(resp.Num))))
	cardInfoSize := cacs.GetFieldSizeSum(cacs.CardInfo{})
	expectLen := int(unsafe.Sizeof(resp.Num)) +
		int(unsafe.Sizeof(resp.NextIndex)) +
		int(resp.Num)*cardInfoSize
	if expectLen != len(data) {
		config.Log.Infof("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectLen)
		return -1
	}
	rrpc.Manager().Set(consts.GetRRPCGetCardsInfo(s.channelID), data)
	return 0
}

// checkLenAndSetRRPC 根据 RTN 决定是否校验长度，并设置 RRPC
// - RTN=0 时校验长度，长度不符返回 -1（断开链接）
// - RTN!=0 时不校验长度（因为协议并未规定 RTN!=0 时没有业务数据而实际不一定）
// - 无论 RTN 是什么，都传递数据给 RRPC，让上层业务处理 RTN 错误
func (s *DoorServer) checkLenAndSetRRPC(rrpcKey string, expectedLen int) int {
	if s.p.GetRtn() == consts.KRtnNormal {
		if len(s.lastRecvBuff) != expectedLen {
			config.Log.Infof("报文长度错误: %d, 期望: %d", len(s.lastRecvBuff), expectedLen)
			return -1 // 只有 RTN=0 且长度错误才断链
		}
	}
	// RTN!=0 或 长度正确，都传给 RRPC，不断链
	rrpc.Manager().Set(rrpcKey, s.lastRecvBuff)
	return 0
}
