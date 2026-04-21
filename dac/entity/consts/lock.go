// Package consts 定义门禁系统的全局常量。
package consts

import "time"

// RedisLockName 分布式锁名称
// RedisLockExpireTime 分布式锁过期时间
// RedisLockExtendTime 分布式锁续期时间
// AliveMarkExpireTime 存活标记过期时间
const (
	RedisLockName = "tdac.worker"

	RedisLockExpireTime = 10 * time.Second
	RedisLockExtendTime = RedisLockExpireTime / 2
	AliveMarkExpireTime = 30 * time.Second
)
