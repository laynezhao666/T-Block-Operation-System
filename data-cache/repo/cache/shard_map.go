// Package cache 支持分片并发安全的map
package cache

import (
	"sync"

	"github.com/spaolacci/murmur3"
)

// DefaultHash 默认的hash函数
func DefaultHash(key string) uint64 {
	code := murmur3.Sum64([]byte(key))
	return code
}

// ShardMap 分片的并发安全的map
type ShardMap[K comparable, V any] struct {
	shardCnt uint64
	hashFun  func(K) uint64
	shards   []*sync.Map
}

// NewShardMap 创建一个分片Map
func NewShardMap[K comparable, V any](shardCnt uint64, hashFunc func(K) uint64) *ShardMap[K, V] {
	shards := make([]*sync.Map, shardCnt)
	for i := range shardCnt {
		shards[i] = &sync.Map{}
	}
	return &ShardMap[K, V]{
		shardCnt: shardCnt,
		hashFun:  hashFunc,
		shards:   shards,
	}
}

// getShard 根据key获取分片对象
func (m *ShardMap[K, V]) getShard(key K) *sync.Map {
	return m.shards[m.hashFun(key)%m.shardCnt]
}

// Get 获取单个Key的值
func (m *ShardMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.getShard(key).Load(key)
	if !ok {
		var zero V
		return zero, ok
	}
	return v.(V), ok
}

// GetMany 批量获取多个Key的值
func (m *ShardMap[K, V]) GetMany(keys []K) map[K]V {
	res := make(map[K]V, len(keys))
	for _, key := range keys {
		if val, ok := m.getShard(key).Load(key); ok {
			res[key] = val.(V)
		}
	}
	return res
}

// SetNx 设置一个K,V，存在则返回旧的,不存在则设置值并返回新的
func (m *ShardMap[K, V]) SetNx(key K, val V) V {
	actual, _ := m.getShard(key).LoadOrStore(key, val)
	return actual.(V)
}

// Keys 返回所有缓存的Key
func (m *ShardMap[K, V]) Keys() []K {
	var keys []K
	for _, shard := range m.shards {
		shard.Range(func(key, _ any) bool {
			keys = append(keys, key.(K))
			return true
		})
	}
	return keys
}
