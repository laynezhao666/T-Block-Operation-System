package osal

import (
	"agent/utils"
	"time"
)

// ExpireTimer 计时器
type ExpireTimer struct {
	lastAccessTime time.Time
	expireTime     time.Duration
}

// NewExpireTimer 创建计时器
func NewExpireTimer(expireTime time.Duration) *ExpireTimer {
	return &ExpireTimer{
		lastAccessTime: time.Unix(0, 0),
		expireTime:     expireTime,
	}
}

// IsExpired 判断是否过期
func (e *ExpireTimer) IsExpired() bool {
	if e == nil {
		return true
	}
	return utils.GetNowLocalTime().After(e.lastAccessTime.Add(e.expireTime))
}

// SetAccess 设置访问时间
func (e *ExpireTimer) SetAccess() {
	if e == nil {
		return
	}
	e.lastAccessTime = utils.GetNowLocalTime()
}
