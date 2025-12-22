package repo

import (
	"collector/entity/config"
	"collector/repo/collector"
	"fmt"

	"collector/repo/kafka"
)

func Init() error {
	sendType := config.GetFeaturesConf().SendType
	switch sendType {
	case config.CollectorSendType:
		collector.Init()
	case config.KafkaSendType:
		kafka.Init()
	default:
		return fmt.Errorf("unknown send type: %v", sendType)
	}
	return nil
}
