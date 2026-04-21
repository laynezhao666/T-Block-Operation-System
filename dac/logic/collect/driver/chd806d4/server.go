// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"dac/entity/config"
	consts2 "dac/logic/collect/driver/chd806d4/consts"

	"dac/entity/utils/rrpc"
	"dac/entity/utils/tlog"
)

// DoorServer CHD 门控器服务器（负责与单个门控器的通信）
type DoorServer struct {
	conn         net.Conn
	channelID    string        // 门控器地址（格式：IP:Port，如 "192.168.1.100:8002"）
	timeout      time.Duration // 超时时间
	seqNo        uint32        // 命令序号（自增）
	sendMutex    sync.Mutex    // 发送锁
	recvMutex    sync.Mutex    // 接收锁
	lastRecvTime time.Time     // 最后接收时间
	isConnected  bool          // 连接状态
	ctx          context.Context
	cancel       context.CancelFunc

	// 事件和告警通道（用于主动上报）
	eventChan chan []byte
	alarmChan chan []byte

	// 断线通知通道（用于重连机制）
	disConnectChan chan struct{}

	// 请求映射表：SeqNo -> RRPC Key
	// 用于将响应包的 SeqNo 映射到对应的 RRPC Key
	requestMap sync.Map // map[uint8]string

	// 日志记录器
	logger tlog.Logger
}

// NewDoorServer 创建新的门控器服务器
func NewDoorServer(channelID string, timeout time.Duration) *DoorServer {
	logger := tlog.NewPrefixLogger(fmt.Sprintf("[CHD@%v]", channelID), config.Log)
	ctx, cancel := context.WithCancel(context.Background())
	return &DoorServer{
		channelID:      channelID,
		timeout:        timeout,
		seqNo:          1,
		lastRecvTime:   time.Now(),
		ctx:            ctx,
		cancel:         cancel,
		eventChan:      make(chan []byte, 100),
		alarmChan:      make(chan []byte, 100),
		disConnectChan: make(chan struct{}, 1),
		logger:         logger,
	}
}

// Connect 连接到门控器
func (s *DoorServer) Connect(ctx context.Context) error {
	s.logger.Infof("正在连接门控器: %s", s.channelID)

	dialer := &net.Dialer{Timeout: s.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", s.channelID)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	s.conn = conn
	s.isConnected = true
	s.logger.Infof("门控器连接成功: %s", s.channelID)

	// 启动接收协程
	go s.recvLoop(ctx)

	return nil
}

// Disconnect 断开连接
func (s *DoorServer) Disconnect() {
	// 先设置标志位，让 recvLoop 知道这是正常关闭
	s.isConnected = false

	// 关闭连接
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Infof("门控器连接已关闭: %s", s.channelID)
}

// IsConnected 检查连接状态
func (s *DoorServer) IsConnected() bool {
	return s.isConnected
}

// Request 发送请求并等待响应
func (s *DoorServer) Request(cid, groupCode, cmdType uint8, data []byte, rrpcKeyFunc func(string, uint8) string) ([]byte, error) {
	if !s.isConnected {
		return nil, fmt.Errorf("门控器未连接")
	}

	// 1. 构建 INFO 部分
	info := make([]byte, 0, 2+len(data))
	info = append(info, groupCode, cmdType)
	info = append(info, data...)

	needLogging := config.C.IsLoggingPacket(s.channelID)
	if needLogging {
		s.logger.Infof("INFO内容: % X", info)
	}

	// 2. 创建数据包并生成 SeqNo
	seqNo := s.getNextSeqNo()
	packet := NewPacket(seqNo, consts2.DefaultADR1, consts2.DefaultADR2, cid, info)

	// 3. 使用传入的函数生成 RRPC Key（包含 seqNo）
	rrpcKey := rrpcKeyFunc(s.channelID, seqNo)

	// 4. 存储 SeqNo -> RRPC Key 的映射关系
	s.requestMap.Store(seqNo, rrpcKey)
	defer s.requestMap.Delete(seqNo) // 请求完成后删除映射

	if needLogging {
		s.logger.Debugf("存储请求映射: SeqNo=%d, RRPCKey=%s", seqNo, rrpcKey)
	}

	// 4. 发送数据包
	s.sendMutex.Lock()
	packedData := packet.Pack()
	_, err := s.conn.Write(packedData)
	s.sendMutex.Unlock()

	if err != nil {
		return nil, fmt.Errorf("发送失败: %v", err)
	}

	// 打印完整的发送数据包（仅在配置开启时）
	if needLogging {
		s.logger.Infof("send: % X", packedData)
	}

	// 5. 等待响应（通过 RRPC）
	respRaw, ok := rrpc.Manager().Get(rrpcKey, s.timeout)
	if !ok {
		return nil, fmt.Errorf("等待响应超时 (SeqNo=%d, RRPCKey=%s, Timeout=%v)", seqNo, rrpcKey, s.timeout)
	}

	respPacket, ok := respRaw.(*Packet)
	if !ok {
		return nil, fmt.Errorf("响应数据类型错误")
	}

	// 6. 检查返回码
	if respPacket.CID != consts2.RTNSuccess {
		return nil, fmt.Errorf("门控器返回错误: RTN=0x%02X (%s)",
			respPacket.CID, consts2.GetRTNMessage(respPacket.CID))
	}

	return respPacket.INFO, nil
}

// recvLoop 接收循环
func (s *DoorServer) recvLoop(ctx context.Context) {
	buffer := make([]byte, 4096)
	tempBuf := make([]byte, 0, 4096)
	needLogging := config.C.IsLoggingPacket(s.channelID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 设置读取超时
		s.conn.SetReadDeadline(time.Now().Add(90 * time.Second))

		n, err := s.conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 超时，继续等待
				continue
			}
			// 连接断开
			// 区分正常关闭和异常断开
			if !s.isConnected {
				// 正常关闭（Close() 已经设置 isConnected = false）
				s.logger.Infof("连接已正常关闭: %s", s.channelID)
			} else {
				// 异常断开
				s.logger.Warnf("连接异常断开: %v", err)
				s.isConnected = false
				// 通知上层进行重连
				select {
				case s.disConnectChan <- struct{}{}:
				default:
				}
			}
			return
		}

		if n > 0 {
			tempBuf = append(tempBuf, buffer[:n]...)
			s.lastRecvTime = time.Now()

			// 打印接收到的原始数据（仅在配置开启时）
			if needLogging {
				s.logger.Infof("recv: % X", buffer[:n])
			}
			// 尝试解析数据包
			for len(tempBuf) > 0 {
				packet, consumed, err := s.tryParsePacket(tempBuf)

				// 如果有消费字节，移除已处理的数据
				if consumed > 0 {
					tempBuf = tempBuf[consumed:]
				}

				// 如果有错误，打印日志（已经丢弃了错误数据）
				if err != nil {
					s.logger.Warnf("%v", err)
					continue
				}

				// 如果没有解析出数据包，说明数据不完整，等待更多数据
				if packet == nil {
					break
				}

				// 处理数据包
				s.handlePacket(packet)
			}
		}
	}
}

// tryParsePacket 尝试解析数据包
// 返回值：(数据包, 消费的字节数, 错误)
// - 如果成功解析：返回 (packet, consumed, nil)
// - 如果数据不完整：返回 (nil, 0, nil)
// - 如果解析错误：返回 (nil, consumed, error) - 消费错误数据
func (s *DoorServer) tryParsePacket(data []byte) (*Packet, int, error) {
	if len(data) == 0 {
		return nil, 0, nil
	}

	// 查找起始符
	startIdx := -1
	for i := 0; i < len(data); i++ {
		if data[i] == consts2.SOI {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		// 没有找到起始符，丢弃所有数据
		return nil, len(data), fmt.Errorf("未找到起始符，丢弃 %d 字节", len(data))
	}

	// 如果起始符不在开头，先丢弃前面的垃圾数据
	if startIdx > 0 {
		s.logger.Debugf("丢弃 %d 字节垃圾数据", startIdx)
		return nil, startIdx, nil
	}

	// 查找结束符
	endIdx := -1
	for i := startIdx + 1; i < len(data); i++ {
		if data[i] == consts2.EOI {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		// 数据不完整，等待更多数据
		return nil, 0, nil
	}

	// 提取完整的数据包
	packetData := data[startIdx : endIdx+1]
	packet, err := Unpack(packetData)
	if err != nil {
		// 解析失败，丢弃这个数据包（从起始符到结束符）
		s.logger.Warnf("数据包解析失败: %v，丢弃 %d 字节", err, endIdx+1)
		return nil, endIdx + 1, nil
	}

	return packet, endIdx + 1, nil
}

// handlePacket 处理接收到的数据包
func (s *DoorServer) handlePacket(packet *Packet) {
	needLogging := config.C.IsLoggingPacket(s.channelID)

	// 判断数据包类型
	// 主动上报的 CID 范围是 0x80-0x8F（高4位为0x8，低4位为重复次数）
	isUpload := packet.CID >= 0x80 && packet.CID <= 0x8F

	// 打印详细的数据包信息（仅在配置开启时）
	if needLogging {
		s.logger.Infof("收到数据包: SeqNo=0x%02X, CID=0x%02X, LTH=%d, INFO=[%d bytes]% X",
			packet.SeqNo, packet.CID, packet.LTH, len(packet.INFO), packet.INFO)
	}

	// 判断数据包类型
	if isUpload {
		// 主动上报（触发包或心跳包）
		s.handleUpload(packet)
	} else {
		// 命令响应（包括成功响应和错误响应）
		s.handleResponse(packet)
	}
}

// handleResponse 处理命令响应
func (s *DoorServer) handleResponse(packet *Packet) {
	needLogging := config.C.IsLoggingPacket(s.channelID)

	// 从请求映射表中获取 RRPC Key
	rrpcKeyRaw, ok := s.requestMap.Load(packet.SeqNo)
	if !ok {
		s.logger.Warnf("收到未知 SeqNo 的响应: SeqNo=%d, CID=0x%02X", packet.SeqNo, packet.CID)
		return
	}

	rrpcKey, ok := rrpcKeyRaw.(string)
	if !ok {
		s.logger.Errorf("RRPC Key 类型错误: SeqNo=%d", packet.SeqNo)
		return
	}

	if needLogging {
		s.logger.Debugf("设置 RRPC 响应: Key=%s, SeqNo=%d", rrpcKey, packet.SeqNo)
	}

	rrpc.Manager().Set(rrpcKey, packet)
}

// handleUpload 处理主动上报
func (s *DoorServer) handleUpload(packet *Packet) {
	repeatCount := packet.CID & 0x0F // 低4位是重复次数

	// 根据协议文档：
	// INFO[0] = 提供服务的 HOST 标识（通常为 0）
	// INFO[1] = 请求服务的种类：0=心跳包，1=触发包
	if len(packet.INFO) >= 2 {
		serviceType := packet.INFO[1] // 服务种类

		// 心跳包：服务种类 = 0
		if serviceType == 0 {
			s.logger.Debugf("收到心跳包")
			return
		}

		// 触发包：服务种类 = 1
		if serviceType == 1 {
			s.logger.Infof("收到触发包 (重复次数=%d)，需要读取事件", repeatCount)

			// 通知 Controller 读取事件
			select {
			case s.eventChan <- packet.INFO:
				s.logger.Debugf("触发包已加入事件队列")
			default:
				s.logger.Warnf("事件通道已满，丢弃触发包")
			}
			return
		}
	}

	// 未知类型的主动上报
	s.logger.Warnf("收到未知类型的主动上报: INFO=% X", packet.INFO)
}

// getNextSeqNo 获取下一个序号
func (s *DoorServer) getNextSeqNo() uint8 {
	seqNo := atomic.AddUint32(&s.seqNo, 1)
	return uint8(seqNo & 0xFF)
}

// getCIDName 获取CID命令名称
func getCIDName(cid uint8) string {
	switch cid {
	case 0x48:
		return "权限认证相关"
	case 0x49:
		return "参数设置"
	case 0x4A:
		return "读取信息"
	case 0x20:
		return "读取时间"
	case 0x21:
		return "设置时间"
	case 0x30:
		return "开门"
	case 0x31:
		return "关门"
	case 0x40:
		return "读取事件"
	default:
		return fmt.Sprintf("未知命令(0x%02X)", cid)
	}
}

// GetEventChan 获取事件通道
func (s *DoorServer) GetEventChan() <-chan []byte {
	return s.eventChan
}

// GetAlarmChan 获取告警通道
func (s *DoorServer) GetAlarmChan() <-chan []byte {
	return s.alarmChan
}
