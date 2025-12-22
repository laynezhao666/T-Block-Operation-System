package modbus

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"agent/utils/byteorder"

	"agent/entity/consts"
)

func splitCommand(command string) (string, uint16, uint16, error) {
	if len(command) != 10 {
		return "", 0, 0, errors.New("command error")
	}

	functionCode := command[0:2]
	switch functionCode {
	case CodeReadCoils, CodeReadDiscreteInputs, CodeReadHoldingRegisters, CodeReadInputRegisters:
	default:
		return "", 0, 0, errors.New("command function code error")
	}

	startAddrStr := command[2:6]
	startAddrBytes, err := hex.DecodeString(startAddrStr)
	if err != nil {
		return "", 0, 0, fmt.Errorf("command start_addr code error: %v", err)
	}
	startAddr := binary.BigEndian.Uint16(startAddrBytes[:2])

	quantityStr := command[6:10]
	quantityBytes, err := hex.DecodeString(quantityStr)
	if err != nil {
		return "", 0, 0, fmt.Errorf("command quantity code error: %v", err)
	}
	quantity := binary.BigEndian.Uint16(quantityBytes[:2])
	return functionCode, startAddr, quantity, nil
}

func readBool(data uint8, pos int, byteOrder byteorder.ByteOrderExtend) (bool, error) {
	pos = pos & 0x07
	switch byteOrder.String() {
	case consts.LittleEndian, consts.BigEndianSwap:
		return ((data >> (7 - pos)) & 0x01) > 0, nil
	case consts.BigEndian, consts.LittleEndianSwap:
		return ((data >> pos) & 0x01) > 0, nil
	default:
		return false, fmt.Errorf("endian config error, endian=%s", byteOrder.String())
	}
}
