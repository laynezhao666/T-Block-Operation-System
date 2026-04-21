// Package consts 定义门禁服务的全局常量。
package consts

// ClientMysql MySQL客户端名称
// ClientRedis Redis客户端名称
// ClientKafka Kafka生产者客户端名称
const (
	ClientMysql = "trpc.mysql.tbos"
	ClientRedis = "trpc.redis.tbos"
	ClientKafka = "trpc.kafka.producer.dacPoint"
)

// ServicePort CGI服务端口
// ServiceIP CGI服务IP
// ServiceName 服务名称
// HTTPTimeout HTTP请求超时时间（毫秒）
var (
	ServicePort uint16 = 31234
	ServiceIP          = ""
	ServiceName        = "dac"
	HTTPTimeout        = 5000
)
