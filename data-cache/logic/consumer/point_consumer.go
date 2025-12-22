package consumer

import (
	"context"
	"data-cache/entity/consts"
	"data-cache/entity/kafkamodel"
	"data-cache/entity/model"
	"data-cache/repo/cache"
	"data-cache/repo/report"
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
	source string
}

// NewPointKafkaConsumer 创建一个测点Kafka消费实例
func NewPointKafkaConsumer(source string) IKafkaConsumer {
	return &pointKafkaConsumer{
		source: source,
	}
}

// BatchHandle 批量消费测点数据
func (obj *pointKafkaConsumer) BatchHandle(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	dataPoints := make([]*model.StdPoint, 0)
	// 指标相关初始化
	nowMs := time.Now().UnixMilli()
	maxFutureTs := nowMs + time.Minute.Milliseconds()
	maxOldTs := nowMs - time.Hour.Milliseconds()
	for _, msg := range msgs {
		// decode消息
		kafkaKey := kafkamodel.PointMsgKey{}
		kafkaValue := kafkamodel.PointMsgValue{}
		if err := json.Unmarshal(msg.Key, &kafkaKey); err != nil {
			log.WarnContextf(ctx, "receive bad key msg from kafka [topic=%s, partition=%d,  msg-key=%s]",
				msg.Topic, msg.Partition, string(msg.Key))
			continue
		}
		if err := json.Unmarshal(msg.Value, &kafkaValue); err != nil {
			log.WarnContextf(ctx, "receive bad value msg from kafka [topic=%s, partition=%d,  value-key=%s]",
				msg.Topic, msg.Partition, string(msg.Value))
			continue
		}
		// 超过当前时间1分钟的点忽略掉,超过1小时前的点忽略掉
		if kafkaKey.PubMs > maxFutureTs || kafkaKey.PubMs < maxOldTs {
			// 采样异常的测点输出日志，方便定位问题
			var samplePoint *kafkamodel.Point
			if len(kafkaValue.Points) > 0 {
				samplePoint = &kafkaValue.Points[0]
			}
			if samplePoint == nil && len(kafkaValue.VirtualPoints) > 0 {
				samplePoint = &kafkaValue.VirtualPoints[0]
			}
			log.WarnContextf(ctx, "receive data out of range time [-1h,+1m], now time:[%s], data time:[%s]"+
				" sample point:[%v]", time.UnixMilli(nowMs).Format(time.DateTime),
				time.UnixMilli(kafkaKey.PubMs).Format(time.DateTime), samplePoint)
			continue
		}
		// decode出测点
		if len(kafkaValue.Points) > 0 {
			dataPoints = append(dataPoints, obj.resolvePoints(ctx, nowMs, &kafkaKey, kafkaValue.Points, false)...)
		}
		if len(kafkaValue.VirtualPoints) > 0 {
			dataPoints = append(dataPoints, obj.resolvePoints(ctx, nowMs, &kafkaKey, kafkaValue.VirtualPoints, true)...)
		}
	}
	// 向各个存储通道写入数据
	go cache.Write(dataPoints)
	return nil
}

func (obj *pointKafkaConsumer) resolvePoints(ctx context.Context, nowMs int64,
	msgKey *kafkamodel.PointMsgKey, points []kafkamodel.Point, virtual bool) []*model.StdPoint {
	// 解析测点类型
	var pointType int32
	// 兼容测点类型
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
	parsedPoints := make([]*model.StdPoint, 0, len(points))
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
		parsedPoints = append(parsedPoints, &model.StdPoint{
			Name:     point.I,
			Quality:  int16(q),
			Value:    val,
			Time:     uint32(msgKey.T),
			Type:     pointType,
			Interval: msgKey.D,
			MozuId:   mozuId,
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
	// 上报指标
	report.KafkaConsumeMsgCnt.ReportWithDim(1, dim)
	report.KafkaConsumePointCnt.ReportWithDim(float64(pointCnt), dim)
	report.KafkaConsumeDelay.ReportWithDim(float64(delay), dim)
}
