// Package redis simply go-redis use
package redis

import (
	"github.com/redis/go-redis/v9"
	"sync"
	"trpc.group/trpc-go/trpc-database/goredis"
	"trpc.group/trpc-go/trpc-go/client"
)

var (
	redisMapLock sync.RWMutex
	redisMap     = make(map[string]redis.UniversalClient)
)

// GetRedis 获取Redis客户端实例
func GetRedis(name string, opts ...client.Option) redis.UniversalClient {
	if cli, ok := redisMap[name]; ok {
		return cli
	}
	cli, err := NewClientProxy(name, opts...)
	if err != nil {
		panic(err)
	}
	return cli
}

// NewClientProxy 创建Redis客户端实例
func NewClientProxy(name string, opts ...client.Option) (redis.UniversalClient, error) {
	redisMapLock.Lock()
	defer redisMapLock.Unlock()
	cli, err := goredis.New(name, opts...)
	if err != nil {
		return nil, err
	}
	redisMap[name] = cli
	return cli, nil
}
