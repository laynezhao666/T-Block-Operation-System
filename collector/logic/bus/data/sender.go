package data

import (
	"fmt"

	"collector/repo/collector"
	"collector/repo/kafka"
	"collector/repo/report"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	HandleType string = "send point"
)

// SendDataToKafka 将构造好的消息转发到kafka，失败则发送到备用
func SendDataToKafka(key []byte, value []byte, kafkaName string) error {
	dimensions := []*metrics.Dimension{
		{
			Name:  report.TargetDimension,
			Value: kafkaName,
		},
	}
	report.SendDataCnt(dimensions, 1)
	err := kafka.SenderManager().GetSenderByName(kafkaName).Send(key, value)
	if err != nil {
		report.SendDataFailCnt(dimensions, 1)
		return fmt.Errorf("key: [%v], [%v] send message to KAFKA fail: [%v]", string(key), kafkaName, err)
	}
	return nil
}

// SendDataToCollectorSender 将构造好的消息转发到云端collector
func SendDataToCollectorSender(key []byte, value []byte, pointType string) ([]byte, error) {
	dimensions := []*metrics.Dimension{
		{
			Name:  report.TargetDimension,
			Value: pointType,
		},
	}
	report.SendDataCnt(dimensions, 1)
	err := collector.Sender().Send(key, value, pointType)
	if err != nil {
		report.SendDataFailCnt(dimensions, 1)
		return nil, fmt.Errorf("key: [%v] send message to COLLECTOR fail: [%+v]", string(key), err)
	}
	return fmt.Appendf([]byte{}, "send to COLLECTOR, key: [%v]", string(key)), nil
}
