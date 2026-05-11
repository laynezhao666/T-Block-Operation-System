package model

import (
	"errors"
	"fmt"
	"go.bug.st/serial"
	"agent/entity/consts"
	"strconv"
	"strings"
)

// Channel 设备通道
type Channel struct {
	Name    string
	Params  string
	Address string
}

// ChannelInfo 描述程序如何访问设备
// 包括通道地址、打开需要的参数等
type ChannelInfo struct {
	// 名称，如: /dev/ttymxc1, 192.168.1.250:161
	Name string
	// 参数，如: 9600:8:N:1
	Params string
	// 地址，与具体协议相关
	Address string
	// 协议版本号
	ProtocolVer string
	// 与设备接口单次通讯超时
	TimeoutMs int
	// 并发协程数
	ParallelCount int
	// 请求包中允许的最大测点数
	PacketMaxPointCount int
	// 扩展参数
	ExtendKV     map[string]string
	DriverExtend string
	ChType       string
}

// ParseRTUParam  rtu 解析
func ParseRTUParam(params string) (int, string, int, int, error) {
	parts := strings.Split(params, consts.CollectParamSep)
	if len(parts) != 4 {
		return 0, "", 0, 0, fmt.Errorf("param [%v] len err", params)
	}

	// 解析波特率
	baudRate, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", 0, 0, errors.New("baudRate parse err")
	}

	// 解析奇偶校验
	parity := parts[1]

	// 解析数据位
	dataBits, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, "", 0, 0, errors.New("dataBits parse err")
	}

	// 解析停止位
	stopBits, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, "", 0, 0, errors.New("stopBits parse err")
	}

	return baudRate, parity, dataBits, stopBits, nil
}

// NormalizeParity parity规范
func NormalizeParity(p string) string {
	switch p {
	case "E", "e":
		return "E"
	case "O", "o":
		return "O"
	default:
		return "N"
	}
}

// MapStopBits stopBits映射
func MapStopBits(n int) serial.StopBits {
	switch n {
	case 2:
		return serial.TwoStopBits
	default:
		return serial.OneStopBit
	}
}

// ParseParity parity解析
func ParseParity(p string) serial.Parity {
	switch p {
	case "E":
		return serial.EvenParity
	case "O":
		return serial.OddParity
	default:
		return serial.NoParity
	}
}
