package thttp

import (
	"net/http"
	"sync"
	"time"
)

// clientMap HTTP客户端缓存池，key为超时时间（毫秒），0表示默认30秒
// mutex 保护clientMap的并发访问
var (
	clientMap = map[int]*http.Client{
		0: {Timeout: 30 * time.Second},
	}
	mutex sync.Mutex
)

// getClient 根据超时时间获取或创建HTTP客户端。
// 使用缓存池避免重复创建，timeout单位为毫秒，0表示使用默认超时。
func getClient(timeout int) *http.Client {
	mutex.Lock()
	defer mutex.Unlock()

	c, ok := clientMap[timeout]
	if ok {
		return c
	}

	c = &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	clientMap[timeout] = c
	return c
}
