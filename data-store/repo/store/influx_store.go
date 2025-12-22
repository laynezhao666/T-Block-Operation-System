package store

import (
	"context"
	"data-store/entity/model"
	"data-store/repo/report"
	tinfluxdb "etrpc-go/client/influxdb"
	"etrpc-go/database/influxdb"
	"etrpc-go/log"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"
)

func init() {
	Register("influx", &InfluxStore{})
}

// InfluxStore InfluxDB存储插件
// usage: 持久化数据到Influxdb, 包含测点数据和测点变化数据
type InfluxStore struct {
	name            string          //名称
	StoreDay        int             `yaml:"store_day,omitempty"`        // 存储时长, 默认100天
	InfluxName      string          `yaml:"influx_name"`                // InfluxDB名称
	InfluxDatabase  string          `yaml:"influx_database,omitempty"`  // 数据库名称, 必填
	DataMeasurement string          `yaml:"data_measurement,omitempty"` // 数据测点测量名称, 默认points
	BathSize        int             `yaml:"bath_size,omitempty"`        // 批次写入大小,默认10000
	influxCli       influxdb.Client // influxdb客户端, influxdb客户端
}

// Setup 根据配置构建InfluxStore存储插件
func (f *InfluxStore) Setup(cfg PlgConfig) (IStorePlugin, error) {
	// 一些默认值
	store := &InfluxStore{
		name:            cfg.Name,
		StoreDay:        100,
		InfluxDatabase:  "tbos",
		DataMeasurement: "points",
		BathSize:        10000,
	}
	if err := cfg.Extra.Decode(store); err != nil {
		return nil, err
	}
	if store.StoreDay <= 0 || store.BathSize <= 0 {
		return nil, fmt.Errorf("cfg item:[store_day, bath_size] must > 0")
	}
	if store.InfluxName == "" {
		return nil, fmt.Errorf("cfg item:[influx_name] must not empty")
	}
	store.influxCli = tinfluxdb.GetClient(store.InfluxName)
	return store, nil
}

// Write 写入数据
func (f *InfluxStore) Write(originPoints []*model.OriginPointMsg) {
	points := make([]*model.Point, 0)
	for _, originPoint := range originPoints {
		points = append(points, originPoint.StdPoints...)
	}
	// 分批异步写入
	innerWg := &sync.WaitGroup{}
	ctx := trpc.BackgroundContext()
	for _, batch := range lo.Chunk(points, f.BathSize) {
		innerWg.Add(1)
		go func(bt []*model.Point) {
			defer innerWg.Done()
			begin := time.Now()
			if err := f.writeInfluxdb(ctx, bt); err != nil {
				report.CountByMozu(report.InfluxWriteFailCnt, bt, model.GetMozuId)
				log.AlarmContextf(ctx, "write points to influxdb fail, total:%d cost:%dms, err: %v",
					len(bt), time.Since(begin).Milliseconds(), err)
			} else {
				report.CountByMozu(report.InfluxWriteSuccessCnt, bt, model.GetMozuId)
			}
			end := time.Now().UnixMilli()
			cost := end - begin.UnixMilli()
			for _, point := range bt {
				point.InfluxTs = end
			}
			report.ValByMozu(report.InfluxWriteCost, bt, model.GetMozuId, float64(cost))
		}(batch)
	}
	innerWg.Wait()
}

func (f *InfluxStore) writeInfluxdb(ctx context.Context, bt []*model.Point) error {
	// 构建InfluxDB存储对象
	data := lo.Map(bt, func(item *model.Point, index int) *influx.Point {
		point, _ := influx.NewPoint(f.DataMeasurement, map[string]string{
			"i": item.Name, // 测点标识
		}, map[string]interface{}{
			"q":    item.Quality,  // 测点质量
			"v":    item.Value,    // 测点质量
			"d":    item.Interval, // 测点周期
			"type": item.Type,     // 测点类型
			"m":    item.MozuId,   // 模组ID
		}, time.Unix(item.Time, 0))
		return point
	})
	// 设置存储的db及时间精度范围
	batchPoints := influxdb.BatchPoints{
		Points:    data,
		Database:  f.InfluxDatabase,
		Precision: "s",
	}
	return retry.Do(func() error {
		return f.influxCli.Write(ctx, batchPoints)
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
}

// Close 关闭连接
func (f *InfluxStore) Close() error {
	return f.influxCli.Close(trpc.BackgroundContext())
}
