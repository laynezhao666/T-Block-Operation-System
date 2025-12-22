// Package timezset 基于Redis的zset实现的时间序列查询接口工具类
package timezset

import (
	"context"
	"etrpc-go/util/arrayutil"
	"fmt"
	"github.com/redis/go-redis/v9"

	"time"
)

// TimeZSet Redis的ZSet实现缓存时序数据,ItemExpire可控制缓存的周期
type TimeZSet struct {
	RedisName  string
	Key        string
	ItemExpire time.Duration

	ScoreEncoder func(time.Time) int64
	ScoreDecoder func(int64) time.Time
	redisCli     redis.UniversalClient
}

// DataItem 数据项定义
type DataItem struct {
	Score time.Time
	Data  any
}

// New 新建一个TimeZSet, 可以设置对应的key/过期时间
func New(cli redis.UniversalClient, key string, itemExpire time.Duration) *TimeZSet {
	return &TimeZSet{
		Key:        key,
		ItemExpire: itemExpire,

		ScoreEncoder: func(t time.Time) int64 { return t.Unix() },
		ScoreDecoder: func(t int64) time.Time { return time.Unix(t, 0) },
		redisCli:     cli,
	}
}

// Add 向有序集合新增数据,注意score为时间unix值, member同名时会更新score, 请根据应用场景设置member
func (z *TimeZSet) Add(ctx context.Context, data []DataItem) error {
	// convert data to zSet items
	members := arrayutil.Map(data, func(d DataItem) redis.Z {
		return redis.Z{
			Score:  float64(z.ScoreEncoder(d.Score)),
			Member: d.Data,
		}
	})
	// add new items, if new items contain expire item, remove it last time
	if err := z.redisCli.ZAdd(ctx, z.Key, members...).Err(); err != nil {
		return err
	}
	return nil
}

// RemoveExpire 移除过期的数据,zSet无法为每个item设置过期时间，需主动移除
func (z *TimeZSet) RemoveExpire(ctx context.Context) error {
	// remove all expire items
	minScore := fmt.Sprintf("(%d", z.ScoreEncoder(z.getMinValidTime()))
	return z.redisCli.ZRemRangeByScore(ctx, z.Key, "-inf", minScore).Err()
}

// GetByTimeRange 查询key对应的数据，根据时间范围查询
func (z *TimeZSet) GetByTimeRange(ctx context.Context, begin, end time.Time) ([]DataItem, error) {
	// valid query range is valid, if query expire time range, fail
	minValidTime := z.getMinValidTime()
	if begin.Before(minValidTime) || end.Before(minValidTime) {
		return nil, fmt.Errorf("query data out of time range, minValidTime is %s", minValidTime.String())
	}
	rangeBy := &redis.ZRangeBy{
		Min: fmt.Sprintf("[%d", z.ScoreEncoder(begin)),
		Max: fmt.Sprintf("[%d", z.ScoreEncoder(end)),
	}
	return z.queryData(ctx, rangeBy)
}

// GetAll 查询key对应的数据，查询所有未过期的
func (z *TimeZSet) GetAll(ctx context.Context) ([]DataItem, error) {
	rangeBy := &redis.ZRangeBy{
		Min: fmt.Sprintf("(%d", z.ScoreEncoder(z.getMinValidTime())),
		Max: "+inf",
	}
	return z.queryData(ctx, rangeBy)
}

func (z *TimeZSet) queryData(ctx context.Context, rangeBy *redis.ZRangeBy) ([]DataItem, error) {
	// do query data
	members, err := z.redisCli.ZRangeByScoreWithScores(ctx, z.Key, rangeBy).Result()
	if err != nil {
		return nil, err
	}
	// convert redis member to DataItem
	return arrayutil.Map(members, func(m redis.Z) DataItem {
		return DataItem{
			Score: z.ScoreDecoder(int64(m.Score)),
			Data:  m.Member,
		}
	}), nil
}

func (z *TimeZSet) getMinValidTime() time.Time {
	return time.Now().Add(-z.ItemExpire)
}
