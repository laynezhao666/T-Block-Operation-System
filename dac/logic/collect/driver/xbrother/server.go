// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/model/marshaller"
	"dac/entity/utils"
	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/xbrother/consts"

	"dac/entity/utils/rrpc"
	"dac/entity/utils/tlog"
)

// DoorServer XBrother协议TCP通信服务端，
// 负责与门禁控制器建立连接、收发报文和处理命令。
type DoorServer struct {
	channelID            string                                  // 通道ID（控制器IP:端口）
	serverName           string                                  // 服务名称
	p                    *Packet                                 // 报文解析器
	timeout              time.Duration                           // 通信超时时间
	conn                 net.Conn                                // TCP连接
	fdMutex              sync.Mutex                              // 连接读写锁
	firstRecvBuff        []byte                                  // 首次接收缓冲区
	lastRecvBuff         []byte                                  // 末次接收缓冲区
	eventChan            chan xbrother.EventUploadReq            // 事件上传通道
	alarmChan            chan xbrother.AlarmUploadReq            // 告警上传通道
	controllerStatusChan chan xbrother.ControllerStatusUploadReq // 控制器状态上传通道
	doorNum              int                                     // 门数量
	baseInfo             driver.ControllerBasicInfo              // 控制器基本信息
	disConnectChan       chan struct{}                           // 断连通知通道
	logger               tlog.Logger                             // 日志记录器
}

// NewDoorServer 创建XBrother协议TCP服务端实例
func NewDoorServer(timeout time.Duration, serverName string,
	channelID string, doorNum int,
	baseInfo driver.ControllerBasicInfo,
) *DoorServer {
	logger := tlog.NewPrefixLogger(fmt.Sprintf("[%v@%v]", baseInfo.ID, channelID), config.Log)
	s := &DoorServer{
		channelID:            channelID,
		serverName:           serverName,
		p:                    NewPacket(logger),
		timeout:              timeout,
		firstRecvBuff:        make([]byte, 0),
		lastRecvBuff:         make([]byte, 0),
		eventChan:            make(chan xbrother.EventUploadReq, consts.ChanInitLength),
		alarmChan:            make(chan xbrother.AlarmUploadReq, consts.ChanInitLength),
		controllerStatusChan: make(chan xbrother.ControllerStatusUploadReq, consts.ChanInitLength),
		doorNum:              doorNum,
		baseInfo:             baseInfo,
		disConnectChan:       make(chan struct{}, consts.ChanInitLength),
		logger:               logger,
	}
	return s
}

// Connect 建立TCP连接并启动数据处理协程
func (s *DoorServer) Connect(ctx context.Context) error {
	dialer := &net.Dialer{Timeout: s.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", s.channelID)
	if err != nil {
		return fmt.Errorf("dial %v error: %w", s.channelID, err)
	}
	func() {
		s.fdMutex.Lock()
		defer s.fdMutex.Unlock()
		s.conn = conn
	}()
	go HandleConn(ctx, s)
	return nil
}

// Close 关闭TCP连接
func (s *DoorServer) Close() {
	s.fdMutex.Lock()
	defer s.fdMutex.Unlock()

	if err := s.conn.Close(); err != nil {
		s.logger.Errorf("close conn error, err: %v", err)
	} else {
		s.logger.Infof("close connection success")
	}
}

// serverLoop 单次服务循环：接收数据并处理命令
func (s *DoorServer) serverLoop() error {
	var err error
	if err = s.recvData(); err != nil {
		s.logger.Errorf("recvData err, try to close, err: %v", err)
		dtcp.FlushAndClose(s.conn, s.timeout)
		return err
	}
	if err = s.handle(); err != nil {
		s.logger.Errorf("handle err, try to close, err: %v", err)
		dtcp.FlushAndClose(s.conn, s.timeout)
	}
	return err
}

// isConnectionClosed 判断错误是否为连接关闭
func isConnectionClosed(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET)
}

// recvData 从TCP连接接收报文数据（分两次接收：首次和末次）
func (s *DoorServer) recvData() error {
	bytes, err := dtcp.ReadN(s.conn, consts.FirstRecvLen, 0)
	if err != nil {
		if isConnectionClosed(err) {
			s.logger.Warnf("disconnected when first recv")
			s.disConnectChan <- struct{}{}
		}
		return fmt.Errorf("first recv error, err: %w", err)
	}
	s.firstRecvBuff = bytes

	var lastRecvLen int = 0
	var ok bool
	if lastRecvLen, ok = s.p.ParseFirstRecv(s.firstRecvBuff); !ok {
		s.logger.Errorf("解析首次报文错误: recv: %s", utils.ToHex(s.firstRecvBuff, " "))
		return nil
	}

	needLogging := config.C.IsLoggingPacket(s.channelID)
	if needLogging {
		s.logger.Infof("recv: %v", utils.ToHex(s.firstRecvBuff, " "))
	}

	bytes, err = dtcp.ReadN(s.conn, lastRecvLen, s.timeout)
	if err != nil {
		if isConnectionClosed(err) {
			s.logger.Warnf("disconnected when second recv")
			s.disConnectChan <- struct{}{}
		}
		return fmt.Errorf("second recv error, err: %w", err)
	}

	if !s.p.ParseLastRecv(bytes) {
		s.logger.Errorf("解析末次报文错误: %v", utils.ToHex(bytes, " "))
	}

	s.lastRecvBuff = s.p.recvData

	if needLogging {
		s.logger.Infof("recv: %v", utils.ToHex(bytes, " "))
	}

	return nil
}

// handle 根据命令码分发到对应的处理函数
func (s *DoorServer) handle() error {
	switch s.p.cmd {
	case consts.CommandSetControllerParams:
		return s.handleSetControllerParams()
	case consts.CommandOpenDoor:
		return s.handleOpenDoor()
	case consts.CommandDoorOpenPermenently:
		return s.handleDoorOpenPermenently()
	case consts.CommandCloseDoor:
		return s.handleCloseDoor()
	case consts.CommandLockDoor:
		return s.handleLockDoor()
	case consts.CommandSetTime:
		return s.handleSetTime()
	case consts.CommandUploadAlarms:
		return s.handleUploadAlarms()
	case consts.CommandUploadControllerStatus:
		return s.handleUploadControllerStatus()
	case consts.CommandUploadEvents:
		return s.handleUploadEvents()
	case consts.CommandSetDoorParams:
		return s.handleSetDoorParams()
	case consts.CommandClearTimeGroups:
		return s.handleClearDoorTimeGroups()
	case consts.CommandAddTimeGroup:
		return s.handleAddTimeGroup()
	case consts.CommandClearCards:
		return s.handleClearCards()
	case consts.CommandClean:
		return s.handleReset()
	case consts.CommandSetAlarm:
		return s.handleSetAlarm()
	case consts.CommandSetFireAlarm:
		return s.handleSetFireAlarm()
	case consts.CommandDeleteCard:
		return s.handleDeleteCard()
	case consts.CommandAddCard:
		return s.handleAddCard()
	default:
		return nil
	}
}

// handleUploadControllerStatus 处理控制器状态上传命令
func (s *DoorServer) handleUploadControllerStatus() error {
	if len(s.lastRecvBuff) < xbrother.GetFieldSizeSum(xbrother.ControllerStatusUploadReq{}) {
		return fmt.Errorf("报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), xbrother.GetFieldSizeSum(xbrother.ControllerStatusUploadReq{}))
	}
	req, err := marshaller.ControllerStatusUploadReqUnMarshal(s.lastRecvBuff)
	if err != nil {
		s.logger.Errorf("ControllerStatusUploadReqUnMarshal error: %v", err)
		return s.sendPacket(consts.CommandUploadControllerStatus, s.p.doorNo,
			marshaller.CommonRespMarshal(xbrother.CommonResp{Rtn: consts.NAK}))
	}

	s.logger.Debugf("接收到控制器上传记录: %+v", req)

	if err = s.saveControllerStatus(req); err != nil {
		s.logger.Errorf(
			"save controller status error, err: %v", err)
		nakResp := marshaller.CommonRespMarshal(
			xbrother.CommonResp{Rtn: consts.NAK})
		return s.sendPacket(
			consts.CommandUploadControllerStatus,
			s.p.doorNo, nakResp)
	}
	okResp := marshaller.CommonRespMarshal(
		xbrother.CommonResp{Rtn: 0})
	return s.sendPacket(
		consts.CommandUploadControllerStatus,
		s.p.doorNo, okResp)
}

// saveControllerStatus 通过RRPC机制保存控制器状态
func (s *DoorServer) saveControllerStatus(req xbrother.ControllerStatusUploadReq) error {
	s.controllerStatusChan <- req
	iErr, ok := rrpc.Manager().Get(consts.GetRRPCSetDoorStatus(s.channelID), s.timeout)
	if !ok {
		return fmt.Errorf("rrpc get save controller status error")
	}
	if iErr != nil {
		err, ok := iErr.(error)
		if !ok {
			return fmt.Errorf("unexpected rrpc get result type, expect error")
		}
		return err
	}
	return nil
}

// SendData 通过TCP连接发送数据（线程安全）
func (s *DoorServer) SendData(data []byte) error {
	var sendLen int
	var err error
	func() {
		s.fdMutex.Lock()
		defer s.fdMutex.Unlock()
		sendLen, err = dtcp.WriteN(s.conn, data, s.timeout)
	}()

	if err != nil {
		dtcp.FlushAndClose(s.conn, s.timeout)
		return err
	}
	if sendLen != len(data) {
		dtcp.FlushAndClose(s.conn, s.timeout)
		return fmt.Errorf("发送注册响应报文失败 bytesSend: %d, try to close", sendLen)
	}

	if config.C.IsLoggingPacket(s.channelID) {
		s.logger.Infof("send: %s", utils.ToHex(data, " "))
	}

	return nil
}

// handleUploadEvents 处理事件上传命令
func (s *DoorServer) handleUploadEvents() error {
	if len(s.lastRecvBuff) < xbrother.GetFieldSizeSum(xbrother.EventUploadReq{}) {
		return fmt.Errorf("报文长度错误: %d, 期望: %d", len(s.lastRecvBuff), xbrother.GetFieldSizeSum(xbrother.EventUploadReq{}))
	}
	req, err := marshaller.EventUploadReqUnMarshal(s.lastRecvBuff)
	if err != nil {
		s.logger.Errorf("EventUploadReqUnMarshal err, err: %s", err.Error())
		return err
	}

	s.logger.Debugf("recv event: %+v", req)

	if err = s.saveEvent(req); err != nil {
		s.logger.Errorf("save event in db error, err: %v", err)
		return s.sendPacket(
			consts.CommandUploadEvents, req.Door,
			marshaller.EventUploadRespMarshal(
				xbrother.EventUploadResp{Seq: 0}))
	}
	return s.sendPacket(
		consts.CommandUploadEvents, req.Door,
		marshaller.EventUploadRespMarshal(
			xbrother.EventUploadResp{Seq: req.Seq}))
}

// saveAlarm 通过RRPC机制保存告警记录
func (s *DoorServer) saveAlarm(req xbrother.AlarmUploadReq) error {
	s.alarmChan <- req
	iErr, ok := rrpc.Manager().Get(consts.GetRRPCSetDriverAlarm(s.channelID), s.timeout)
	if !ok {
		return fmt.Errorf("rrpc get save alarm error")
	}
	if iErr != nil {
		err, ok := iErr.(error)
		if !ok {
			return fmt.Errorf("unexpected rrpc get result type, expect error")
		}
		return err
	}
	return nil
}

// handleUploadAlarms 处理告警上传命令
func (s *DoorServer) handleUploadAlarms() error {
	expectedSize := xbrother.GetFieldSizeSum(xbrother.AlarmUploadReq{})
	if len(s.lastRecvBuff) < expectedSize {
		return fmt.Errorf(
			"报文长度错误: %d, 期望: %d",
			len(s.lastRecvBuff), expectedSize)
	}
	req, err := marshaller.AlarmUploadReqUnMarshal(s.lastRecvBuff)
	if err != nil {
		s.logger.Errorf(
			"AlarmUploadReqUnMarshal err, err: %s", err.Error())
		return s.sendPacket(
			consts.CommandUploadAlarms, req.Door,
			marshaller.AlarmUploadRespMarshal(
				xbrother.AlarmUploadResp{Seq: 0}))
	}

	s.logger.Debugf("recv alarm: %+v", req)

	if err = s.saveAlarm(req); err != nil {
		s.logger.Errorf(
			"save alarm in db error, err: %s", err.Error())
		return s.sendPacket(
			consts.CommandUploadAlarms, req.Door,
			marshaller.AlarmUploadRespMarshal(
				xbrother.AlarmUploadResp{Seq: 0}))
	}
	return s.sendPacket(
		consts.CommandUploadAlarms, req.Door,
		marshaller.AlarmUploadRespMarshal(
			xbrother.AlarmUploadResp{Seq: req.Seq}))
}

// sendPacket 构建并发送响应报文
func (s *DoorServer) sendPacket(cmd uint8, door uint8, data []byte) error {
	s.p.BuildPacket(cmd, door, data)
	if err := s.SendData(s.p.SendData()); err != nil {
		return fmt.Errorf("send packet error, cmd: %d, door: %d: %s", cmd, door, err.Error())
	}
	return nil
}

// handleCommonResp 处理通用RRPC响应（将接收到的数据设置到RRPC管理器）
func (s *DoorServer) handleCommonResp(rrpcKey string) error {
	if len(s.lastRecvBuff) < xbrother.GetFieldSizeSum(xbrother.CommonResp{}) {
		return fmt.Errorf("报文长度错误: %d, 期望: %d", len(s.lastRecvBuff), xbrother.GetFieldSizeSum(xbrother.CommonResp{}))
	}
	rrpc.Manager().Set(rrpcKey, s.lastRecvBuff)
	return nil
}

// handleAddCard 处理添加卡响应
func (s *DoorServer) handleAddCard() error {
	return s.handleCommonResp(consts.GetRRPCAddCard(s.channelID))
}

// handleDeleteCard 处理删除卡响应
func (s *DoorServer) handleDeleteCard() error {
	return s.handleCommonResp(consts.GetRRPCDeleteCard(s.channelID))
}

// handleLockDoor 处理锁门响应
func (s *DoorServer) handleLockDoor() error {
	return s.handleCommonResp(consts.GetRRPCLockDoor(s.channelID))
}

// handleSetFireAlarm 处理设置火警响应
func (s *DoorServer) handleSetFireAlarm() error {
	return s.handleCommonResp(consts.GetRRPCSetFireAlarm(s.channelID))
}

// handleSetAlarm 处理设置告警响应
func (s *DoorServer) handleSetAlarm() error {
	return s.handleCommonResp(consts.GetRRPCSetAlarm(s.channelID))
}

// handleReset 处理重置控制器响应
func (s *DoorServer) handleReset() error {
	return s.handleCommonResp(consts.GetRRPCClean(s.channelID))
}

// handleClearCards 处理清空卡响应
func (s *DoorServer) handleClearCards() error {
	return s.handleCommonResp(consts.GetRRPCClearCards(s.channelID))
}

// handleAddTimeGroup 处理添加时间组响应
func (s *DoorServer) handleAddTimeGroup() error {
	return s.handleCommonResp(consts.GetRRPCAddTimeGroup(s.channelID))
}

// handleClearDoorTimeGroups 处理清空门时间组响应
func (s *DoorServer) handleClearDoorTimeGroups() error {
	return s.handleCommonResp(consts.GetRRPCClearDoorTimeGroups(s.channelID))
}

// handleSetDoorParams 处理设置门参数响应
func (s *DoorServer) handleSetDoorParams() error {
	return s.handleCommonResp(consts.GetRRPCSetDoorParams(s.channelID))
}

// handleSetTime 处理设置时间响应
func (s *DoorServer) handleSetTime() error {
	return s.handleCommonResp(consts.GetRRPCSetTime(s.channelID))
}

// handleCloseDoor 处理关门响应
func (s *DoorServer) handleCloseDoor() error {
	return s.handleCommonResp(consts.GetRRPCCloseDoor(s.channelID))
}

// handleDoorOpenPermenently 处理常开门响应
func (s *DoorServer) handleDoorOpenPermenently() error {
	return s.handleCommonResp(consts.GetRRPCDoorOpenPermanently(s.channelID))
}

// handleOpenDoor 处理开门响应
func (s *DoorServer) handleOpenDoor() error {
	return s.handleCommonResp(consts.GetRRPCOpenDoor(s.channelID))
}

// handleSetControllerParams 处理设置控制器参数响应
func (s *DoorServer) handleSetControllerParams() error {
	return s.handleCommonResp(consts.GetRRPCSetControllerParams(s.channelID))
}

// HandleConn 处理TCP连接的主循环，持续接收和处理数据直到连接关闭
func HandleConn(ctx context.Context, s *DoorServer) {
	s.logger.Infof("connection: local addr: %s, remote addr:%s",
		s.conn.LocalAddr().String(), s.conn.RemoteAddr().String())
	var exitFlag = false
	for !exitFlag {
		select {
		case <-ctx.Done():
			s.logger.Infof("stop server handler")
			return
		default:
			// 处理数据
			if err := s.serverLoop(); err != nil {
				exitFlag = true
			}
			break
		}
	}
	s.Close()
}

// Request 构建并发送请求报文到控制器
func (s *DoorServer) Request(cmd uint8, doorNo uint8, data []byte) error {
	var packet Packet
	packet.BuildPacket(cmd, doorNo, data)
	sendBuff := packet.SendData()
	return s.SendData(sendBuff)
}
