package kbt

import (
	"bufio"
	"context"
	"agent/logic/collector/device/model"
	"time"

	serial "go.bug.st/serial"
	"trpc.group/trpc-go/trpc-go/log"
)

// SerialOptions 串口参数
type SerialOptions struct {
	Port       string
	Baud       int
	DataBits   int
	StopBits   int
	Parity     string
	ReadTO     time.Duration
	WriteTO    time.Duration
	HardwareFC bool
	SoftwareFC bool
}

// SerialPort 串口
type SerialPort struct {
	p   serial.Port
	opt SerialOptions
	r   *bufio.Reader
}

// OpenSerial 打开串口
func OpenSerial(opt SerialOptions) (*SerialPort, error) {
	mode := &serial.Mode{
		BaudRate: opt.Baud,
		DataBits: opt.DataBits,
		StopBits: model.MapStopBits(opt.StopBits),
		Parity:   model.ParseParity(opt.Parity),
	}
	p, err := serial.Open(opt.Port, mode)
	if err != nil {
		return nil, err
	}
	_ = p.SetReadTimeout(opt.ReadTO)
	return &SerialPort{p: p, opt: opt, r: bufio.NewReader(p)}, nil
}

func (s *SerialPort) Close() error { return s.p.Close() }

func (s *SerialPort) WriteAll(b []byte) error {
	for len(b) > 0 {
		n, err := s.p.Write(b)
		if err != nil {
			return err
		}
		b = b[n:]
	}
	return nil
}

// Read 读取数据
func (s *SerialPort) Read(ctx context.Context, max int) ([]byte, error) {
	var out []byte
	for {
		// 根据ctx动态设定一次ReadTimeout
		to := s.opt.ReadTO
		if dl, ok := ctx.Deadline(); ok {
			if remain := time.Until(dl); remain > 0 && (to <= 0 || remain < to) {
				to = remain
			}
		}
		_ = s.p.SetReadTimeout(to)

		b := make([]byte, 128)
		n, err := s.p.Read(b)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			continue
		}
		out = append(out, b[:n]...)
		if max > 0 && len(out) >= max {
			return out, nil
		}
		// 检查是否有完整帧
		if len(out) >= FrameLength {
			return out, nil
		}
	}
}

// KBT1000协议常量
const (
	FrameLength = 24 // 固定帧长度
	SyncWordLen = 6  // 同步字长度
	CtrlWordLen = 6  // 控制字长度
	InfoWordLen = 6  // 信息字长度
)

// SyncWord 同步字 EB 90 EB 90 EB 90
var SyncWord = []byte{0xEB, 0x90, 0xEB, 0x90, 0xEB, 0x90}

// KBTFrame KBT1000协议帧结构
type KBTFrame struct {
	SyncWord   []byte   // 同步字 6字节
	CtrlWord   []byte   // 控制字 6字节 (B7-B12)
	InfoWord1  []byte   // 信息字1 6字节 (B13-B18)
	InfoWord2  []byte   // 信息字2 6字节 (B19-B24)
	DeviceAddr byte     // 设备地址
	LineStatus [64]bool // 64路线路状态
}

// validateCtrlWord 验证控制字（B7-B12）
func validateCtrlWord(data []byte) bool {
	if len(data) != 6 {
		log.Debugf("控制字长度错误：期望6字节，实际%d字节", len(data))
		return false
	}

	// 验证固定字节
	if data[0] != 0x71 { // B7
		log.Debugf("控制字B7验证失败：期望0x71，实际0x%02X", data[0])
		return false
	}
	if data[1] != 0xF4 { // B8
		log.Debugf("控制字B8验证失败：期望0xF4，实际0x%02X", data[1])
		return false
	}

	// CRC校验 (B7-B11 -> B12)
	//calculatedCRC := calculateCRC(data[0:5])
	//if calculatedCRC != data[5] {
	//	log.Debugf("控制字CRC验证失败：期望0x%02X，实际0x%02X，数据：%X", calculatedCRC, data[5], data[0:5])
	//	return false
	//}

	log.Debugf("控制字验证成功：%X", data)
	return true
}

// validateInfoWord 验证信息字
func validateInfoWord(data []byte) bool {
	if len(data) != 6 {
		log.Debugf("信息字长度错误：期望6字节，实际%d字节", len(data))
		return false
	}

	// CRC校验 (功能码+数据域 -> 最后一个字节)
	//calculatedCRC := calculateCRC(data[0:5])
	//if calculatedCRC != data[5] {
	//	log.Debugf("信息字CRC验证失败：期望0x%02X，实际0x%02X，数据：%X", calculatedCRC, data[5], data[0:5])
	//	return false
	//}

	log.Debugf("信息字验证成功：%X", data)
	return true
}

// calculateCRC 计算CRC校验码
// 注：使用简单的异或校验
func calculateCRC(data []byte) byte {
	var crc byte = 0
	for _, b := range data {
		crc ^= b
	}
	return crc
}

// parseLineStatus 解析64路线路状态
func parseLineStatus(frame *KBTFrame) {
	// 信息字1数据 (B14-B17)
	info1Data := frame.InfoWord1[1:5]

	// 信息字2数据 (B20-B23)
	info2Data := frame.InfoWord2[1:5]

	// 解析信息字1 (线路1-32)
	for byteIdx := 0; byteIdx < 4; byteIdx++ {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			lineNum := byteIdx*8 + bitIdx + 1
			frame.LineStatus[lineNum-1] = (info1Data[byteIdx] & (1 << bitIdx)) != 0
		}
	}

	// 解析信息字2 (线路33-64)
	for byteIdx := 0; byteIdx < 4; byteIdx++ {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			lineNum := 32 + byteIdx*8 + bitIdx + 1
			frame.LineStatus[lineNum-1] = (info2Data[byteIdx] & (1 << bitIdx)) != 0
		}
	}
}
