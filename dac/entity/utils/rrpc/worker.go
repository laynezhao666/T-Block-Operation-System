// Package rrpc 提供请求-响应式RPC管理器，支持异步等待响应。
package rrpc

import (
	"context"
	"errors"
	"time"

	"dac/entity/utils/rrpc/plugins"
	"github.com/google/uuid"
)

// readWriteCloser 消息代理的读写关闭接口
type readWriteCloser interface {
	Write(id, msgID string, payload []byte) error
	Read() (string, []byte, error)
	Close() error
}

// Client RRPC客户端，封装消息代理的读写操作
type Client struct {
	b   readWriteCloser
	ctx context.Context
}

// NewClient 创建一个新的 rrpc client，注意不要重复创建相同的 client，其它参数设定如下
// 当 plugin = "kafka" 时
//
//	args[0] = []string, 表示 brokers
//	args[1] = string, 表示写入请求消息到该 topic
//	args[2] = string, 表示从该 topic 读取响应消息
func NewClient(ctx context.Context,
	plugin string, args ...interface{},
) (*Client, error) {
	var p readWriteCloser = nil
	var err error
	switch plugin {
	case "kafka":
		brokers := args[0].([]string)
		reqTopic := args[1].(string)
		respTopic := args[2].(string)
		p, err = plugins.NewKafkaProxy(
			ctx, brokers, reqTopic, respTopic)
	default:
		err = errors.New("unsupported plugin")
	}
	c := &Client{b: p, ctx: ctx}
	if err == nil {
		go c.read()
	}
	return c, err
}

// Request 发送RRPC请求并等待响应，超时返回错误
func (c *Client) Request(clientID string,
	payload []byte, timeout time.Duration,
) ([]byte, error) {
	msgID := uuid.NewString()
	err := c.b.Write(clientID, msgID, payload)
	if err != nil {
		return nil, err
	}
	v, ok := Manager().Get(msgID, timeout)
	if !ok {
		return nil, errors.New("timeout")
	}
	b, ok := v.([]byte)
	if !ok {
		return nil, errors.New("类型断言失败")
	}
	return b, nil
}

// read 后台协程，持续读取响应消息并通知等待方
func (c *Client) read() {
	defer func() {
		_ = c.b.Close()
	}()

	var msgID string
	var payload []byte
	var err error

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		if msgID, payload, err = c.b.Read(); err != nil {
			time.Sleep(time.Second)
			continue
		}

		Manager().Set(msgID, payload)
	}
}
