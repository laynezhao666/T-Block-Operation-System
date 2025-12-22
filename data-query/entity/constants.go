package entity

const (
	StoreRedisName = "trpc.redis.idc.test.store" // Redis实例类常量

	RedisChangedPointsKey = "ChangedPoints" // Redis中存储变更点信息的key

	ReadInfluxType                      = "InfluxDB"
	ReadCacheType                       = "Cache"
	DefaultQueryChangedBatchSize        = 500
	DefaultQueryChangedConcurrencyLimit = 10
)
