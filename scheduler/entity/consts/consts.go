// Package consts 存储常量信息
package consts

const (
	TBosRedisName = "trpc.redis.tbos" // Redis实例类常量
	TBosMySQLName = "trpc.mysql.tbos" // MySQL实例类常量

	MySQLFetchBatchSize = 50000 // 数据库读取数据量批次大小
	RedisJoinFieldSep   = "#"   // Redis拼接多个字段的分隔符

	CommonFieldSeq = "#" // 通用的字段分隔符
)
