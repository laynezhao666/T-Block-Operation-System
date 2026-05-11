package dianzong

import (
	"context"
	"fmt"
	"net"
	"time"
)

// IChannel 通道接口，抽象串口和TCP的通信方式
type IChannel interface {
	// Open 打开通道
	Open() error
	// Close 关闭通道
	Close() error
	// IsOpen 判断通道是否打开
	IsOpen() bool
	// Write 写数据
	Write(data []byte) error
	// ReadFrame 读取完整帧（按SOI/EOI标识）
	ReadFrame(ctx context.Context) ([]byte, error)
	// ReadByLength 按长度读取（先读头部，再读剩余）
	ReadByLength(ctx context.Context, firstLen int, parseRemainLen func([]byte) int) ([]byte, error)
	// Drain 清空输入缓冲
	Drain(timeout time.Duration, maxBytes int) []byte
	// ReOpen 重新打开连接
	ReOpen() error
}

// TCPChannel TCP通道实现
type TCPChannel struct {
	address   string // IP:Port
	timeout   time.Duration
	conn      net.Conn
	connected bool
}

// TCPOptions TCP选项
type TCPOptions struct {
	Address string        // IP:Port
	Timeout time.Duration // 读写超时
}

// NewTCPChannel 创建TCP通道
func NewTCPChannel(opt TCPOptions) *TCPChannel {
	if opt.Timeout <= 0 {
		opt.Timeout = 3 * time.Second
	}
	return &TCPChannel{
		address: opt.Address,
		timeout: opt.Timeout,
	}
}

// Open 打开TCP连接
func (t *TCPChannel) Open() error {
	if t.conn != nil {
		return nil
	}
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return fmt.Errorf("tcp dial failed: %w", err)
	}
	t.conn = conn
	t.connected = true
	return nil
}

// Close 关闭TCP连接
func (t *TCPChannel) Close() error {
	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		t.connected = false
		return err
	}
	return nil
}

// IsOpen 判断连接是否打开
func (t *TCPChannel) IsOpen() bool {
	return t.connected && t.conn != nil
}

// ReOpen 重新打开连接
func (t *TCPChannel) ReOpen() error {
	_ = t.Close()
	return t.Open()
}

// Write 写数据
func (t *TCPChannel) Write(data []byte) error {
	if t.conn == nil {
		return fmt.Errorf("tcp connection not open")
	}
	_ = t.conn.SetWriteDeadline(time.Now().Add(t.timeout))
	defer t.conn.SetWriteDeadline(time.Time{})

	for len(data) > 0 {
		n, err := t.conn.Write(data)
		if err != nil {
			t.connected = false
			return err
		}
		data = data[n:]
	}
	return nil
}

// ReadFrame 按SOI/EOI标识读取完整帧
func (t *TCPChannel) ReadFrame(ctx context.Context) ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("tcp connection not open")
	}

	readCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 找SOI
	for {
		b, err := t.readByteCtx(readCtx)
		if err != nil {
			return nil, err
		}
		if b == SOI {
			break
		}
	}

	// 收集直到EOI
	out := []byte{SOI}
	for {
		b, err := t.readByteCtx(readCtx)
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

// ReadByLength 按长度读取（先读firstLen字节头部，再解析剩余长度读取）
func (t *TCPChannel) ReadByLength(ctx context.Context, firstLen int, parseRemainLen func([]byte) int) ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("tcp connection not open")
	}

	readCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 第一步：读取头部
	headBuf := make([]byte, firstLen)
	n, err := t.readFullCtx(readCtx, headBuf)
	if err != nil {
		return nil, fmt.Errorf("read head failed: %w", err)
	}
	if n != firstLen {
		return nil, fmt.Errorf("read head incomplete: got %d, want %d", n, firstLen)
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
	n, err = t.readFullCtx(readCtx, remainBuf)
	if err != nil {
		// 允许部分读取，只要以EOI结束
		if n > 0 && remainBuf[n-1] == EOI {
			return append(headBuf, remainBuf[:n]...), nil
		}
		return nil, fmt.Errorf("read remain failed: %w", err)
	}

	return append(headBuf, remainBuf[:n]...), nil
}

// Drain 清空输入缓冲
func (t *TCPChannel) Drain(timeout time.Duration, maxBytes int) []byte {
	if t.conn == nil || timeout <= 0 {
		return nil
	}
	if maxBytes <= 0 {
		maxBytes = 4096
	}

	_ = t.conn.SetReadDeadline(time.Now().Add(timeout))
	defer t.conn.SetReadDeadline(time.Time{})

	var out []byte
	buf := make([]byte, 256)
	deadline := time.Now().Add(timeout)

	for {
		n, err := t.conn.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
			if len(out) >= maxBytes {
				break
			}
			deadline = time.Now().Add(timeout)
			continue
		}
		if time.Now().After(deadline) {
			break
		}
		if err != nil {
			break
		}
	}
	return out
}

// readByteCtx 带context读取单字节
func (t *TCPChannel) readByteCtx(ctx context.Context) (byte, error) {
	buf := make([]byte, 1)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		// 设置短超时轮询
		to := 100 * time.Millisecond
		if deadline, ok := ctx.Deadline(); ok {
			remain := time.Until(deadline)
			if remain <= 0 {
				return 0, ctx.Err()
			}
			if remain < to {
				to = remain
			}
		}
		_ = t.conn.SetReadDeadline(time.Now().Add(to))

		n, err := t.conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			t.connected = false
			return 0, err
		}
		if n == 1 {
			return buf[0], nil
		}
	}
}

// readFullCtx 带context读取指定长度
func (t *TCPChannel) readFullCtx(ctx context.Context, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		to := 100 * time.Millisecond
		if deadline, ok := ctx.Deadline(); ok {
			remain := time.Until(deadline)
			if remain <= 0 {
				return total, ctx.Err()
			}
			if remain < to {
				to = remain
			}
		}
		_ = t.conn.SetReadDeadline(time.Now().Add(to))

		n, err := t.conn.Read(buf[total:])
		if n > 0 {
			total += n
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			t.connected = false
			return total, err
		}
	}
	return total, nil
}

// 常量定义（TCP相关）
const (
	FirstRecvLen        = 13        // 首次接收长度（SOI + VER + ADR + CID1 + RTN + LENGTH(4) + 部分INFO）
	MaxRecvPacketLen    = 18 + 4095 // 最大接收包长度
	LengthOffsetInASCII = 9         // LENGTH字段在ASCII帧中的偏移（从SOI之后算起）
)

// ParseRemainLenFromHead 从头部解析剩余需要读取的长度
// 头部格式: SOI(1) + VER(2) + ADR(2) + CID1(2) + RTN(2) + LENGTH(4) = 13 ASCII字节
// LENGTH = LCHKSUM(1nibble) + LENID(3nibbles)，LENID是INFO的ASCII长度
func ParseRemainLenFromHead(head []byte) int {
	if len(head) < FirstRecvLen {
		return 0
	}
	// LENGTH字段在偏移9-12（4个ASCII字符）
	// 解析LENGTH字段获取LENID
	length, err := readHexU16(head[9:13])
	if err != nil {
		return 0
	}
	lenID := length & 0x0FFF
	// 剩余长度 = LENID(INFO的ASCII长度) + CHKSUM(4) + EOI(1)
	return int(lenID) + 5
}

// 确保 TCPChannel 实现了 IChannel 接口
var _ IChannel = (*TCPChannel)(nil)
