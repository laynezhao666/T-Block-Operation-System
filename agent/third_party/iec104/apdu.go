package iec104

import (
	"fmt"
	"strings"
)

// APDU 104数据包
type APDU struct {
	APCI     *APCI
	ASDU     *ASDU
	Len      int
	ASDULen  int
	CtrType  byte
	CtrFrame interface{}
	Signals  []*Signal
}

// ParseAPDU 解析APDU
func ParseAPDU(input []byte) (*APDU, error) {
	apdu := new(APDU)
	if input == nil || len(input) < apciLen {
		return apdu, fmt.Errorf("APDU报文[%X]非法", input)
	}
	apci := &APCI{
		Start:   input[APCIOffsetStart],
		ApduLen: int(input[APCIOffsetApduLen]),
		Ctr1:    input[APCIOffsetCtr1],
		Ctr2:    input[APCIOffsetCtr2],
		Ctr3:    input[APCIOffsetCtr3],
		Ctr4:    input[APCIOffsetCtr4],
	}
	fType, ctrFrame, err := apci.ParseCtr()
	if err != nil {
		return apdu, fmt.Errorf("APDU报文[%X]解析控制域异常: %v", input, err)
	}
	data := input[apciLen:]
	asdu := new(ASDU)
	var asduLength int
	signals := make([]*Signal, 0)
	if len(data) < 1 {
		asduLength = 0
	} else {
		signals, err = asdu.ParseASDU(data)
		if err != nil {
			return apdu, fmt.Errorf("APDU报文[%X]解析ASDU域[%X]异常: %v", input, data, err)
		}
		asduLength = len(data)
	}
	apdu.APCI = apci
	apdu.ASDU = asdu
	apdu.Len = apci.ApduLen
	apdu.ASDULen = asduLength
	apdu.CtrType = fType
	apdu.CtrFrame = ctrFrame
	apdu.Signals = signals
	return apdu, nil
}

// String 定义结构体输出格式
func (a APDU) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("APDU: {Len: %d, ASDULen:  %d, CtrType %v, CtrFrame: %v, ",
		a.Len, a.ASDULen, a.CtrType, a.CtrFrame))
	sb.WriteString("Signs: [")
	for _, s := range a.Signals {
		sb.WriteString(fmt.Sprintf("{%+v}", s))
		sb.WriteString(",")
	}
	sb.WriteString("]}")
	return sb.String()
}
