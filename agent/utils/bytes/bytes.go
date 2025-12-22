package ubytes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"agent/entity/definition/datatype"
	"strconv"
)

// 通用：任意类型转字节（指定字节序）
func toBytes(v interface{}, byteOrder binary.ByteOrder) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, byteOrder, v)
	return buf.Bytes(), err
}

// Modbus线圈布尔值专用编码
func boolToModbusCoilBytes(b bool) []byte {
	if b {
		return []byte{0xFF, 0x00}
	}
	return []byte{0x00, 0x00}
}

// ConvertStringToBytes 主转换函数
func ConvertStringToBytes(val string, dataType datatype.DataType, byteOrder binary.ByteOrder) ([]byte, error) {
	switch dataType {
	case datatype.BoolType:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return boolToModbusCoilBytes(b), nil

	case datatype.Int16Type:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, err
		}
		return toBytes(int16(v), byteOrder)

	case datatype.Uint16Type:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return nil, err
		}
		return toBytes(uint16(v), byteOrder)

	case datatype.Int32Type:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return nil, err
		}
		return toBytes(int32(v), byteOrder)

	case datatype.Uint32Type:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, err
		}
		return toBytes(uint32(v), byteOrder)

	case datatype.Int64Type:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return toBytes(v, byteOrder)

	case datatype.Uint64Type:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return toBytes(v, byteOrder)

	case datatype.FloatType:
		f, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return nil, err
		}
		return toBytes(float32(f), byteOrder)

	case datatype.DoubleType:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		return toBytes(f, byteOrder)

	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}

// BytesToUint16 ff
func BytesToUint16(data []byte, byteOrder binary.ByteOrder) (uint16, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid data length: expected 2 bytes, got %d", len(data))
	}
	buf := bytes.NewReader(data)
	var value uint16
	if err := binary.Read(buf, byteOrder, &value); err != nil {
		return 0, err
	}
	return value, nil
}
