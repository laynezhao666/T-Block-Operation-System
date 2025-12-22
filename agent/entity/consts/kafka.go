package consts

const (
	DefaultKafkaMaxAttempt     = 3
	DefaultKafkaWriteTimeoutMs = 3000

	DefaultNorthReportSecond = 0 // 北向分钟级上报周期的秒数，比如这里值为0，标识每分钟第0秒上报

	EnableMozuTopic = "mozu_topic"
)
