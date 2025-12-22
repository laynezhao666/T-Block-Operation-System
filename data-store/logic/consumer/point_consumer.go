// Package consumer 核心业务逻辑层，消费测点kafka
package consumer

import (
	"common/entity/consts"
	"context"
	"data-store/entity/kafkamodel"
	"data-store/entity/model"
	"data-store/repo/report"
	"data-store/repo/store"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"trpc.group/trpc-go/trpc-go/log"
)

// IKafkaConsumer kafka测点数据消费接口
type IKafkaConsumer interface {
	// BatchHandle 批量消费测点数据
	BatchHandle(ctx context.Context, msgs []*sarama.ConsumerMessage) error
}

type pointKafkaConsumer struct {
	source string // kafka标识
}

// NewPointKafkaConsumer 创建一个测点Kafka消费实例
func NewPointKafkaConsumer(source string) IKafkaConsumer {
	return &pointKafkaConsumer{
		source: source,
	}
}

// BatchHandle 批量消费测点数据
func (obj *pointKafkaConsumer) BatchHandle(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	originPoints := make([]*model.OriginPointMsg, 0, len(msgs))
	// 指标相关初始化
	nowMs := time.Now().UnixMilli()
	maxFutureTs := nowMs + time.Minute.Milliseconds()
	for _, msg := range msgs {
		// decode消息
		kafkaKey := kafkamodel.PointMsgKey{}
		kafkaValue := kafkamodel.PointMsgValue{}
		if err := json.Unmarshal(msg.Key, &kafkaKey); err != nil {
			log.WarnContextf(ctx, "receive bad key msg from kafka [topic=%s, partition=%d, offset=%d, msg-key=%s]",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Key))
			continue
		}
		if err := json.Unmarshal(msg.Value, &kafkaValue); err != nil {
			log.WarnContextf(ctx, "receive bad value msg from kafka [topic=%s, partition=%d, offset=%d, value-key=%s]",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
			continue
		}
		// 超过当前时间1分钟的点忽略掉
		if kafkaKey.PubMs > maxFutureTs {
			log.WarnContextf(ctx, "receive data of future time from kafka, now time:[%d], data time:[%d], exceed:[%d]ms",
				nowMs, kafkaKey.PubMs, kafkaKey.PubMs-nowMs)
			continue
		}
		// decode出测点
		parsedPoints := make([]*model.Point, 0, len(kafkaValue.Points)+len(kafkaValue.VirtualPoints))
		if len(kafkaValue.Points) > 0 {
			parsedPoints = append(parsedPoints, obj.resolvePoints(ctx, nowMs, &kafkaKey, kafkaValue.Points, false)...)
		}
		if len(kafkaValue.VirtualPoints) > 0 {
			parsedPoints = append(parsedPoints, obj.resolvePoints(ctx, nowMs, &kafkaKey, kafkaValue.VirtualPoints, true)...)
		}
		// 组装写入消息
		originPoints = append(originPoints, &model.OriginPointMsg{
			KafkaKey:  msg.Key,
			KafkaVal:  msg.Value,
			StdPoints: parsedPoints,
		})
	}
	// 向各个存储通道写入数据
	store.BatchWritePoint(originPoints)
	return nil
}

func (obj *pointKafkaConsumer) resolvePoints(ctx context.Context, nowMs int64,
	msgKey *kafkamodel.PointMsgKey, points []kafkamodel.Point, virtual bool) []*model.Point {
	// 解析测点类型
	var pointType int32
	if msgKey.Type > 0 {
		pointType = msgKey.Type
	} else {
		switch msgKey.DID {
		case consts.PointTypeStdStr:
			pointType = consts.PointTypeStd
		case consts.PointTypeAlarmStr:
			pointType = consts.PointTypeAlarm
		default:
			pointType = consts.PointTypeCollect
		}
		if virtual {
			pointType = consts.PointTypeVirtual
		}
	}

	// 解析模组ID
	var mozuId int32
	if msgKey.MID != "" {
		if val, err := strconv.ParseInt(msgKey.MID, 10, 32); err == nil {
			mozuId = int32(val)
		} else {
			log.WarnContextf(ctx, "receive bad point, parse mozuId failed, val:%s", msgKey.MID)
		}
	}

	// 解析所有测点
	parsedPoints := make([]*model.Point, 0, len(points))
	for _, point := range points {
		val, err := strconv.ParseFloat(point.V, 64)
		if err != nil {
			val = 0
		}
		q, err := strconv.ParseInt(point.Q, 10, 32)
		if err != nil {
			log.WarnContextf(ctx, "receive bad point, parse quality failed, val:%s", point.Q)
			continue
		}
		parsedPoints = append(parsedPoints, &model.Point{
			Name:       point.I,
			Quality:    int32(q),
			Value:      val,
			Time:       msgKey.T,
			Type:       pointType,
			Interval:   msgKey.D,
			MozuId:     mozuId,
			CollectTs:  msgKey.PubMs,
			ConsumerTs: nowMs,
		})
	}

	// 上报指标数据
	obj.reportMetric(pointType, msgKey.D, mozuId, len(parsedPoints), nowMs-msgKey.PubMs)

	return parsedPoints
}

func (obj *pointKafkaConsumer) reportMetric(pointType, interval, mozuId int32, pointCnt int, delay int64) {
	dim := map[string]string{
		report.KafkaSourceDimKey:   obj.source,
		report.PointTypeDimKey:     fmt.Sprint(pointType),
		report.PointIntervalDimKey: fmt.Sprint(interval),
		report.MozuIdDimKey:        fmt.Sprint(mozuId),
	}
	report.KafkaConsumeMsgCnt.ReportWithDim(1, dim)
	report.KafkaConsumePointCnt.ReportWithDim(float64(pointCnt), dim)
	report.KafkaConsumeDelay.ReportWithDim(float64(delay), dim)
}
