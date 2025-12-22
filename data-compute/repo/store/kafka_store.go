// Package store 数据写入到测点Kafka
package store

import (
	"common/entity/consts"
	"context"
	"data-compute/entity/kafkamodel"
	"data-compute/entity/model"
	"data-compute/repo/report"
	"encoding/json"
	"etrpc-go/log"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
)

var (
	kafkaStoreObj  IKafkaStore
	kafkaStoreOnce sync.Once
)

// IKafkaStore Kafka测点存储接口
type IKafkaStore interface {
	BatchWrite(allPoints []*model.Point, dataType int32)
}

// GetKafkaStore 获取Kafka存储对象
func GetKafkaStore() IKafkaStore {
	kafkaStoreOnce.Do(func() {
		obj := &kafkaStoreImpl{}
		obj.majorKafka = kafka.NewClientProxy(consts.TbosMajorKafkaName)
		backupCfg := client.Config(consts.TbosBackupKafkaName)
		if backupCfg.ServiceName != "" {
			obj.backupKafka = kafka.NewClientProxy(consts.TbosBackupKafkaName)
		}
		kafkaStoreObj = obj
	})
	return kafkaStoreObj
}

type kafkaStoreImpl struct {
	majorKafka  kafka.Client
	backupKafka kafka.Client
}

// BatchWrite 批量往Kafka写入数据
func (obj *kafkaStoreImpl) BatchWrite(allPoints []*model.Point, dataType int32) {
	if len(allPoints) == 0 {
		return
	}
	ctx := trpc.BackgroundContext()
	mozuPoints := lo.GroupBy(allPoints, func(item *model.Point) int32 {
		return item.MozuId
	})
	for mozuId, points := range mozuPoints {
		now := time.Now()
		// 生成Kafka的Key
		kafkaKey := kafkamodel.KafkaMsgKey{
			T:     now.Unix(),
			D:     dataType,
			Type:  consts.PointTypeStd,
			MID:   fmt.Sprint(mozuId),
			PubMs: now.UnixMilli(),
		}
		keyBytes, _ := json.Marshal(kafkaKey)

		dim := map[string]string{
			report.PointIntervalDimKey: fmt.Sprint(dataType),
			report.PointMozuIdDimKey:   fmt.Sprint(mozuId),
		}
		chunkPoints := lo.Chunk(points, 2000)
		log.Debugf("begin write points to kafka, total: %d, type:%d, mozuId:%d", len(points), dataType, mozuId)
		for _, chunk := range chunkPoints {
			begin := time.Now().UnixMilli()
			// 生成Kafka的测点Value
			kafkaPoints := lo.Map(chunk, func(item *model.Point, index int) *kafkamodel.KafkaMsgPoint {
				return &kafkamodel.KafkaMsgPoint{
					I: item.Name,
					Q: fmt.Sprintf("%d", item.Quality),
					V: fmt.Sprintf("%f", item.Value),
					T: fmt.Sprintf("%d", item.Time),
				}
			})
			kafkaValue := kafkamodel.KafkaMsgValue{
				Points: kafkaPoints,
			}
			valueBytes, _ := json.Marshal(kafkaValue)
			// 向Kafka推送数据
			source, success, err := obj.pushPoints(ctx, keyBytes, valueBytes)
			end := time.Now().UnixMilli()
			if err != nil {
				log.AlarmContextf(ctx, "write point to major kafka fail, err: %v", err)
			}
			// 保存推送的信息
			for _, point := range chunk {
				point.SendTms = end // 记录推送时间
				if !success {
					point.Quality = int32(consts.QualityPushKafkaErr) // 修改测点质量为推送失败
				}
			}
			// 上报部分指标数据
			dim[report.PointSourceDimKey] = source
			if success {
				report.PointPushSuccessCnt.ReportWithDim(float64(len(chunk)), dim)
			} else {
				report.PointPushFailCnt.ReportWithDim(float64(len(chunk)), dim)
			}
			report.PointPushCost.ReportWithDim(float64(end-begin), dim)
		}
	}
}

func (obj *kafkaStoreImpl) pushPoints(ctx context.Context, key, value []byte) (string, bool, error) {
	// 主用Kafka推送
	source := consts.TbosMajorKafkaName
	success := true
	err := retry.Do(func() error {
		return obj.majorKafka.Produce(ctx, key, value)
	}, retry.Attempts(3),
		retry.RetryIf(func(err error) bool { return err != nil }))
	// 备用Kafka推送
	if err != nil {
		success = false
		if obj.backupKafka != nil {
			backupErr := retry.Do(func() error {
				return obj.backupKafka.Produce(ctx, key, value)
			}, retry.Attempts(3),
				retry.RetryIf(func(err error) bool { return err != nil }))
			// 推送成功，算备用Kafka成功,推送失败，算主用Kafka失败
			if backupErr == nil {
				success = true
				source = consts.TbosBackupKafkaName
			}
		}
	}
	return source, success, err
}
