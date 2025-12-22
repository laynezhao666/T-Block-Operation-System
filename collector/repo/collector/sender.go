// Package collector 边端collector(NUC上部署)转发给云端collector相关
package collector

import (
	"fmt"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"

	collectorPb "trpcprotocol/collector"
)

const (
	stdPointServiceName     string = "idc-tbos-collector-std"
	collectPointServiceName string = "idc-tbos-collector-collect"
	CollectPointType        string = "collect_point"
	StdPointType            string = "std_point"
)

var (
	sender *CollectorSender
)

// CollectorSender 边端collector转发相关
type CollectorSender struct {
	dataBusProxy             collectorPb.DataBusClientProxy
	collectPointForwardProxy collectorPb.CollectPointForwardClientProxy
}

// Init 初始化
func Init() {
	sender = &CollectorSender{
		collectPointForwardProxy: collectorPb.NewCollectPointForwardClientProxy(
			client.WithServiceName(collectPointServiceName),
		),
		dataBusProxy: collectorPb.NewDataBusClientProxy(
			client.WithServiceName(stdPointServiceName),
		),
	}
}

// Sender 获取sender
func Sender() *CollectorSender {
	return sender
}

// Send 数据转发上报
func (s *CollectorSender) Send(key []byte, value []byte, pointType string) error {
	req := &collectorPb.ReqSend{
		Key:   key,
		Value: value,
	}
	switch pointType {
	case CollectPointType:
		_, err := s.collectPointForwardProxy.Forward(trpc.BackgroundContext(), req)
		if err != nil {
			return err
		}
	case StdPointType:
		_, err := s.dataBusProxy.Send(trpc.BackgroundContext(), req)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown point type [%v]", pointType)
	}
	return nil
}
