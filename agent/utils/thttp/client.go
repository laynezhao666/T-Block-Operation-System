package thttp

import (
	"net/http"
	"sync"
	"time"
)

var (
	clientMap = map[int]*http.Client{
		0: {Timeout: 30 * time.Second},
	}
	mutex sync.Mutex
)

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
