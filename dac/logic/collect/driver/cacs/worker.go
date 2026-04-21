package cacs

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/model/driver/cacs"
	"dac/entity/model/marshaller"
	"dac/entity/utils"
	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/cacs/consts"
	"dac/logic/dlm"
)

var worker = NewDoorWorker(consts.KPassiveReceivePort)

// 限流相关常量
const (
	minRejectInterval = 500 * time.Millisecond // 最小拒绝间隔
	maxRejectCount    = 5                      // 连续拒绝次数阈值
	cooldownDuration  = 10 * time.Second       // 冷却时间
)

// rejectInfo 记录被拒绝连接的信息
type rejectInfo struct {
	count      int       // 连续被拒绝次数
	lastReject time.Time // 上次被拒绝时间
	cooldownAt time.Time // 冷却期结束时间
}

type DoorWorker struct {
	port      int
	listenIP  string
	listener  *net.Listener
	timeoutMs time.Duration
	mutex     sync.Mutex
	IsFirst   bool
	IsStart   bool
	// MAC 地址白名单：key 为 MAC 地址（大写无分隔符），value 为 Controller
	macAddrMap      map[string]*Controller
	macAddrMapMutex sync.RWMutex
	// 待校验连接：key 为远程地址（IP:Port），value 为 DoorServer（等待注册包校验 MAC）
	pendingServerMap      map[string]*DoorServer
	pendingServerMapMutex sync.RWMutex
	// 已校验连接：key 为 MAC 地址，value 为 DoorServer
	serverMap      map[string]*DoorServer
	serverMapMutex sync.RWMutex
	// 限流记录：key 为 MAC 地址，value 为拒绝信息
	rejectInfoMap      map[string]*rejectInfo
	rejectInfoMapMutex sync.RWMutex
}

func NewDoorWorker(port int) *DoorWorker {
	w := &DoorWorker{
		port:             port,
		listenIP:         "0.0.0.0",
		timeoutMs:        consts.KDefaultTimeoutMs,
		IsFirst:          true,
		macAddrMap:       make(map[string]*Controller),
		pendingServerMap: make(map[string]*DoorServer),
		serverMap:        make(map[string]*DoorServer),
		rejectInfoMap:    make(map[string]*rejectInfo),
	}
	return w
}

func (w *DoorWorker) Start() bool {
	if w.IsStart {
		return true
	}
	config.Log.Infof("开始监听：" + w.listenIP + ":" + strconv.Itoa(w.port))
	listener, err := net.Listen("tcp4", w.listenIP+":"+strconv.Itoa(w.port))
	if err != nil {
		config.Log.Infof("监听失败：%v", err)
		return false
	}

	w.listener = &listener
	go w.Loop()
	w.IsStart = true
	return true
}

func (w *DoorWorker) Loop() {
	for {
		conn, err := (*w.listener).Accept()
		if err != nil {
			continue
		}
		remoteAddr := conn.RemoteAddr().String()

		// 使用 remoteAddr（IP:Port）作为 key，避免 K8s SNAT 场景下不同设备互踢
		// 关闭之前的旧连接（如果存在于 pendingServerMap）
		func() {
			w.pendingServerMapMutex.Lock()
			defer w.pendingServerMapMutex.Unlock()
			if tmpServer, ok := w.pendingServerMap[remoteAddr]; ok {
				tmpServer.Close()
				delete(w.pendingServerMap, remoteAddr)
			}
		}()

		// 创建 DoorServer，先放入 pendingServerMap，等待注册包校验 MAC
		// serverName 使用 remoteAddr 以便后续移除
		server := NewDoorServer(w.timeoutMs, conn, remoteAddr, remoteAddr, 0)
		func() {
			w.pendingServerMapMutex.Lock()
			defer w.pendingServerMapMutex.Unlock()
			w.pendingServerMap[remoteAddr] = server
		}()

		config.Log.Infof("新连接接入，等待注册包校验 MAC，remoteAddr: %s", remoteAddr)

		// 启动连接处理协程，等待注册包
		go HandleConnWithMACVerify(server, w)
	}
}

// HandleConnWithMACVerify 处理连接，带 MAC 校验
func HandleConnWithMACVerify(s *DoorServer, w *DoorWorker) {
	config.Log.Infof("New connect from %s, waiting for register packet", s.conn.RemoteAddr())

	// 等待第一个数据包（应该是注册包）
	ret, err := s.recvData()
	if ret < 0 {
		if err != nil {
			config.Log.Infof("recvData err: %v, close connection", err)
		}
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	// 检查是否为注册包
	if s.p.RecvCommand() != consts.KCommandRequestRegister {
		config.Log.Infof("第一个包不是注册包，拒绝连接，remoteIP: %s", s.serverName)
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	// 解析注册包，获取 MAC 地址
	if len(s.lastRecvBuff) != consts.KRegisterReqLen {
		config.Log.Infof("注册报文长度错误: %d, 期望: %d", len(s.lastRecvBuff), consts.KRegisterReqLen)
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	req, err := marshaller.RequestControllerRegisterUnMarshal(s.lastRecvBuff)
	if err != nil {
		config.Log.Infof("RequestControllerRegisterUnMarshal error: %v", err)
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	// 获取 MAC 地址（大写无分隔符）
	macAddr := strings.ToUpper(utils.ToHex(req.MAC[:], ""))
	config.Log.Infof("收到注册包，MAC 地址: %s，序列号: %d，名称: %s", macAddr, req.Seq, string(req.Name[:]))

	// 校验 MAC 地址白名单
	var controller *Controller
	var found bool
	func() {
		w.macAddrMapMutex.RLock()
		defer w.macAddrMapMutex.RUnlock()
		controller, found = w.macAddrMap[macAddr]
	}()

	if !found {
		// 检查是否在冷却期内
		if w.isInCooldown(macAddr) {
			// 静默丢弃，不打印日志（避免日志风暴）
			s.Close()
			w.removePendingServer(s.serverName)
			return
		}
		config.Log.Infof("MAC 地址不在白名单中，拒绝连接，MAC: %s, remoteAddr: %s", macAddr, s.serverName)
		// 记录拒绝信息，可能触发冷却期
		w.recordReject(macAddr)
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	// MAC 校验通过，清除该 MAC 的拒绝记录
	w.clearRejectInfo(macAddr)

	config.Log.Infof("MAC 地址校验通过，MAC: %s, remoteAddr: %s", macAddr, s.serverName)

	// 检查分布式锁：只有持有锁的 Pod 才处理门控器连接
	// 这样可以避免多 Pod 同时处理导致 Redis 压力过大
	if !dlm.GetWorker().HasLock() {
		config.Log.Infof("当前 Pod 未持有分布式锁，拒绝连接，MAC: %s", macAddr)
		s.Close()
		w.removePendingServer(s.serverName)
		return
	}

	// 更新 server 的 controllerID
	s.controllerID = controller.baseInfo.ID
	s.channelID = macAddr

	// 为这个连接创建独立的 context，用于控制协程退出
	serverCtx, serverCancel := context.WithCancel(controller.ctx)
	s.ctx = serverCtx
	s.cancel = serverCancel

	// 关闭该 MAC 地址之前的旧连接（如果存在，处理门控器重连场景）
	func() {
		w.serverMapMutex.Lock()
		defer w.serverMapMutex.Unlock()
		if oldServer, ok := w.serverMap[macAddr]; ok {
			config.Log.Infof("门控器重连，关闭旧连接及其协程，MAC: %s", macAddr)
			// 先取消旧连接的 context，让协程退出
			if oldServer.cancel != nil {
				oldServer.cancel()
			}
			// 再关闭 TCP 连接
			oldServer.Close()
		}
		w.serverMap[macAddr] = s
	}()

	// 从 pendingServerMap 移除
	w.removePendingServer(s.serverName)

	// 完成注册包处理
	s.controllerSeq = req.Seq
	s.controllerMACAddr = req.MAC
	s.controllerName = string(req.Name[:])
	s.eventAlarmSeq = req.EventAlarmSeq

	// 发送注册响应
	data := marshaller.ResponseControllerRegisterMarshal(cacs.ControllerRegisterResp{Id: s.controllerSeq})
	s.p.BuildResponsePacket(consts.KCommandResponseRegister, 0, data)
	sendBuff := s.p.SendData()
	var sendLen int
	func() {
		s.fdMutex.Lock()
		defer s.fdMutex.Unlock()
		sendLen, err = dtcp.WriteN(s.conn, sendBuff, s.timeoutMS)
	}()
	if sendLen != len(sendBuff) {
		config.Log.Infof("发送注册响应报文失败, bytesSend: %d", sendLen)
		s.Close()
		serverCancel()
		w.removeServer(macAddr)
		return
	}
	s.isRegistered = true

	// 重连时重置 Controller 状态（不创建新协程）
	controller.resetForReconnect()

	// 通知 Controller 连接已建立
	select {
	case controller.connectChan <- struct{}{}:
	default:
		// channel 满了，忽略（说明上一次通知还没被处理）
		config.Log.Infof("connectChan 已满，跳过通知")
	}

	// 启动门状态接收（使用 Server 独立的 ctx）
	go s.recvDoorStatus(serverCtx)

	// 启动事件告警保存（使用 Server 独立的 ctx）
	go s.saveEventAlarm(serverCtx)

	// 继续处理后续数据包
	for {
		select {
		case <-serverCtx.Done():
			config.Log.Infof("Server context 已取消，退出主循环，MAC: %s", macAddr)
			s.Close()
			w.removeServer(macAddr)
			return
		default:
			// 检查是否仍持有分布式锁，如果丢失锁则主动断开连接
			// 这样可以让持有锁的 Pod 接管连接
			if !dlm.GetWorker().HasLock() {
				config.Log.Infof("当前 Pod 丢失分布式锁，主动断开连接，MAC: %s", macAddr)
				serverCancel()
				s.Close()
				w.removeServer(macAddr)
				return
			}

			if s.passiveListenServerLoop() < 0 {
				config.Log.Infof("passiveListenServerLoop 返回错误，退出，MAC: %s", macAddr)
				serverCancel() // 取消 context，让 recvDoorStatus 和 saveEventAlarm 退出
				s.Close()
				w.removeServer(macAddr)
				return
			}
		}
	}
}

// removePendingServer 从 pendingServerMap 中移除
func (w *DoorWorker) removePendingServer(remoteAddr string) {
	w.pendingServerMapMutex.Lock()
	defer w.pendingServerMapMutex.Unlock()
	delete(w.pendingServerMap, remoteAddr)
}

// removeServer 从 serverMap 中移除
func (w *DoorWorker) removeServer(macAddr string) {
	w.serverMapMutex.Lock()
	defer w.serverMapMutex.Unlock()
	delete(w.serverMap, macAddr)
}

// GetServerByMAC 根据 MAC 地址获取 DoorServer
func (w *DoorWorker) GetServerByMAC(macAddr string) (*DoorServer, bool) {
	w.serverMapMutex.RLock()
	defer w.serverMapMutex.RUnlock()
	server, ok := w.serverMap[strings.ToUpper(macAddr)]
	return server, ok
}

// isInCooldown 检查指定 MAC 是否在冷却期内
func (w *DoorWorker) isInCooldown(macAddr string) bool {
	w.rejectInfoMapMutex.RLock()
	defer w.rejectInfoMapMutex.RUnlock()
	if info, ok := w.rejectInfoMap[macAddr]; ok {
		if time.Now().Before(info.cooldownAt) {
			return true
		}
	}
	return false
}

// recordReject 记录一次拒绝，达到阈值后进入冷却期
func (w *DoorWorker) recordReject(macAddr string) {
	w.rejectInfoMapMutex.Lock()
	defer w.rejectInfoMapMutex.Unlock()

	now := time.Now()
	info, ok := w.rejectInfoMap[macAddr]
	if !ok {
		info = &rejectInfo{}
		w.rejectInfoMap[macAddr] = info
	}

	// 如果距离上次拒绝时间太短，增加计数
	if now.Sub(info.lastReject) < minRejectInterval {
		info.count++
	} else {
		// 间隔较长，重置计数
		info.count = 1
	}
	info.lastReject = now

	// 达到阈值，进入冷却期
	if info.count >= maxRejectCount {
		info.cooldownAt = now.Add(cooldownDuration)
		config.Log.Infof("MAC %s 连续被拒绝 %d 次，进入 %v 冷却期", macAddr, info.count, cooldownDuration)
		info.count = 0 // 重置计数，冷却期结束后重新统计
	}
}

// clearRejectInfo 清除指定 MAC 的拒绝记录（MAC 校验通过时调用）
func (w *DoorWorker) clearRejectInfo(macAddr string) {
	w.rejectInfoMapMutex.Lock()
	defer w.rejectInfoMapMutex.Unlock()
	delete(w.rejectInfoMap, macAddr)
}
