package dianzong

import (
	"agent/logic/collector/device/model"
	"bufio"
	"context"
	"fmt"
	"time"

	serial "go.bug.st/serial"
)

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

type SerialPort struct {
	p         serial.Port
	opt       SerialOptions
	r         *bufio.Reader
	connected bool
}

// NewSerialChannel 创建串口通道（不立即打开）
func NewSerialChannel(opt SerialOptions) *SerialPort {
	return &SerialPort{
		opt:       opt,
		connected: false,
	}
}

// OpenSerial 打开串口并返回实例
func OpenSerial(opt SerialOptions) (*SerialPort, error) {
	sp := NewSerialChannel(opt)
	if err := sp.Open(); err != nil {
		return nil, err
	}
	return sp, nil
}

// Open 实现IChannel接口 - 打开串口
func (s *SerialPort) Open() error {
	if s.p != nil {
		return nil
	}
	mode := &serial.Mode{
		BaudRate: s.opt.Baud,
		DataBits: s.opt.DataBits,
		StopBits: model.MapStopBits(s.opt.StopBits),
		Parity:   model.ParseParity(s.opt.Parity),
	}
	p, err := serial.Open(s.opt.Port, mode)
	if err != nil {
		return err
	}
	_ = p.SetReadTimeout(s.opt.ReadTO)
	s.p = p
	s.r = bufio.NewReader(p)
	s.connected = true
	return nil
}

// Close 实现IChannel接口 - 关闭串口
func (s *SerialPort) Close() error {
	if s.p != nil {
		err := s.p.Close()
		s.p = nil
		s.r = nil
		s.connected = false
		return err
	}
	return nil
}

// IsOpen 实现IChannel接口 - 判断是否打开
func (s *SerialPort) IsOpen() bool {
	return s.connected && s.p != nil
}

// ReOpen 实现IChannel接口 - 重新打开
func (s *SerialPort) ReOpen() error {
	_ = s.Close()
	return s.Open()
}

// Write 实现IChannel接口 - 写数据
func (s *SerialPort) Write(data []byte) error {
	if s.p == nil {
		return fmt.Errorf("serial port not open")
	}
	for len(data) > 0 {
		n, err := s.p.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

// WriteAll 实现IChannel接口 - 写数据
func (s *SerialPort) WriteAll(b []byte) error {
	return s.Write(b)
}

// Drain 实现IChannel接口 - 清空输入缓冲
func (s *SerialPort) Drain(veryShort time.Duration, maxBytes int) []byte {
	return s.DrainInput(veryShort, maxBytes)
}

// DrainInput —— 快速排空输入缓冲（非阻塞/短等待）—— //
// 以 veryShort 为静默窗口，只要窗口内还有字节到达就继续读；
// 直到窗口内"无新数据"，或达到 maxBytes 上限（0=不限制）。
func (s *SerialPort) DrainInput(veryShort time.Duration, maxBytes int) []byte {
	if s.p == nil {
		return nil
	}
	if veryShort <= 0 {
		veryShort = 30 * time.Millisecond
	}
	// 临时把读超时设为 veryShort
	_ = s.p.SetReadTimeout(veryShort)

	var out []byte
	buf := make([]byte, 256)
	deadline := time.Now().Add(veryShort)

	for {
		n, err := s.p.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
			if maxBytes > 0 && len(out) >= maxBytes {
				break
			}
			// 只要读到数据，就把"静默截止时间"往后推
			deadline = time.Now().Add(veryShort)
			continue
		}
		// n==0: 这是一次"读超时"，看看静默窗口是否已结束
		if time.Now().After(deadline) {
			break
		}
		if err != nil {
			break
		}
		// 否则继续尝试直到 veryShort 窗口结束
	}

	// 还原为端口默认 ReadTO
	_ = s.p.SetReadTimeout(s.opt.ReadTO)
	return out
}

// PurgeInput —— 直接丢弃输入缓冲（如果底层串口驱动支持）- 未启用
func (s *SerialPort) PurgeInput() {
	if s.p == nil {
		return
	}
	if rp, ok := s.p.(interface{ ResetInputBuffer() error }); ok {
		_ = rp.ResetInputBuffer()
	} else {
		_ = s.p.SetReadTimeout(1 * time.Millisecond)
		_, _ = s.p.Read(make([]byte, 1024))
	}
}

// ReadFrame 实现IChannel接口 - 读取完整帧（按SOI/EOI标识）
func (s *SerialPort) ReadFrame(ctx context.Context) ([]byte, error) {
	if s.p == nil {
		return nil, fmt.Errorf("serial port not open")
	}
	sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// 找 SOI
	for {
		b, err := s.readByteCtx(sCtx)
		if err != nil {
			return nil, err
		}
		if b == SOI {
			break
		}
	}
	// 收集余下直到 EOI
	out := []byte{SOI}
	for {
		b, err := s.readByteCtx(sCtx)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
		if b == EOI {
			return out, nil
		}
		// 防止无限增长
		if len(out) > MaxRecvPacketLen {
			return nil, fmt.Errorf("frame too large: %d bytes", len(out))
		}
	}
}

// ReadByLength 实现IChannel接口 - 按长度读取
func (s *SerialPort) ReadByLength(ctx context.Context, firstLen int, parseRemainLen func([]byte) int) ([]byte, error) {
	if s.p == nil {
		return nil, fmt.Errorf("serial port not open")
	}

	sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 第一步：读取头部
	headBuf := make([]byte, firstLen)
	for i := 0; i < firstLen; i++ {
		b, err := s.readByteCtx(sCtx)
		if err != nil {
			return nil, fmt.Errorf("read head byte %d failed: %w", i, err)
		}
		headBuf[i] = b
	}

	// 第二步：解析剩余长度并读取
	remainLen := parseRemainLen(headBuf)
	if remainLen <= 0 {
		return headBuf, nil
	}
	if remainLen > MaxRecvPacketLen {
		return nil, fmt.Errorf("remain length too large: %d", remainLen)
	}

	remainBuf := make([]byte, remainLen)
	for i := 0; i < remainLen; i++ {
		b, err := s.readByteCtx(sCtx)
		if err != nil {
			// 允许部分读取，只要以EOI结束
			if i > 0 && remainBuf[i-1] == EOI {
				return append(headBuf, remainBuf[:i]...), nil
			}
			return nil, fmt.Errorf("read remain byte %d failed: %w", i, err)
		}
		remainBuf[i] = b
	}

	return append(headBuf, remainBuf...), nil
}

func (s *SerialPort) readByteCtx(ctx context.Context) (byte, error) {
	buf := make([]byte, 1)

	for {
		// 根据 ctx 剩余时间与串口配置 ReadTO 取较小值
		to := s.opt.ReadTO
		if deadline, ok := ctx.Deadline(); ok {
			remain := time.Until(deadline)
			if remain <= 0 {
				return 0, ctx.Err()
			}
			if to <= 0 || remain < to {
				to = remain
			}
		}
		_ = s.p.SetReadTimeout(to)

		n, err := s.p.Read(buf)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			// 一次读超时（非错误）。看 ctx 是否已取消/超期；没取消就继续读。
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			continue
		}
		return buf[0], nil
	}
}

// 确保 SerialPort 实现了 IChannel 接口
var _ IChannel = (*SerialPort)(nil)
