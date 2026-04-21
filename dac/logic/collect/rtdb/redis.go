// Package rtdb 提供门禁测点数据的Redis读写功能。
package rtdb

import (
	"context"
	"fmt"
	"strconv"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/rt"
	"dac/entity/redis"
	"dac/entity/utils"
	redis2 "dac/repo/redis"

	redis3 "github.com/redis/go-redis/v9"
)

// getQua 从Redis返回值中解析测点质量
func getQua(v interface{}) consts.Quality {
	switch x := v.(type) {
	case int:
		return consts.Quality(x)
	case string:
		if n, err := strconv.ParseInt(x, 0, 64); err == nil {
			return consts.Quality(n)
		}
	}
	return consts.QualityUncertain
}

// getTimestamp 从Redis返回值中解析时间戳
func getTimestamp(v interface{}) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		if n, err := strconv.ParseInt(x, 0, 64); err == nil {
			return n
		}
	}
	return 0
}

// GetRedisPoints 从Redis批量读取测点的实时值、质量和时间戳
func GetRedisPoints(ctx context.Context, points rt.Points) error {
	pipeline := redis.GetClient().TxPipeline()
	defer func() {
		pipeline.Discard()
	}()

	for i := range points {
		p := &points[i]

		_ = pipeline.HMGet(ctx, utils.GenerateFullPointID(p.ID), consts.KeyValue, consts.KeyQua, consts.KeyTimestamp)
	}

	results, err := redis2.ExecPipeline(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("exec pipeline error: %w", err)
	}

	for i := range results {
		p := &points[i]

		p.Rtd = rt.NewRTValue()

		r, ok := results[i].(*redis3.SliceCmd)
		if !ok {
			continue
		}
		values := r.Val()
		if v, ok := values[0].(string); ok {
			p.Rtd.Pv = v
		}
		p.Rtd.Qua = getQua(values[1])
		p.Rtd.Timestamp = getTimestamp(values[2])
	}

	return nil
}

// SetRedisPoints 将测点数据批量写入Redis（包含实时值、质量和时间戳）
func SetRedisPoints(ctx context.Context, points rt.Points) {
	pipeline := redis.GetClient().TxPipeline()
	defer func() {
		pipeline.Discard()
	}()

	for i := range points {
		p := &points[i]

		_ = pipeline.HMSet(ctx,
			utils.GenerateFullPointID(p.ID),
			consts.KeyValue, p.Rtd.Pv,
			consts.KeyQua, int(p.Rtd.Qua),
			consts.KeyTimestamp, p.Rtd.Timestamp)
	}

	_, err := redis2.ExecPipeline(ctx, pipeline)
	if err != nil {
		config.Log.Warnf("SetPoints exec pipeline error: %v", err)
	}
}
