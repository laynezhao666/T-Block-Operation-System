// Package dislock 存放分布式锁函数
package dislock

import (
	"context"
	tredis "etrpc-go/client/redis"
	"time"

	"github.com/pkg/errors"
	"trpc.group/trpc-go/trpc-database/goredis/redlock"
	"trpc.group/trpc-go/trpc-go/log"
)

// DisLock 分布式锁，简化每次调用需要写大量锁逻辑
func DisLock(ctx context.Context, redisName, key string, handler func(), options ...redlock.Option) error {
	// 1、获取分布式锁实例
	lockCli, err := redlock.New(tredis.GetRedis(redisName))
	if err != nil {
		return errors.Wrapf(err, "create redis lock obj fail")
	}
	opts := []redlock.Option{redlock.WithKeyExpiration(60 * time.Second)}
	opts = append(opts, options...)
	// 2、尝试加锁，默认过期时间为60s，如需更长可通过options参数调整
	lock, err := lockCli.TryLock(ctx, key, opts...)
	if err != nil {
		return errors.Wrapf(err, "get redis lock fail, key: %s", key)
	}
	// 3、defer释放锁
	defer func(lock redlock.Mutex, ctx context.Context) {
		err := lock.Unlock(ctx)
		if err != nil {
			log.ErrorContextf(ctx, "release redis lock fail, key: %s, err: %v", key, err)
		}
	}(lock, ctx)
	// 4、执行业务逻辑
	handler()
	return nil
}
