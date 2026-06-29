package iec104

import (
	"encoding/binary"
	"fmt"
)

// ASDU 应用服务数据单元
type ASDU struct {
	TypeID        byte   // 类型标识
	Sequence      byte   // 是(1)否(0)连续
	Num           byte   // 可变结构限定词
	Cause         uint16 // 传输原因
	PublicAddress uint16 // 公共地址
}

// ParseASDU 解析asdu
func (asdu *ASDU) ParseASDU(asduBytes []byte) ([]*Signal, error) {
	if asduBytes == nil || len(asduBytes) < asduLen {
		return nil, fmt.Errorf("asdu[%X]非法", asduBytes)
	}
	asdu.TypeID = asduBytes[ASDUOffsetTypeID]
	// 数据是否连续
	asdu.Sequence, asdu.Num = asdu.ParseVariable(asduBytes[ASDUOffsetSequenceAndNum])
	asdu.Cause = binary.LittleEndian.Uint16([]byte{asduBytes[ASDUOffsetCause0], asduBytes[ASDUOffsetCause1]})
	asdu.PublicAddress = binary.LittleEndian.Uint16([]byte{asduBytes[ASDUOffsetPublicAddress0],
		asduBytes[ASDUOffsetPublicAddress1]})
	switch asdu.TypeID {
	case MSpNa1:
		return parseMSpNa1(asdu, asduBytes)
	case MMeNc1:
		return parseMMeNc1(asdu, asduBytes)
	case MItNa1:
		return parseMItNa1(asdu, asduBytes)
	case CIcNa1, CCiNa1, MEiNA1:
		return nil, nil
	default:
		return nil, fmt.Errorf("暂不支持的数据类型:%d", uint(asdu.TypeID))
	}
}

// ParseVariable 解析asdu可变结构限定词
func (asdu *ASDU) ParseVariable(b byte) (sq byte, length byte) {
	return b >> 7, b & 0x7F
}
