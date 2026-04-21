// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"encoding/binary"
	"fmt"
)

// Packet CHD 协议数据包结构
type Packet struct {
	SOI    uint8  // 起始符 0x7E
	SeqNo  uint8  // 命令包序号
	ADR1   uint8  // 组内地址
	ADR2   uint8  // 分组地址
	CID    uint8  // 数据包类别 (CID2/RTN/UP-REP，命令/应答/主动上传)
	LTH    uint16 // 参数长度校验 (高字节=校验码, 低12位=长度*2)
	INFO   []byte // 参数信息
	CHKSUM uint16 // 桢校验和
	EOI    uint8  // 结束符 0x0D
}

// NewPacket 创建新的数据包
func NewPacket(seqNo, adr1, adr2, cid uint8, info []byte) *Packet {
	p := &Packet{
		SOI:   0x7E,
		SeqNo: seqNo,
		ADR1:  adr1,
		ADR2:  adr2,
		CID:   cid,
		INFO:  info,
		EOI:   0x0D,
	}
	p.calcLTH()
	p.calcChecksum()
	return p
}

// Pack 打包数据（转换为 ASCII 格式）
func (p *Packet) Pack() []byte {
	// 1. 先计算校验码
	p.calcLTH()
	p.calcChecksum()

	// 2. 构建原始字节数组（不含 SOI 和 EOI）
	rawData := make([]byte, 0, 7+len(p.INFO))
	rawData = append(rawData, p.SeqNo)
	rawData = append(rawData, p.ADR1)
	rawData = append(rawData, p.ADR2)
	rawData = append(rawData, p.CID)
	rawData = append(rawData, byte(p.LTH>>8), byte(p.LTH&0xFF)) // 高字节在前
	rawData = append(rawData, p.INFO...)
	rawData = append(rawData, byte(p.CHKSUM>>8), byte(p.CHKSUM&0xFF)) // 高字节在前

	// 3. 转换为 ASCII 格式
	asciiData := make([]byte, 0, 1+len(rawData)*2+1)
	asciiData = append(asciiData, p.SOI) // SOI 直接发送

	// 将每个字节拆分为两个 ASCII 字符
	for _, b := range rawData {
		high := (b >> 4) & 0x0F // 高4位
		low := b & 0x0F         // 低4位
		asciiData = append(asciiData, byteToASCII(high))
		asciiData = append(asciiData, byteToASCII(low))
	}

	asciiData = append(asciiData, p.EOI) // EOI 直接发送

	return asciiData
}

// Unpack 解包数据（从 ASCII 格式解析）
func Unpack(data []byte) (*Packet, error) {
	if len(data) < 9 { // 最小长度：SOI(1) + 至少14个ASCII字符 + EOI(1)
		return nil, fmt.Errorf("数据包长度不足: %d", len(data))
	}

	// 1. 检查起始符和结束符
	if data[0] != 0x7E {
		return nil, fmt.Errorf("起始符错误: 0x%02X", data[0])
	}
	if data[len(data)-1] != 0x0D {
		return nil, fmt.Errorf("结束符错误: 0x%02X", data[len(data)-1])
	}

	// 2. 提取 ASCII 数据部分（去掉 SOI 和 EOI）
	asciiData := data[1 : len(data)-1]
	if len(asciiData)%2 != 0 {
		return nil, fmt.Errorf("ASCII 数据长度必须是偶数: %d", len(asciiData))
	}

	// 3. 将 ASCII 转换为原始字节
	rawData := make([]byte, len(asciiData)/2)
	for i := 0; i < len(asciiData); i += 2 {
		high, err := asciiToByte(asciiData[i])
		if err != nil {
			return nil, fmt.Errorf("ASCII 转换错误[%d]: %v", i, err)
		}
		low, err := asciiToByte(asciiData[i+1])
		if err != nil {
			return nil, fmt.Errorf("ASCII 转换错误[%d]: %v", i+1, err)
		}
		rawData[i/2] = (high << 4) | low
	}

	// 4. 解析数据包字段
	// 最小长度：SeqNo(1) + ADR1(1) + ADR2(1) + CID(1) + LTH(2) + CHKSUM(2) = 8
	if len(rawData) < 8 {
		return nil, fmt.Errorf("原始数据长度不足: %d (最小需要8字节)", len(rawData))
	}

	p := &Packet{
		SOI:    0x7E,
		SeqNo:  rawData[0],
		ADR1:   rawData[1],
		ADR2:   rawData[2],
		CID:    rawData[3],
		LTH:    binary.BigEndian.Uint16(rawData[4:6]),
		CHKSUM: binary.BigEndian.Uint16(rawData[len(rawData)-2:]),
		EOI:    0x0D,
	}

	// 5. 提取 INFO 部分
	if len(rawData) > 8 {
		p.INFO = rawData[6 : len(rawData)-2]
	}

	// 6. 校验数据包
	if err := p.Verify(); err != nil {
		return nil, err
	}

	return p, nil
}

// calcLTH 计算参数长度校验码
func (p *Packet) calcLTH() {
	infoLen := len(p.INFO)
	asciiLen := infoLen * 2 // ASCII 码个数

	// 将长度拆分为 3 个 4 位数
	d0 := (asciiLen >> 0) & 0x0F
	d1 := (asciiLen >> 4) & 0x0F
	d2 := (asciiLen >> 8) & 0x0F

	// 累加并模 16
	sum := (d0 + d1 + d2) & 0x0F

	// 求补（取反加1）
	checksum := (^sum + 1) & 0x0F

	// 组合：高4位=校验码，低12位=长度
	p.LTH = uint16(checksum<<12) | uint16(asciiLen&0x0FFF)
}

// calcChecksum 计算桢校验和
func (p *Packet) calcChecksum() {
	// 构建需要校验的数据（SeqNo 到 INFO 结束）
	data := make([]byte, 0, 5+len(p.INFO))
	data = append(data, p.SeqNo)
	data = append(data, p.ADR1)
	data = append(data, p.ADR2)
	data = append(data, p.CID)
	data = append(data, byte(p.LTH>>8), byte(p.LTH&0xFF))
	data = append(data, p.INFO...)

	// 将每个字节转换为 ASCII 后累加
	var sum uint32
	for _, b := range data {
		high := (b >> 4) & 0x0F
		low := b & 0x0F
		sum += uint32(byteToASCII(high))
		sum += uint32(byteToASCII(low))
	}

	// 模 65536 后取补
	sum = sum & 0xFFFF
	p.CHKSUM = uint16((^sum + 1) & 0xFFFF)
}

// Verify 校验数据包
func (p *Packet) Verify() error {
	// 1. 校验长度
	infoLen := len(p.INFO)
	asciiLen := infoLen * 2
	expectedLen := asciiLen & 0x0FFF
	actualLen := int(p.LTH & 0x0FFF)

	if actualLen != expectedLen {
		return fmt.Errorf("长度校验失败: 期望=%d, 实际=%d", expectedLen, actualLen)
	}

	// 2. 校验长度校验码
	d0 := (asciiLen >> 0) & 0x0F
	d1 := (asciiLen >> 4) & 0x0F
	d2 := (asciiLen >> 8) & 0x0F
	d3 := int((p.LTH >> 12) & 0x0F)

	if ((d0 + d1 + d2 + d3) & 0x0F) != 0 {
		return fmt.Errorf("长度校验码错误")
	}

	// 3. 校验桢校验和
	oldChecksum := p.CHKSUM
	p.calcChecksum()
	if p.CHKSUM != oldChecksum {
		return fmt.Errorf("桢校验和错误: 期望=0x%04X, 实际=0x%04X", p.CHKSUM, oldChecksum)
	}

	return nil
}

// String 返回数据包的字符串表示（用于调试）
func (p *Packet) String() string {
	return fmt.Sprintf("Packet{SeqNo=0x%02X, ADR1=0x%02X, ADR2=0x%02X, "+
		"CID=0x%02X, LTH=0x%04X, INFO=%d bytes, CHKSUM=0x%04X}",
		p.SeqNo, p.ADR1, p.ADR2, p.CID, p.LTH, len(p.INFO), p.CHKSUM)
}

// ============ 辅助函数 ============

// byteToASCII 将 4 位数值转换为 ASCII 字符
func byteToASCII(b uint8) uint8 {
	if b <= 9 {
		return '0' + b // 0-9 -> '0'-'9'
	}
	return 'A' + (b - 10) // 10-15 -> 'A'-'F'
}

// asciiToByte 将 ASCII 字符转换为 4 位数值
func asciiToByte(ascii uint8) (uint8, error) {
	if ascii >= '0' && ascii <= '9' {
		return ascii - '0', nil
	}
	if ascii >= 'A' && ascii <= 'F' {
		return ascii - 'A' + 10, nil
	}
	if ascii >= 'a' && ascii <= 'f' {
		return ascii - 'a' + 10, nil
	}
	return 0, fmt.Errorf("无效的 ASCII 字符: 0x%02X", ascii)
}
