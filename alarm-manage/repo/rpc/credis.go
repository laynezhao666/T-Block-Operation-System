package rpc

import (
	"sync"
	"time"

	"etrpc-go/client/redis"
	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"
)

const (
	AlarmRedisName = "trpc.redis.tbos.alert" // Redis实例类常量
)

var (
	RedisStoreImpl *RedisStore
	rOnce          sync.Once
)

// RedisStore RedisStore
type RedisStore struct {
}

// NewRedisApi NewRedisApi
func NewRedisApi() *RedisStore {
	rOnce.Do(func() {
		RedisStoreImpl = &RedisStore{}
	})
	return RedisStoreImpl
}

// TryLock 尝试获取锁
func (v *RedisStore) TryLock(lockKey string, timeDuration time.Duration) bool {
	cli := redis.GetRedis(AlarmRedisName)
	success, err := cli.SetNX(trpc.BackgroundContext(), lockKey, 1, timeDuration).Result()
	if err != nil {
		log.Errorf("TryLock Failed to SetNX: %v", err)
		return false
	}
	return success
}

// UnLock 释放锁
func (v *RedisStore) UnLock(lockKey string) {
	cli := redis.GetRedis(AlarmRedisName)
	_, err := cli.Del(trpc.BackgroundContext(), lockKey).Result()
	if err != nil {
		log.Errorf("UnLock Failed to Del: %v", err)
	}
}
