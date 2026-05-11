package adam

import (
	"agent/logic/collector/device/model"
	"bufio"
	"context"
	"time"

	serial "go.bug.st/serial"
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

// ReadUntilCR 读到 <CR> (0x0D) 或 ctx 超时/取消
func (s *SerialPort) ReadUntilCR(ctx context.Context, max int) ([]byte, error) {
	var out []byte
	for {
		// 根据 ctx 动态设定一次 ReadTimeout
		to := s.opt.ReadTO
		if dl, ok := ctx.Deadline(); ok {
			if remain := time.Until(dl); remain > 0 && (to <= 0 || remain < to) {
				to = remain
			}
		}
		_ = s.p.SetReadTimeout(to)

		b := []byte{0}
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
		out = append(out, b[0])
		if b[0] == 0x0D {
			return out, nil
		}
		if max > 0 && len(out) >= max {
			return out, nil
		}
	}
}
