// Package redis 提供Redis操作的封装函数。
package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ExecPipeline 执行Redis管道中的所有命令并返回结果。
func ExecPipeline(ctx context.Context, pipeliner redis.Pipeliner) ([]redis.Cmder, error) {
	return pipeliner.Exec(ctx)
}

// MGet 批量获取多个key的值。
func MGet(ctx context.Context, client redis.UniversalClient, keys ...string) ([]interface{}, error) {
	return client.MGet(ctx, keys...).Result()
}

// HMSet 批量设置Hash字段的值。
func HMSet(ctx context.Context, client redis.UniversalClient, key string, fvs ...interface{}) error {
	return client.HMSet(ctx, key, fvs...).Err()
}

// HMGet 批量获取Hash中指定字段的值。
func HMGet(ctx context.Context, client redis.UniversalClient, key string, fields ...string) ([]interface{}, error) {
	return client.HMGet(ctx, key, fields...).Result()
}

// HDel 删除Hash中的指定字段。
func HDel(ctx context.Context, client redis.UniversalClient, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

// HKeys 获取Hash中所有字段名。
func HKeys(ctx context.Context, client redis.UniversalClient, key string) ([]string, error) {
	return client.HKeys(ctx, key).Result()
}
