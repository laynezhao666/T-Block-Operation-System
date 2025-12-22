package store

import (
	"data-store/entity/model"
	"data-store/repo/report"
	"errors"
	"etrpc-go/log"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"
)

func init() {
	Register("kafka", &KafkaStore{})
}

// KafkaStore Kafka数据转存插件
// usage: 将数据转存到kafka，供下游使用
type KafkaStore struct {
	name      string       //kafka插件名称
	KafkaName string       `yaml:"kafka_name"` // kafka对应从client名称
	kafkaCli  kafka.Client // kafka客户端
}

// Setup 构建Kafka存储插件
func (k *KafkaStore) Setup(cfg PlgConfig) (IStorePlugin, error) {
	store := &KafkaStore{
		name: cfg.Name,
	}
	if err := cfg.Extra.Decode(store); err != nil {
		return nil, err
	}
	if store.KafkaName == "" {
		return nil, errors.New("cfg item:[kafka_name] is require")
	}
	store.kafkaCli = kafka.NewClientProxy(store.KafkaName)
	return store, nil
}

// Write 写入数据
func (k *KafkaStore) Write(points []*model.OriginPointMsg) {
	// kafka写入数据很快，直接顺序写入
	ctx := trpc.BackgroundContext()
	for _, point := range points {
		// 这个加个重试，避免网络抖动
		begin := time.Now()
		dim := map[string]string{report.MozuIdDimKey: fmt.Sprint(point.StdPoints[0].MozuId)}
		err := retry.Do(func() error {
			return k.kafkaCli.Produce(ctx, point.KafkaKey, point.KafkaVal)
		}, retry.Attempts(3), retry.RetryIf(func(err error) bool { return err != nil }))
		cost := time.Since(begin).Milliseconds()
		report.KafkaWriteCost.ReportWithDim(float64(cost), dim)
		if err != nil {
			report.KafkaWriteFailCnt.ReportWithDim(float64(len(point.StdPoints)), dim)
			log.AlarmContextf(ctx, "store point to kafka[%s] fail, err: %v", k.name, err)
		} else {
			report.KafkaWriteSuccessCnt.ReportWithDim(float64(len(point.StdPoints)), dim)
		}
	}
}

// Close 关闭连接
func (k *KafkaStore) Close() error {
	return nil
}
