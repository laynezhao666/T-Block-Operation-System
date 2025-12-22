// Package tbos_point 处理测点数据，上报给tbos
package tbos_point

import (
	"context"
	"time"

	"collector/entity/config"
	"collector/logic/bus/data"
	"collector/repo/collector"
	"collector/repo/report"
	"collector/utils"

	"trpc.group/trpc-go/trpc-go/metrics"

	"etrpc-go/log"
	pb "trpcprotocol/collector"
)

const (
	mainKafkaName   string = "mainStdPoint"
	backupKafkaName string = "backupStdPoint"
)

var (
	sendType string = ""
)

var (
	fixedDimensions = []*metrics.Dimension{
		{
			Name:  report.HandleTypeDimension,
			Value: data.HandleType,
		},
		{
			Name:  report.SendTypeDimension,
			Value: sendType,
		},
		{
			Name:  report.PointTypeDimension,
			Value: "tbos point",
		},
	}
)

// Init 初始化
func Init() {
	sendType = config.GetFeaturesConf().SendType
}

// SendHandle 处理数据上报请求
func SendHandle(ctx context.Context, req *pb.ReqSend) {
	defer utils.HandlePanic("tbos_point")
	startTime := time.Now().UnixMilli()
	handleCnt := 1
	handleFailCnt := 0
	defer func() {
		endTime := time.Now().UnixMilli()
		t := float64(endTime - startTime)
		additionalDimensions := []*metrics.Dimension{
			{
				Name:  report.UpstreamIpDimension,
				Value: utils.GetUpstreamIp(ctx),
			},
		}
		reportHandleMetric(float64(handleCnt), float64(handleFailCnt), t, additionalDimensions)
	}()
	key := req.GetKey()
	value := req.GetValue()
	switch sendType {
	case config.CollectorSendType:
		// 配置collector为true，则转发云端kafka，一般nuc上的collector会做这项配置
		_, err := data.SendDataToCollectorSender(key, value, collector.StdPointType)
		if err != nil {
			handleFailCnt = 1
			log.ErrorContextf(ctx, "send to COLLECTOR fail: [%v]", err)
			return
		}
		log.InfoContextf(ctx, "send to COLLECTOR success, key: [%v]", string(key))
	case config.KafkaSendType:
		fallthrough
	default:
		errors := make([]error, 0)
		err := data.SendDataToKafka(key, value, mainKafkaName)
		if err == nil {
			log.InfoContextf(ctx, "send to MAIN kafka success, key: [%v]", string(key))
			return
		}
		// 失败则走备用通道
		log.WarnContextf(ctx, "key: [%v], send message to MAIN kafka fail:[ %+v], try backup kafka broker", string(key), err)
		errors = append(errors, err)
		err = data.SendDataToKafka(key, value, backupKafkaName)
		if err != nil {
			handleFailCnt = 1
			errors = append(errors, err)
			log.ErrorContextf(ctx, "key: [%v], send message to MAIN and BACKUP kafka fail: [%+v]", string(key), errors)
			return
		}
		log.InfoContextf(ctx, "send to BACKUP kafka success, key: [%v]", string(key))
	}
}

func reportHandleMetric(handleCnt, handleFailCnt, latency float64, additionalDimensions []*metrics.Dimension) {
	dimensions := append(fixedDimensions, additionalDimensions...)
	report.HandleCnt(dimensions, handleCnt)
	report.HandleFailCnt(dimensions, handleFailCnt)
	report.HandleLatency(dimensions, latency)
}
