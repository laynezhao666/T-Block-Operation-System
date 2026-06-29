package ping

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"agent/entity/consts"
	"agent/entity/definition"
	model2 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/osal"

	elog "etrpc-go/log"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ICMP相关常量
const (
	maxPacketSize  = 4096  // 最大包大小
	defaultDataLen = 64    // 默认数据长度
	icmpHeaderLen  = 8     // ICMP头部长度
	defaultTimeout = 1000  // 默认超时时间（毫秒）
	maxWaitMs      = 10000 // 最大等待时间（毫秒）
)

// PingDevice 实现 IDevice 接口的Ping设备
type PingDevice struct {
	gid  definition.DeviceGidType
	name string

	ip        string        // 目标IP地址
	dataLen   int           // ICMP数据长度
	timeoutMs int           // 超时时间（毫秒）
	id        uint16        // ICMP标识符
	seq       uint16        // ICMP序列号
	timeout   time.Duration // 超时时间
}

// NewPingDevice 创建一个新的Ping设备实例
func NewPingDevice(gid definition.DeviceGidType, name string) *PingDevice {
	return &PingDevice{
		gid:       gid,
		name:      name,
		dataLen:   defaultDataLen,
		timeoutMs: defaultTimeout,
		id:        uint16(rand.Intn(65535)),
		seq:       0,
	}
}

// Open 打开通道，解析通道信息
func (d *PingDevice) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	// 解析参数：params格式为 "timeout_ms;data_len"
	params := strings.Split(chanInfo.Params, ";")

	// 解析超时时间
	if len(params) > 0 && params[0] != "" {
		if t, err := strconv.Atoi(params[0]); err == nil && t > 0 && t < maxWaitMs {
			d.timeoutMs = t
		}
	}

	// 解析数据长度
	if len(params) > 1 && params[1] != "" {
		if l, err := strconv.Atoi(params[1]); err == nil && l > 0 && l < maxPacketSize {
			d.dataLen = l
		}
	}

	// 解析IP地址（格式可能是 "ip:port" 或 "ip"）
	d.ip = chanInfo.Name
	if pos := strings.Index(chanInfo.Name, ":"); pos != -1 {
		d.ip = chanInfo.Name[:pos]
	}

	// 验证IP地址格式
	if net.ParseIP(d.ip) == nil {
		return consts.QualityConfigError
	}

	d.timeout = time.Duration(d.timeoutMs) * time.Millisecond

	return consts.QualityOk
}

// Close 关闭通道
func (d *PingDevice) Close() consts.Quality {
	return consts.QualityOk
}

// Request 发送采集指令，执行ICMP ping
func (d *PingDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
	if packet == nil {
		return consts.QualityUncertain, model2.MessageStatistics{}
	}

	// 减慢请求速度（与C++实现保持一致）
	time.Sleep(1 * time.Second)

	// 执行ping操作
	qua := d.doPing()

	// 填充测点值
	d.fillPoints(packet, qua)

	return qua, model2.MessageStatistics{
		SendCount:    1,
		SuccessCount: boolToUint64(qua == consts.QualityOk),
	}
}

// RequestPing 发送最小化指令
func (d *PingDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	// 减慢请求速度
	time.Sleep(1 * time.Second)
	return d.doPing()
}

// Control 发送控制指令
func (d *PingDevice) Control(_ *model.ControlProtocolPacket, _ string) consts.Quality {
	return consts.QualityOk
}

// doPing 执行ICMP ping操作
func (d *PingDevice) doPing() consts.Quality {
	// 递增序列号
	d.seq++

	// 创建ICMP连接
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return consts.QualityDriverOpenFailed
	}
	defer conn.Close()

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(d.timeout)); err != nil {
		return consts.QualityDriverOpenFailed
	}

	// 构建ICMP Echo Request报文
	msg := d.buildICMPMessage()
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return consts.QualityCmdSendError
	}

	// 解析目标地址
	dst, err := net.ResolveIPAddr("ip4", d.ip)
	if err != nil {
		return consts.QualityConfigError
	}

	// 发送ICMP请求
	n, err := conn.WriteTo(msgBytes, dst)
	if err != nil || n != len(msgBytes) {
		return consts.QualityCmdSendError
	}

	// 循环接收响应，直到匹配的回复或超时
	recvBuf := make([]byte, maxPacketSize)
	for {
		n, peer, err := conn.ReadFrom(recvBuf)
		if err != nil {
			// 检查是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				elog.Errorf("[Ping] ip=%s id=%d seq=%d 接收超时", d.ip, d.id, d.seq)
				return consts.QualityCmdRespTimeout
			}
			elog.Errorf("[Ping] ip=%s id=%d seq=%d 接收错误: %v", d.ip, d.id, d.seq, err)
			return consts.QualityCmdRespError
		}

		// 解析并验证响应
		matched, reason := d.parseResponse(recvBuf[:n], peer)
		if matched {
			return consts.QualityOk
		}

		// 不匹配时记录日志，继续等待下一个报文
		elog.Debugf("[Ping] ip=%s id=%d seq=%d 忽略不匹配的ICMP报文: %s, peer=%v, len=%d",
			d.ip, d.id, d.seq, reason, peer, n)
	}
}

// buildICMPMessage 构建ICMP Echo Request消息
func (d *PingDevice) buildICMPMessage() *icmp.Message {
	// 创建数据部分，填充'z'字符（与C++实现保持一致）
	data := make([]byte, d.dataLen)
	for i := range data {
		data[i] = 'z'
	}

	// 将id和seq放入数据前4字节
	binary.BigEndian.PutUint16(data[0:2], d.id)
	binary.BigEndian.PutUint16(data[2:4], d.seq)

	return &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(d.id),
			Seq:  int(d.seq),
			Data: data,
		},
	}
}

// parseResponse 解析并验证ICMP响应，返回是否匹配以及不匹配的原因
func (d *PingDevice) parseResponse(data []byte, peer net.Addr) (bool, string) {
	// 解析ICMP消息
	msg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), data)
	if err != nil {
		return false, fmt.Sprintf("解析ICMP消息失败: %v", err)
	}

	// 验证消息类型是否为Echo Reply
	if msg.Type != ipv4.ICMPTypeEchoReply {
		return false, fmt.Sprintf("非EchoReply类型: type=%v code=%d", msg.Type, msg.Code)
	}

	// 获取Echo响应体
	echo, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return false, "无法转换为Echo类型"
	}

	// 验证来源IP是否匹配
	if peer != nil {
		peerStr := peer.String()
		if peerStr != d.ip {
			return false, fmt.Sprintf("来源IP不匹配: expect=%s got=%s", d.ip, peerStr)
		}
	}

	// 验证ID是否匹配
	if uint16(echo.ID) != d.id {
		return false, fmt.Sprintf("ID不匹配: expect=%d got=%d", d.id, echo.ID)
	}

	// 验证seq是否匹配
	if uint16(echo.Seq) != d.seq {
		return false, fmt.Sprintf("Seq不匹配: expect=%d got=%d (可能是过期的回复)", d.seq, echo.Seq)
	}

	return true, ""
}

// fillPoints 填充测点值
func (d *PingDevice) fillPoints(packet *model.CollectProtocolPacket, qua consts.Quality) {
	if packet == nil {
		return
	}

	now := utils.GetNowUTCTimeStamp()
	for _, point := range packet.Points {
		// 将质量值作为测点值存储
		point.RtVal.Pv = osal.NewVariantWithValue(int(qua))
		point.RtVal.Qua = qua
		point.RtVal.Tms = now
	}
}

// boolToUint64 布尔转uint64辅助函数
func boolToUint64(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}
