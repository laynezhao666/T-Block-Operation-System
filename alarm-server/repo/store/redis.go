// Package store store
package store

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"etrpc-go/client/redis"
	"etrpc-go/log"

	gredis "github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"

	"alarm-server/conf"
	"alarm-server/entity/model"
	"alarm-server/utils/common"
	"alarm-server/utils/modcall"
)

const (
	// 记录策略计算信息
	ValidRedisName      = "trpc.redis.tbos.valid" // Redis实例类常量
	ValidKeyTemplate    = "v_%d:%d"               // v_mozuId:rid
	LockMozuKeyTemplate = "lock_%s_%d"
	// Lua 脚本：检查策略运行时间并更新
	DefaultKeyDuration        = 60 * time.Second
	DefaultValidCacheInterval = 60
)

var (
	RedisStoreImpl *RedisStore
	once           sync.Once
)

// RedisStore RedisStore
type RedisStore struct {
}

// GetRedisStoreApi GetRedisStoreApi
func GetRedisStoreApi() *RedisStore {
	once.Do(func() {
		RedisStoreImpl = &RedisStore{}
	})
	return RedisStoreImpl
}

// BatchStoreRuleRecord 批量存储策略执行记录
func (v *RedisStore) BatchStoreRuleRecord(record map[string]*model.ValidStoreData) {
	now := time.Now()
	defer func() {
		modcall.RecordWriteRedisDelay(time.Since(now))
	}()
	if len(record) == 0 {
		return
	}
	recordMap := make(map[string]map[string]string)
	for _, item := range record {
		key := fmt.Sprintf(ValidKeyTemplate, item.MozuId, item.Rid)
		if _, ok := recordMap[key]; !ok {
			recordMap[key] = map[string]string{}
		}
		jsonData, err := json.Marshal(item)
		if err != nil {
			log.Errorf("json.Marshal BatchStoreRuleRecord %v, err:%s", item, err.Error())
			continue
		}
		recordMap[key][item.Gid] = string(jsonData)
	}
	batchSize := conf.ServerConf.ValidRedisCacheConfig.MSetBatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	var cnt int32 = int32(len(recordMap))
	var i int32 = 0
	cli := redis.GetRedis(ValidRedisName)
	pipe := cli.Pipeline()
	for key, fields := range recordMap {
		i++
		pipe.HSet(trpc.BackgroundContext(), key, fields)
		pipe.Expire(trpc.BackgroundContext(), key, DefaultKeyDuration)
		if i%batchSize == 0 {
			_, err := pipe.Exec(trpc.BackgroundContext())
			if err != nil {
				log.Errorf("BatchStoreRuleRecord Failed to execute pipeline: %v", err)
				return
			}
			pipe = cli.Pipeline()
		} else if i == cnt {
			_, err := pipe.Exec(trpc.BackgroundContext())
			if err != nil {
				log.Errorf("BatchStoreRuleRecord Failed to execute pipeline: %v", err)
				return
			}
		}
	}
}

// geneInvalidRecord 生成无效策略记录
// 对于没有上报/时间超时的记录，生成 策略未上报的记录
func (v *RedisStore) geneInvalidRecord(mozuId, rid int64, gid string) *model.ValidStoreData {
	return &model.ValidStoreData{
		MozuId:      int32(mozuId),
		Rid:         rid,
		Gid:         gid,
		Success:     false,
		Fired:       false,
		ErrorCode:   -1,
		ErrorName:   "策略状态未上报",
		ErrorDetail: "-",
	}
}

// BatchGetRuleRecord 批量获取策略执行记录
// @param mozuId int64
// @param keyList []int64  存储当前查询rid列表
// @param strategyCache map[int64]map[string]model.StrategyCacheData 设备缓存
// @return recordMap [mozuId][gid]record, err
func (v *RedisStore) BatchGetRuleRecord(mozuId int64, keyList []int64,
	strategyCache map[int64]*common.OrderedMap[string, model.StrategyCacheData]) (
	map[int64]map[string]*model.ValidStoreData, error) {
	now := time.Now()
	defer func() {
		modcall.RecordReadRedisDelay(mozuId, time.Since(now))
	}()
	if len(keyList) == 0 {
		return nil, nil
	}
	batchSize := conf.ServerConf.ValidRedisCacheConfig.MGetBatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	chunkList := lo.Chunk(keyList, int(batchSize))
	chunkValuesMap := make(map[int64]map[string]string)
	// 防止Redis cpu利用率过高，不使用并发查询
	for _, chunk := range chunkList {
		cgi := redis.GetRedis(ValidRedisName)
		pipe := cgi.Pipeline()
		cmds := make([]*gredis.MapStringStringCmd, len(chunk))
		for index, rid := range chunk {
			hashKey := fmt.Sprintf(ValidKeyTemplate, mozuId, rid)
			cmds[index] = pipe.HGetAll(trpc.BackgroundContext(), hashKey)
		}
		_, err := pipe.Exec(trpc.BackgroundContext())
		if err != nil {
			log.Errorf("BatchGetRuleRecord Failed to execute pipeline: %v", err)
			return nil, err
		}
		for index, cmd := range cmds {
			if cmd.Err() != nil {
				log.Warnf("Error getting HSET for mozuId:%d, rid:%d", mozuId, chunk[index])
				continue
			}
			if cmd.Val() != nil {
				chunkValuesMap[chunk[index]] = cmd.Val()
			}
		}
	}
	resRecordData := make(map[int64]map[string]*model.ValidStoreData)
	nowTs := time.Now().Unix()
	for _, rid := range keyList {
		needGidMap := strategyCache[rid]
		if needGidMap == nil {
			continue
		}
		needGidMap.Range(func(gid string, sValue model.StrategyCacheData) {
			validData := &model.ValidStoreData{}
			recordData, ok := chunkValuesMap[rid][gid]
			if !ok {
				validData = v.geneInvalidRecord(mozuId, rid, gid)
			} else {
				err := json.Unmarshal([]byte(recordData), validData)
				if err != nil {
					log.Errorf("BatchGetRuleRecord Failed to Unmarshal: %v", err)
					return
				}
				if nowTs-validData.EvalTime > DefaultValidCacheInterval {
					validData = v.geneInvalidRecord(mozuId, rid, gid)
				}
			}
			if _, ok := resRecordData[rid]; !ok {
				resRecordData[rid] = make(map[string]*model.ValidStoreData)
			}
			resRecordData[rid][gid] = validData
		})
	}
	return resRecordData, nil
}

// BatchDelRuleRecord 批量删除策略执行记录
// param mozuId int64
// param delRecord map[int64][]string map[rid][]gid
func (v *RedisStore) BatchDelRuleRecord(mozuId int64, delRecord map[int64][]string) {
	if len(delRecord) == 0 {
		return
	}
	batchSize := conf.ServerConf.ValidRedisCacheConfig.MDelBatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	var cnt int32 = int32(len(delRecord))
	var i int32 = 0
	cli := redis.GetRedis(ValidRedisName)
	pipe := cli.Pipeline()
	for rid, fields := range delRecord {
		i++
		key := fmt.Sprintf(ValidKeyTemplate, mozuId, rid)
		pipe.HDel(trpc.BackgroundContext(), key, fields...)
		if i%batchSize == 0 {
			_, err := pipe.Exec(trpc.BackgroundContext())
			if err != nil {
				log.Errorf("BatchDelRuleRecord Failed to del key: %v", err)
				return
			}
			pipe = cli.Pipeline()
		} else if i == cnt {
			_, err := pipe.Exec(trpc.BackgroundContext())
			if err != nil {
				log.Errorf("BatchDelRuleRecord Failed to del key: %v", err)
				return
			}
		}
	}
}

// TryLock 尝试获取锁
func (v *RedisStore) TryLock(service string, mozuId int64) bool {
	key := fmt.Sprintf(LockMozuKeyTemplate, service, mozuId)
	cli := redis.GetRedis(ValidRedisName)
	success, err := cli.SetNX(trpc.BackgroundContext(), key, 1, time.Minute).Result()
	if err != nil {
		log.Errorf("TryLock Failed to SetNX: %v", err)
		return false
	}
	return success
}

// UnLock 释放锁
func (v *RedisStore) UnLock(service string, mozuId int64) {
	key := fmt.Sprintf(LockMozuKeyTemplate, service, mozuId)
	cli := redis.GetRedis(ValidRedisName)
	_, err := cli.Del(trpc.BackgroundContext(), key).Result()
	if err != nil {
		log.Errorf("UnLock Failed to Del: %v", err)
	}
}
