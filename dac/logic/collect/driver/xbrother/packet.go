// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"bytes"
	"fmt"
	"math/rand"
	"unsafe"

	"dac/entity/utils/dtcp"
	"dac/logic/collect/driver/xbrother/consts"

	"dac/entity/utils/tlog"
)

// Packet XBrother协议数据包，负责协议帧的构建和解析
type Packet struct {
	header  uint8  // 帧头标识
	randNum uint8  // 随机数（保留字段）
	cmd     uint8  // 命令码
	address uint8  // 控制器地址
	doorNo  uint8  // 门编号
	length  uint16 // 数据长度

	sendBuff    *bytes.Buffer // 发送缓冲区
	recvBuff    *bytes.Buffer // 接收缓冲区
	recvData    []byte        // 接收到的数据
	finalCalcCS uint8         // 最终计算的校验和
	firstCS     uint8         // 首次接收的校验和
	tail        uint8         // 帧尾标识
	rtn         uint8         // 返回码
	hasRtn      bool          // 是否包含返回码

	logger tlog.Logger // 日志记录器
}

// NewPacket 创建新的协议数据包实例
func NewPacket(logger tlog.Logger) *Packet {
	p := new(Packet)
	p.logger = logger
	return p
}

// ParseFirstRecv 解析首次接收的数据帧头部，返回剩余需接收的字节数
func (p *Packet) ParseFirstRecv(buff []byte) (int, bool) {
	if buff == nil {
		return 0, false
	}
	p.recvBuff = bytes.NewBuffer(buff)
	if !p.ParseHeader() {
		return 0, false
	}
	p.ParseRand()
	p.ParseCommand()
	p.ParseAddress()
	p.ParseDoorNo()
	p.ParseLength()
	secondRecvLen := int(p.length) + consts.CSAndETXLen
	p.firstCS = CS8Update(0, buff)
	return secondRecvLen, true
}

// ParseLastRecv 解析后续接收的数据，校验CS和ETX
func (p *Packet) ParseLastRecv(buff []byte) bool {
	p.recvBuff = bytes.NewBuffer(buff)
	p.recvData = p.recvBuff.Next(int(p.length))
	p.finalCalcCS = CS8Update(p.firstCS, p.recvData)
	if recvCS := dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.finalCalcCS)))); recvCS != p.finalCalcCS {
		p.logger.Errorf("cs 不一致，接收: %x, 计算: %x\n", recvCS, p.finalCalcCS)
		return false
	}
	if recvEXT := dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.tail)))); recvEXT != consts.ETX {
		p.logger.Errorf("ext 不一致， 接收: %x, 期望: %x\n", recvEXT, consts.ETX)
		return false
	}
	return true
}

// BuildPacket 构建发送数据包，包含帧头、命令、数据和校验
func (p *Packet) BuildPacket(requestCmd uint8, doorNo uint8, data []byte) {
	p.sendBuff = bytes.NewBuffer(make([]byte, 0))
	p.FillHeader()
	p.randNum = generateRandomUint8() // 保留字段，默认值，无实际作用
	p.sendBuff.Write(dtcp.WriteUint8(p.randNum))
	p.sendBuff.Write(dtcp.WriteUint8(requestCmd))
	p.sendBuff.Write(dtcp.WriteUint8(p.address))
	p.sendBuff.Write(dtcp.WriteUint8(doorNo))
	p.sendBuff.Write(dtcp.WriteUint16Little(uint16(len(data))))
	p.sendBuff.Write(data)
	p.sendBuff.Write(dtcp.WriteUint8(CS8Update(0, p.sendBuff.Bytes())))
	p.sendBuff.Write(dtcp.WriteUint8(consts.ETX))
}

// ParseHeader 解析帧头标识
func (p *Packet) ParseHeader() bool {
	p.header = dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.header))))
	if p.header != consts.STX {
		fmt.Printf("recv header: %x , expect header: %x\n", p.header, consts.STX)
		return false
	}
	return true
}

// SendData 获取待发送的完整数据
func (p *Packet) SendData() []byte {
	return p.sendBuff.Bytes()
}

// ParseLength 解析数据长度字段
func (p *Packet) ParseLength() {
	p.length = dtcp.ReadUint16Big(p.recvBuff.Next(int(unsafe.Sizeof(p.length))))
}

// ParseCommand 解析命令码字段
func (p *Packet) ParseCommand() {
	p.cmd = dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.cmd))))
}

// ParseRand 解析随机数字段
func (p *Packet) ParseRand() {
	p.randNum = dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.randNum))))
}

// ParseAddress 解析控制器地址字段
func (p *Packet) ParseAddress() {
	p.address = dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.address))))
}

// ParseDoorNo 解析门编号字段
func (p *Packet) ParseDoorNo() {
	p.doorNo = dtcp.ReadUint8(p.recvBuff.Next(int(unsafe.Sizeof(p.doorNo))))
}

// FillHeader 填充帧头标识到发送缓冲区
func (p *Packet) FillHeader() {
	p.sendBuff.Write(dtcp.WriteUint8(consts.STX))
}

// FillCS 填充校验和到发送缓冲区
func (p *Packet) FillCS() {
	p.sendBuff.Write(dtcp.WriteUint8(p.finalCalcCS))
}

// CS8Update 计算XOR校验和
func CS8Update(cs8 uint8, buff []byte) uint8 {
	var crc = cs8
	for _, b := range buff {
		crc ^= b
	}
	return crc
}

// generateRandomUint8 生成随机uint8值
func generateRandomUint8() uint8 {
	randomUint8 := uint8(rand.Intn(256))
	return randomUint8
}
