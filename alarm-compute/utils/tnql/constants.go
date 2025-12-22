package tnql

const (
	// SecondsInDay 一天 86400 秒
	SecondsInDay int = 86400

	// SecondsInDay 一天 86400 秒
	MinutesInDay int = 1440

	// SecondsInMinute 一分钟 60 秒
	SecondsInMinute int = 60
	// ExprCustomDataKeyData ExprCustomDataKeyData
	ExprCustomDataKeyData string = "data"

	// ExprCustomDataKeyDelayConfig ExprCustomDataKeyDelayConfig
	// ExprCustomDataKeyDelayConfig string = "delayConfig"

	// ExprCustomDataKeyDryRun 用于获取函数参数，不进行计算
	ExprCustomDataKeyDryRun string = "dryrun"

	// ExprCustomDataKeySimple 简单表达式，即没有函数，只使用 kv 值进行计算的数据（不使用历史数据），等同于原来的实时测点计算
	ExprCustomDataKeySimple string = "simple"

	// ExprNoDelay 不需要获取 Delay 数据，直接用当前数据
	ExprNoDelay int = 0

	// ResultTypeBoolean = 1

	// DefaultInterval 默认间隔
	DefaultInterval int = 1
	// MaxHBasePointFetchNum 单次请求 hbase 最大测点总数 （测点数 * 时间）
	MaxHBasePointFetchNum int = 10000

	// JpArgsNum 跳变参数个数
	JpArgsNum int = 3

	// MaxOnePointNum 获取一个测点最多的数量
	//  参考依据：一天 1440 分钟
	MaxOnePointNum int = MinutesInDay

	// TenMinuteInterval 按10分钟取测点
	TenMinuteInterval int = 600

	// MinuteInterval 按分钟取测点
	MinuteInterval int = 60

	// SecondInterval 按秒取测点
	SecondInterval int = 5
)
