// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"bytes"
	"fmt"
	"unsafe"

	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/cacs/consts"
)

// Packet CACS协议数据包，负责协议帧的构建和解析。
// 帧格式：Header(4) + Length(4) + Command(4) + [RTN(4)] + Data(N) + CRC16(2)
type Packet struct {
	header       uint32        // 帧头标识
	length       uint32        // 数据长度（含命令码）
	cmd          uint32        // 命令码
	sendBuff     *bytes.Buffer // 发送缓冲区
	recvBuff     *bytes.Buffer // 接收缓冲区
	recvData     []byte        // 接收到的数据
	finalCalcCRC uint16        // 最终计算的CRC校验值
	firstCRC     uint16        // 首次接收的CRC校验值
	rtn          uint32        // 返回码
	hasRtn       bool          // 是否包含返回码
}

// ParseFirstRecv 解析首次接收的数据帧头部，输出剩余需接收的字节数
func (p *Packet) ParseFirstRecv(buff []byte, len *uint32) bool {
	if buff == nil {
		return false
	}
	p.recvBuff = bytes.NewBuffer(buff)
	if !p.ParseHeader() {
		return false
	}

	p.ParseLength()
	p.ParseCommand()

	*len = p.DataLen()
	p.firstCRC = CRC16(buff)
	return true
}

// ParseLastRecv 解析后续接收的数据，校验CRC
func (p *Packet) ParseLastRecv(buff []byte) bool {
	p.recvBuff = bytes.NewBuffer(buff)

	if !p.ParseReturn() {
		return false
	}

	// 计算实际数据长度（减去CRC和可能的RTN字段）
	var recvDataLen int
	recvDataLen = len(buff) - int(unsafe.Sizeof(p.finalCalcCRC))
	if p.hasRtn {
		recvDataLen -= int(unsafe.Sizeof(p.rtn))
	}
	p.recvData = p.recvBuff.Next(recvDataLen)
	p.finalCalcCRC = CRC16Update(
		p.firstCRC, buff,
		len(buff)-int(unsafe.Sizeof(p.finalCalcCRC)))

	// 校验CRC
	var recvCRC uint16
	recvCRC = dtcp.ReadUint16Little(
		p.recvBuff.Next(int(unsafe.Sizeof(p.finalCalcCRC))))
	if recvCRC != p.finalCalcCRC {
		fmt.Printf(
			"crc 不一致，接收: %x, 计算: %x\n",
			recvCRC, p.finalCalcCRC)
		return false
	}
	return true
}

// BuildRequestPacket 构建请求数据包
func (p *Packet) BuildRequestPacket(
	requestCmd uint32, data []byte,
) {
	p.sendBuff = bytes.NewBuffer(make([]byte, 0))
	p.FillHeader()
	l := uint32(unsafe.Sizeof(requestCmd)) + uint32(len(data))
	p.sendBuff.Write(dtcp.WriteUint32Little(l))
	p.sendBuff.Write(dtcp.WriteUint32Little(requestCmd))
	p.sendBuff.Write(data)
	p.FillCRC(p.sendBuff.Bytes())
}

// BuildResponsePacket 构建响应数据包（含返回码）
func (p *Packet) BuildResponsePacket(
	responseCmd uint32, rtn uint32, data []byte,
) {
	p.sendBuff = bytes.NewBuffer(make([]byte, 0))
	p.FillHeader()
	l := uint32(unsafe.Sizeof(responseCmd)) +
		uint32(unsafe.Sizeof(rtn)) + uint32(len(data))
	p.sendBuff.Write(dtcp.WriteUint32Little(l))
	p.sendBuff.Write(dtcp.WriteUint32Little(responseCmd))
	p.sendBuff.Write(dtcp.WriteUint32Little(rtn))
	p.sendBuff.Write(data)
	p.FillCRC(p.sendBuff.Bytes())
}

// RecvCommand 获取接收到的命令码
func (p *Packet) RecvCommand() uint32 {
	return p.cmd
}

// SendData 获取待发送的完整数据
func (p *Packet) SendData() []byte {
	return p.sendBuff.Bytes()
}

// DataLen 获取数据部分长度（不含命令码）
func (p *Packet) DataLen() uint32 {
	return p.length - uint32(unsafe.Sizeof(p.cmd))
}

// ParseReturn 判断当前数据包是否包含RTN字段。
// 门控器主动上报的数据包不含RTN字段。
func (p *Packet) ParseReturn() bool {
	switch p.cmd {
	case consts.KCommandRequestRegister, // 注册请求（主动上报）
		consts.KCommandUploadDoorStatus,        // 门状态上报（主动上报）
		consts.KCommandUploadControllerStatus,  // 控制器状态上报（主动上报）
		consts.KCommandRequestUploadEventAlarm: // 事件告警上报（主动上报）
		p.hasRtn = false
		return true
	default:
		p.rtn = dtcp.ReadUint32Little(p.recvBuff.Next(int(unsafe.Sizeof(p.rtn))))
		p.hasRtn = true
	}
	return true
}

// ParseHeader 解析帧头标识
func (p *Packet) ParseHeader() bool {
	p.header = dtcp.ReadUint32Little(p.recvBuff.Next(int(unsafe.Sizeof(p.header))))
	if p.header != consts.KHeader {
		fmt.Printf("recv header: %x , expect header: %x\n", p.header, consts.KHeader)
		return false
	}
	return true
}

// ParseLength 解析数据长度字段
func (p *Packet) ParseLength() {
	p.length = dtcp.ReadUint32Little(p.recvBuff.Next(int(unsafe.Sizeof(p.length))))
}

// ParseCommand 解析命令码字段
func (p *Packet) ParseCommand() {
	p.cmd = dtcp.ReadUint32Little(p.recvBuff.Next(int(unsafe.Sizeof(p.cmd))))
}

// FillHeader 填充帧头标识到发送缓冲区
func (p *Packet) FillHeader() {
	p.sendBuff.Write(dtcp.WriteUint32Little(consts.KHeader))
}

// FillCRC 计算并填充CRC校验值到发送缓冲区
func (p *Packet) FillCRC(buf []byte) {
	p.sendBuff.Write(dtcp.WriteUint16Little(CRC16(buf)))
}

// GetRtn 获取RTN值
func (p *Packet) GetRtn() uint32 {
	return p.rtn
}

// HasRtn 判断数据包是否包含RTN字段
func (p *Packet) HasRtn() bool {
	return p.hasRtn
}
