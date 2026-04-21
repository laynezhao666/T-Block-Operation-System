// Package dtcp 提供TCP协议数据的字节序读写工具函数。
// 支持大端序（Big Endian）和小端序（Little Endian）两种字节序。
package dtcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// WriteUint64Little 将uint64值按小端序编码为字节切片。
func WriteUint64Little(value uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		fmt.Println("WriteUint32Little error:", err)
	}
	return buf.Bytes()
}

// ReadUint64Little 从字节切片中按小端序解码uint64值。
func ReadUint64Little(b []byte) uint64 {
	buf := bytes.NewReader(b)
	var value uint64
	err := binary.Read(buf, binary.LittleEndian, &value)
	if err != nil {
		fmt.Println("ReadUint32Little error:", err)
	}
	return value
}

// WriteUint32Little 将uint32值按小端序编码为字节切片。
func WriteUint32Little(value uint32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		fmt.Println("WriteUint32Little error:", err)
	}
	return buf.Bytes()
}

// ReadUint32Little 从字节切片中按小端序解码uint32值。
func ReadUint32Little(b []byte) uint32 {
	buf := bytes.NewReader(b)
	var value uint32
	err := binary.Read(buf, binary.LittleEndian, &value)
	if err != nil {
		fmt.Println("ReadUint32Little error:", err)
	}
	return value
}

// WriteUint16Little 将uint16值按小端序编码为字节切片。
func WriteUint16Little(value uint16) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		fmt.Println("WriteUint16Little error:", err)
	}
	return buf.Bytes()
}

// ReadUint16Little 从字节切片中按小端序解码uint16值。
func ReadUint16Little(b []byte) uint16 {
	buf := bytes.NewReader(b)
	var value uint16
	err := binary.Read(buf, binary.LittleEndian, &value)
	if err != nil {
		fmt.Println("ReadUint16Little error:", err)
	}
	return value
}

// WriteUint64Big 将uint64值按大端序编码为字节切片。
func WriteUint64Big(value uint64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Println("WriteUint32Big error:", err)
	}
	return buf.Bytes()
}

// ReadUint64Big 从字节切片中按大端序解码uint64值。
func ReadUint64Big(b []byte) uint64 {
	buf := bytes.NewReader(b)
	var value uint64
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		fmt.Println("ReadUint32Big error:", err)
	}
	return value
}

// WriteUint32Big 将uint32值按大端序编码为字节切片。
func WriteUint32Big(value uint32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Println("WriteUint32Big error:", err)
	}
	return buf.Bytes()
}

// ReadUint32Big 从字节切片中按大端序解码uint32值。
func ReadUint32Big(b []byte) uint32 {
	buf := bytes.NewReader(b)
	var value uint32
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		fmt.Println("ReadUint32Big error:", err)
	}
	return value
}

// WriteUint16Big 将uint16值按大端序编码为字节切片。
func WriteUint16Big(value uint16) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Println("WriteUint16Big error:", err)
	}
	return buf.Bytes()
}

// ReadUint16Big 从字节切片中按大端序解码uint16值。
func ReadUint16Big(b []byte) uint16 {
	buf := bytes.NewReader(b)
	var value uint16
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		fmt.Println("ReadUint16Big error:", err)
	}
	return value
}

// ReadString 将字节切片转换为字符串。
func ReadString(b []byte) string {
	return string(b)
}

// WriteUint8 将uint8值编码为单字节切片。
func WriteUint8(value uint8) []byte {
	return []byte{value}
}

// ReadUint8 从字节切片中读取第一个字节作为uint8值。
func ReadUint8(b []byte) uint8 {
	return b[0]
}
