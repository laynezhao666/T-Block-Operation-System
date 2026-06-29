package iec104

import (
	"encoding/binary"
	"math"
	"time"
)

// parseMSpNa1 解析MSpNa1类型，不带游标的单点遥信，3个字节的地址，1个字节的值
func parseMSpNa1(asdu *ASDU, asduBytes []byte) ([]*Signal, error) {
	signals := make([]*Signal, 0)
	tp := uint(asdu.TypeID)
	timeNow := time.Now().UnixMilli()
	if asdu.Sequence == 1 {
		fixed := asduLen + seqAddrLen
		size := len(asduBytes[fixed:]) / int(asdu.Num)
		msg := asduBytes[fixed:]
		address := binary.LittleEndian.Uint32(append([]byte{asduBytes[asduLen], asduBytes[asduLen+1],
			asduBytes[asduLen+2]}, 0x00))
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: address + uint32(i/size),
				Value:   msg[i],
				Cts:     timeNow,
			})
		}
	} else {
		size := len(asduBytes[asduLen:]) / int(asdu.Num)
		msg := asduBytes[asduLen:]
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: binary.LittleEndian.Uint32(append([]byte{msg[i], msg[i+1], msg[i+2]}, 0x00)),
				Value:   msg[i+3],
				Cts:     timeNow,
			})
		}
	}
	return signals, nil
}

func parseMItNa1(asdu *ASDU, asduBytes []byte) ([]*Signal, error) {
	signals := make([]*Signal, 0)
	tp := uint(asdu.TypeID)
	timeNow := time.Now().UnixMilli()
	if asdu.Sequence == 1 {
		fixed := asduLen + seqAddrLen
		size := len(asduBytes[fixed:]) / int(asdu.Num)
		msg := asduBytes[fixed:]
		address := binary.LittleEndian.Uint32(append([]byte{asduBytes[asduLen], asduBytes[asduLen+1],
			asduBytes[asduLen+2]}, 0x00))
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: address + uint32(i/size),
				Value:   int32(binary.LittleEndian.Uint32(msg[i : i+4])),
				Cts:     timeNow,
			})
		}
	} else {
		size := len(asduBytes[asduLen:]) / int(asdu.Num)
		msg := asduBytes[asduLen:]
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: binary.LittleEndian.Uint32(append([]byte{msg[i], msg[i+1], msg[i+2]}, 0x00)),
				Value:   int32(binary.LittleEndian.Uint32(msg[i+3 : i+7])),
				Cts:     timeNow,
			})
		}
	}
	return signals, nil
}

// parseMMeNc1 解析MMeNc1类型，带品质描述的浮点值，每个遥测值占5个字节
func parseMMeNc1(asdu *ASDU, asduBytes []byte) ([]*Signal, error) {
	signals := make([]*Signal, 0)
	tp := uint(asdu.TypeID)
	timeNow := time.Now().UnixMilli()
	if asdu.Sequence == 1 {
		fixed := asduLen + seqAddrLen
		size := len(asduBytes[fixed:]) / int(asdu.Num)
		msg := asduBytes[fixed:]
		address := binary.LittleEndian.Uint32(append([]byte{asduBytes[asduLen], asduBytes[asduLen+1],
			asduBytes[asduLen+2]}, 0x00))
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: address + uint32(i/size),
				Value:   float64(math.Float32frombits(binary.LittleEndian.Uint32(msg[i : i+4]))),
				Quality: msg[i+4],
				Cts:     timeNow,
			})
		}
	} else {
		size := len(asduBytes[asduLen:]) / int(asdu.Num)
		msg := asduBytes[asduLen:]
		lenn := int(asdu.Num) * size
		for i := 0; i < lenn; i += size {
			signals = append(signals, &Signal{
				TypeID:  tp,
				Address: binary.LittleEndian.Uint32(append([]byte{msg[i], msg[i+1], msg[i+2]}, 0x00)),
				Value:   float64(math.Float32frombits(binary.LittleEndian.Uint32(msg[i+3 : i+7]))),
				Quality: msg[i+7],
				Cts:     timeNow,
			})
		}
	}
	return signals, nil
}
