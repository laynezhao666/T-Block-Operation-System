// Package kafka kafka 相关
package kafka

import (
	"agent/entity/config"
	"agent/entity/definition"
	"agent/utils"
)

var (
	kafkaWriter    *definition.KafkaWriterType
	forwardsWriter []*definition.KafkaWriterType
)

// GetWriter 获取写kafka的writer
func GetWriter() *definition.KafkaWriterType {
	return kafkaWriter
}

// GetForwardsWriter 获取写kafka的writer
func GetForwardsWriter() []*definition.KafkaWriterType {
	return forwardsWriter
}

// Init 初始化kafka
func Init() {
	kafkaWriter = utils.NewKafkaWriter(
		config.GetRB().Distributor.Kafka.Brokers, nil,
		config.GetRB().Distributor.Kafka.Topic.Points,
		config.GetRB().Distributor.Kafka.SASL.Mechanism,
		config.GetRB().Distributor.Kafka.SASL.Username,
		config.GetRB().Distributor.Kafka.SASL.Password,
	)

	forwards := config.GetRB().Distributor.Forwards
	forwardsWriter = make([]*definition.KafkaWriterType, len(forwards))
	for i := range forwards {
		forwardsWriter[i] = utils.NewKafkaWriter(forwards[i].Brokers, nil, forwards[i].Topic.Points,
			forwards[i].SASL.Mechanism, forwards[i].SASL.Username, forwards[i].SASL.Password)
	}
}

// 暂不需要自定义分区策略
//type Balancer struct {
//	offset uint32
//}
//
//func (b *Balancer) Balance(msg kafka.Message, partitions ...int) (partition int) {
//	var (
//		k      KafkaKey
//		offset uint32
//	)
//	l := uint32(len(partitions))
//	if err := json.Unmarshal(msg.Key, &k); err != nil || len(k.BalancerKey) == 0 {
//		// 若反序列化失败或模组 Gid 为空，则退化为 RoundRobin
//		offset = atomic.AddUint32(&b.offset, 1) - 1
//	} else {
//		// 否则计算设备 Gid 的 BKDR hash
//		offset = bkdr([]byte(k.BalancerKey))
//	}
//	return partitions[offset%l]
//}
//
//func bkdr(b []byte) uint32 {
//	var (
//		seed uint32 = 131 // 31 131 1313 13131 131313 etc..
//		hash uint32 = 0
//	)
//	l := len(b)
//	for i := 0; i < l; i++ {
//		hash = hash*seed + uint32(b[i])
//	}
//	return hash
//}
