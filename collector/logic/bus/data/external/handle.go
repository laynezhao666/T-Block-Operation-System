// Package external 提供给外部上报数据等使用
package external

import (
	"context"
	"fmt"
	"time"

	"collector/entity/config"
	"collector/entity/model"
	"collector/logic/bus/data"
	"collector/repo/report"
	"collector/utils"

	"trpc.group/trpc-go/trpc-go/metrics"

	"etrpc-go/log"
	pb "trpcprotocol/collector"

	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	mainKafkaName   string = "mainExternalData"
	backupKafkaName string = "backupExternalData"
)

var (
	sendType string = ""
	trace    bool   = false
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
			Value: "external point",
		},
	}
)

// Init 初始化
func Init() {
	sendType = config.GetFeaturesConf().SendType
	trace = config.GetFeaturesConf().Trace
}

// DataHandle 处理数据上报请求
func DataHandle(ctx context.Context, req *pb.ExternalData) (*emptypb.Empty, error) {
	defer utils.HandlePanic("external_point")
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
	key := &model.MessageKey{
		WorkerID:  req.GetPlatformId(),
		Seq:       req.GetSeq(),
		Timestamp: req.GetCollectS(),
		PubMs:     req.GetPushMs(),
		Interval:  req.GetInterval(),
		MozuId:    fmt.Sprintf("%v", req.GetMozuId()),
	}
	points := req.GetPoints()
	value := &model.MessageValue{
		Interval: req.GetInterval(),
		Points:   make([]model.Point, 0, len(points)),
	}
	for _, p := range points {
		value.Points = append(value.Points, model.Point{
			Name:      p.GetI(),
			Value:     p.GetV(),
			Quality:   p.GetQ(),
			Timestamp: p.GetT(),
		})
	}
	log.Infof("handle external point, req: [%v], key: [%v], value: [%v]", req, key, value)
	// keyByte, err := jsoniter.Marshal(key)
	// if err != nil {
	// 	log.ErrorContextf(ctx, "marshal key [%v] fail: [%v]", key, err)
	// 	return nil, err
	// }
	// valueByte, err := jsoniter.Marshal(value)
	// if err != nil {
	// 	log.ErrorContextf(ctx, "marshal value [%v] fail: [%v]", value, err)
	// 	return nil, err
	// }

	// err = data.SendDataToKafka(keyByte, valueByte, mainKafkaName)
	// if err == nil {
	// 	log.Infof("send to MAIN kafka success, key: [%v]", key)
	// 	return &emptypb.Empty{}, nil
	// }
	// log.Warnf("key: [%v], send message to MAIN kafka fail:[ %+v], try backup kafka broker", key, err)
	// err = data.SendDataToKafka(keyByte, valueByte, backupKafkaName)
	// if err != nil {
	// 	log.Errorf("key: [%v], send message to BACKUP kafka fail: [%+v]", key, err)
	// 	return nil, errs.New(errcode.ErrSendFail, "send to kafka fail")
	// }
	// log.Infof("send to BACKUP kafka success, key: [%v]", key)
	return &emptypb.Empty{}, nil
}

func reportHandleMetric(handleCnt, handleFailCnt, latency float64, additionalDimensions []*metrics.Dimension) {
	dimensions := append(fixedDimensions, additionalDimensions...)
	report.HandleCnt(dimensions, handleCnt)
	report.HandleFailCnt(dimensions, handleFailCnt)
	report.HandleLatency(dimensions, latency)
}
