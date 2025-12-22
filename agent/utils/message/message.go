// Package message 消息
package message

import (
	"fmt"
	"agent/utils/osal"
	"time"
)

// MethodType 方法类型
type MethodType int

// String MethodType 转字符串
func (t MethodType) String() string {
	switch t {
	case MethodInvalid:
		return "MethodInvalid"
	case MethodGet:
		return "MethodGet"
	case MethodPost:
		return "MethodPost"
	case MethodPut:
		return "MethodPut"
	case MethodDelete:
		return "MethodDelete"
	default:
		return fmt.Sprintf("[MethodUnknown=%d]", t)
	}
}

const (
	MethodInvalid MethodType = -1
	MethodGet     MethodType = 0
	MethodPost    MethodType = 2
	MethodPut     MethodType = 3
	MethodDelete  MethodType = 4
)

// TopicType 类型
type TopicType int

const (
	TopicInvalid      TopicType = -1
	TopicDevice       TopicType = 0
	TopicPointControl TopicType = 1
	TopicTemplate     TopicType = 2
)

// String TopicType 转字符串
func (t TopicType) String() string {
	switch t {
	case TopicInvalid:
		return "TopicInvalid"
	case TopicDevice:
		return "TopicDevice"
	case TopicPointControl:
		return "TopicPointControl"
	case TopicTemplate:
		return "TopicTemplate"
	default:
		return fmt.Sprintf("[TopicUnknown=%d]", t)
	}
}

// PatternType 模式类型
type PatternType int

const (
	PatternInvalid PatternType = -1
	PatternNotice  PatternType = 0
	PatternReqRep  PatternType = 1
)

// String PatternType 转字符串
func (p PatternType) String() string {
	switch p {
	case PatternInvalid:
		return "PatternInvalid"
	case PatternNotice:
		return "PatternNotice"
	case PatternReqRep:
		return "PatternReqRep"
	default:
		return fmt.Sprintf("[PatternUnknown=%d]", p)
	}
}

// IMessage 消息接口
type IMessage interface {
	Method() MethodType
	Topic() TopicType
	Pattern() PatternType
	String() string
}

type iMessage struct {
	method  MethodType
	topic   TopicType
	pattern PatternType
}

// String iMessage 转字符串
func (m iMessage) String() string {
	return fmt.Sprintf("{method: %v, topic: %v, pattern: %v}", m.method, m.topic, m.pattern)
}

// Method MethodType
func (m iMessage) Method() MethodType {
	return m.method
}

// Topic TopicType
func (m iMessage) Topic() TopicType {
	return m.topic
}

// Pattern PatternType
func (m iMessage) Pattern() PatternType {
	return m.pattern
}

// NoticeMessage 通知消息
type NoticeMessage struct {
	iMessage
}

// NewNoticeMessage 返回新通知消息实例
func NewNoticeMessage(topic TopicType, method MethodType) *NoticeMessage {
	return &NoticeMessage{
		iMessage{
			method:  method,
			topic:   topic,
			pattern: PatternNotice,
		},
	}
}

type reqRepMessage struct {
	iMessage
	sem *osal.Semaphore
}

// Post 发送消息
func (m *reqRepMessage) Post() {
	if m == nil {
		return
	}
	m.sem.Post()
}

// Wait 等待消息
func (m *reqRepMessage) Wait(t time.Duration) bool {
	if m == nil {
		return false
	}
	return m.sem.Wait(t)
}

// NewReqRepMessage 返回新请求消息实例
func NewReqRepMessage(topic TopicType, method MethodType) *reqRepMessage {
	return &reqRepMessage{
		iMessage{
			method:  method,
			topic:   topic,
			pattern: PatternReqRep,
		},
		osal.NewSemaphore(1),
	}
}
