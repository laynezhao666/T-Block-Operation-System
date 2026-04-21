// Package dtcp 提供TCP连接的底层读写工具函数。
package dtcp

import (
	"fmt"
	"io"
	"net"
	"time"
)

// WriteN 向TCP连接写入指定数据，支持超时控制。
// 返回实际写入的字节数和可能的错误。
func WriteN(conn net.Conn, data []byte,
	timeout time.Duration,
) (int, error) {
	if conn == nil {
		return -1, fmt.Errorf("conn is nil")
	}

	var err error
	if timeout == 0 {
		err = conn.SetWriteDeadline(time.Time{})
	} else {
		err = conn.SetWriteDeadline(time.Now().Add(timeout))
	}
	if err != nil {
		return -1, err
	}

	totalLen := len(data)
	remainLen := len(data)
	currentWrite := 0

	// 循环写入直到所有数据发送完毕
	for remainLen > 0 {
		currentWrite, err = conn.Write(data)
		if err != nil {
			break
		}
		data = data[currentWrite:]
		remainLen -= currentWrite
	}
	err = conn.SetWriteDeadline(time.Time{})
	return totalLen - remainLen, err
}

// ReadN 从TCP连接读取指定字节数的数据，支持超时控制
func ReadN(conn net.Conn, n int,
	timeout time.Duration,
) ([]byte, error) {
	if conn == nil {
		return nil, fmt.Errorf("conn is nil")
	}

	var err error
	if timeout == 0 {
		err = conn.SetReadDeadline(time.Time{})
	} else {
		err = conn.SetReadDeadline(time.Now().Add(timeout))
	}
	if err != nil {
		return nil, err
	}

	buf := make([]byte, n)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}

	err = conn.SetReadDeadline(time.Time{})
	return buf, err
}

// Flush 清空TCP连接缓冲区中的残留数据
func Flush(conn net.Conn, timeout time.Duration) {
	for {
		if _, err := ReadN(conn, 1, timeout); err != nil {
			break
		}
	}
}

// Close 关闭TCP连接
func Close(conn net.Conn) {
	_ = conn.Close()
}

// FlushAndClose 先清空缓冲区再关闭TCP连接
func FlushAndClose(conn net.Conn, timeout time.Duration) {
	if conn == nil {
		return
	}
	Flush(conn, timeout)
	Close(conn)
}
