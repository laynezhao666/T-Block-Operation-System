package sockclient

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

const (
	defaultTimeout        = time.Second * 10
	defaultOnceReadMaxLen = 1024 * 1024 * 10 // 10M
)

// Transport socket transport
type Transport struct {
	host    string
	timeout time.Duration
	conn    net.Conn
}

// NewTransport new transport
func NewTransport(host string, timeout time.Duration) *Transport {
	if timeout < 1 {
		timeout = defaultTimeout
	}

	return &Transport{
		host:    host,
		timeout: timeout,
	}
}

// Connect connect
func (t *Transport) Connect() error {
	return t.connect()
}

// Close close
func (t *Transport) Close() {
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
}

// Write write
func (t *Transport) Write(b []byte) (int, error) {
	if t.conn == nil {
		if err := t.connect(); err != nil {
			return -1, err
		}
	}
	if err := t.conn.SetWriteDeadline(time.Now().Add(t.timeout)); err != nil {
		return -1, err
	}

	n, err := t.conn.Write(b)
	if err != nil {
		t.dealNetError(err)
	}
	return n, err

}

// ReadAtLeast read at least
func (t *Transport) ReadAtLeast(size int) ([]byte, error) {
	if t.conn == nil {
		if err := t.connect(); err != nil {
			return nil, err
		}
	}

	if err := t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, err
	}

	b := make([]byte, size)
	var n int
	for n < len(b) {
		onceLen, err := t.conn.Read(b[n:])
		if err != nil {
			t.dealNetError(err)
			return b, err
		}
		n += onceLen
	}

	return b, nil
}

// ReadWithEndChars read with end chars
func (t *Transport) ReadWithEndChars(headChars string, tailChars string) ([]byte, error) {
	if t.conn == nil {
		if err := t.connect(); err != nil {
			return nil, err
		}
	}

	if err := t.conn.SetReadDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, err
	}

	totalBytes := make([]byte, 0, defaultOnceReadMaxLen)

	for {
		b := make([]byte, defaultOnceReadMaxLen)
		onceLen, err := t.conn.Read(b)
		if err != nil {
			t.dealNetError(err)
			return totalBytes, err
		}

		totalBytes = append(totalBytes, b[:onceLen]...)
		if !bytes.HasPrefix(totalBytes, []byte(headChars)) {
			t.Close()
			return nil, errors.New("not found head chars")
		}
		if bytes.HasSuffix(totalBytes, []byte(tailChars)) {
			break
		}
	}

	return totalBytes, nil
}

func (t *Transport) connect() error {
	var err error
	if t.conn, err = net.Dial("tcp", t.host); err != nil {
		return err
	}
	return nil
}

func (t *Transport) dealNetError(err error) {
	var netError net.Error
	if err == io.EOF || (errors.As(err, &netError) && netError.Timeout()) {
		if t.conn != nil {
			_ = t.conn.Close()
			t.conn = nil
		}
	}
}
