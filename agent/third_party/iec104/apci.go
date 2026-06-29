package iec104

import (
	"encoding/binary"
	"fmt"
)

// APCI APCI
type APCI struct {
	Start   byte // 起始
	ApduLen int  // Apdu长度
	Ctr1    byte // 控制域1
	Ctr2    byte // 控制域2
	Ctr3    byte // 控制域3
	Ctr4    byte // 控制域4
}

// IFrame I帧
type IFrame struct {
	Send uint16
	Recv uint16
}

// SFrame S帧
type SFrame struct {
	Recv uint16
}

// UFrame U帧
type UFrame struct {
	cmd [4]byte // 命令
}

func convert4BytesToSlice(b [4]byte) []byte {
	return []byte{b[0], b[1], b[2], b[3]}
}

// parseBigEndianUint16 转换大端Uint16
func parseBigEndianUInt16(i uint16) []byte {
	bytes := make([]byte, 2, 2)
	binary.BigEndian.PutUint16(bytes, i)
	return bytes
}

// parseLittleEndianUint16 转换小端Uint16
func parseLittleEndianUInt16(i uint16) []byte {
	bytes := make([]byte, 2, 2)
	binary.LittleEndian.PutUint16(bytes, i)
	return bytes
}

// convertBytes 构造待发送的数据包
func convertBytes(data []byte) []byte {
	sendData := make([]byte, 0, 2+len(data))
	iBytes := parseBigEndianUInt16(uint16(len(data)))
	sendData = append(sendData, startFrame)
	sendData = append(sendData, iBytes[1])
	sendData = append(sendData, data...)
	return sendData
}

// ParseCtr 解析控制域
func (apci *APCI) ParseCtr() (byte, interface{}, error) {
	switch {
	case apci.Ctr1&1 == iFrame:
		// I帧
		t, f := apci.parseIFrame()
		return t, f, nil
	case apci.Ctr1&3 == sFrame:
		// S帧
		t, f := apci.parseSFrame()
		return t, f, nil
	case apci.Ctr1&3 == uFrame:
		// U帧
		t, f := apci.parseUFrame()
		return t, f, nil
	default:
		return 0xFF, nil, fmt.Errorf("未知APCI帧类型")
	}
}

// parseIFrame 解析I帧
func (apci *APCI) parseIFrame() (byte, IFrame) {
	send := uint16(apci.Ctr1)>>1 + uint16(apci.Ctr2)<<7
	recv := uint16(apci.Ctr3)>>1 + uint16(apci.Ctr4)<<7
	return iFrame, IFrame{
		Send: send,
		Recv: recv,
	}
}

// parseIFrame 解析S帧
func (apci *APCI) parseSFrame() (byte, SFrame) {
	recv := uint16(apci.Ctr3)>>1 + uint16(apci.Ctr4)<<7
	return sFrame, SFrame{
		Recv: recv,
	}
}

// parseIFrame 解析U帧
func (apci *APCI) parseUFrame() (byte, UFrame) {
	cmd := [4]byte{apci.Ctr1, apci.Ctr2, apci.Ctr3, apci.Ctr4}
	return uFrame, UFrame{
		cmd: cmd,
	}
}
