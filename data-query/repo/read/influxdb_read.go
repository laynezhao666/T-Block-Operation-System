package read

import (
	"context"
	"data-query/entity"
	"encoding/json"
	tinfluxdb "etrpc-go/client/influxdb"
	"etrpc-go/database/influxdb"
	"etrpc-go/log"
	"fmt"
	"strings"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func init() {
	Register("influx", &InfluxRead{})
}

// InfluxRead InfluxDB存储插件
// usage: 持久化数据到Influxdb, 包含测点数据和测点变化数据
type InfluxRead struct {
	name            string          //名称
	StoreDay        int             `yaml:"store_day,omitempty"`        // 存储时长, 默认100天
	StoreDuration   time.Duration   `yaml:"-"`                          // 存储时长
	InfluxName      string          `yaml:"influx_name"`                // InfluxDB名称
	InfluxDatabase  string          `yaml:"influx_database,omitempty"`  // 数据库名称, 必填
	DataMeasurement string          `yaml:"data_measurement,omitempty"` // 数据测点测量名称, 默认points
	BathSize        int             `yaml:"bath_size,omitempty"`        // 批次写入大小,默认10000
	influxCli       influxdb.Client // influxdb客户端, influxdb客户端
}

// GetType 判断插件类型
func (f *InfluxRead) GetType() string {
	return entity.ReadInfluxType
}

// Setup 构建influxdb查询插件
// Setup 根据配置构建InfluxStore存储插件
func (f *InfluxRead) Setup(cfg PlgConfig) (IReadPlugin, error) {
	// 一些默认值
	plugin := &InfluxRead{
		name:            cfg.Name,
		StoreDay:        100,
		InfluxDatabase:  "tbos",
		DataMeasurement: "points",
		BathSize:        10000,
	}
	if err := cfg.Extra.Decode(plugin); err != nil {
		return nil, err
	}
	if plugin.StoreDay <= 0 || plugin.BathSize <= 0 {
		return nil, fmt.Errorf("cfg item:[store_day, bath_size] must > 0")
	}
	if plugin.InfluxName == "" {
		return nil, fmt.Errorf("cfg item:[influx_name] must not empty")
	}
	plugin.influxCli = tinfluxdb.GetClient(plugin.InfluxName)
	plugin.StoreDuration = time.Duration(plugin.StoreDay) * time.Hour * 24
	return plugin, nil
}

// CanRead 判断当前时间范围能否查询
func (f *InfluxRead) CanRead(begin, end int64) bool {
	minTs := time.Now().Add(-f.StoreDuration).Unix()
	return begin >= minTs && end >= minTs
}

// ReadRange 查询时间范围数据
func (f *InfluxRead) ReadRange(ctx context.Context, pointNames []string, begin, end int64) (map[string][]*entity.Point, error) {
	if len(pointNames) == 0 {
		return make(map[string][]*entity.Point), nil
	}

	conditions := lo.Map(pointNames, func(pointName string, index int) string {
		return fmt.Sprintf("i='%s'", pointName)
	})
	// 为了后续数据补齐，往前多查一个点
	//_, err := f.ReadLatest(ctx, pointNames, begin-1)
	//if err != nil {
	//	return nil, err
	//}
	// 为了补齐begin点的数据, 多取一分钟的数据，避免两次请求influxdb
	query, err := f.executeQuery(ctx, begin-65, end, conditions)
	if err != nil {
		return nil, err
	}

	points := f.processQueryResults(query)

	res := lo.GroupBy(points, func(point *entity.Point) string {
		return point.Name
	})
	return res, nil
}

func (f *InfluxRead) executeQuery(ctx context.Context, begin, end int64, conditions []string) (*influx.Response, error) {
	rsp, err := f.influxCli.Query(ctx, influxdb.Query{
		Command: fmt.Sprintf("SELECT time, i, v FROM %s WHERE time >= %d and time <= %d and q = 0 and (%s)",
			f.DataMeasurement, begin*1000000000, end*1000000000, strings.Join(conditions, " OR ")),
		Database:  f.InfluxDatabase,
		Precision: "s",
	})

	if err != nil || rsp.Err != "" {
		errMsg := "request influxdb fail"
		if err == nil {
			errMsg = fmt.Sprintf("query influxdb fail, query err: %s", rsp.Err)
		}
		return nil, errors.Wrapf(err, errMsg)
	}

	return rsp, nil
}

func (f *InfluxRead) processQueryResults(rsp *influx.Response) []*entity.Point {
	points := make([]*entity.Point, 0)

	for _, queryRes := range rsp.Results {
		if queryRes.Err != "" {
			log.Errorf("query influxdb data err:%s", queryRes.Err)
			continue
		}

		for _, row := range queryRes.Series {
			for _, v := range row.Values {
				if len(v) != 3 || v[0] == nil || v[1] == nil || v[2] == nil {
					log.Errorf("query influxdb data err, data:%v", v)
					continue
				}

				ts, err1 := v[0].(json.Number).Int64()
				val, err2 := v[2].(json.Number).Float64()
				//qlt, err2 := v[2].(json.Number).Int64()

				if err1 != nil || err2 != nil {
					log.Errorf("decode influxdb data err, data:%v", v)
					continue
				}

				point := &entity.Point{
					Name: v[1].(string),
					Time: ts,
					//Quality: qlt,
					Value: val,
				}
				points = append(points, point)
			}
		}
	}

	return points
}

// ReadLatest 查询某个时间点前最新数据
func (f *InfluxRead) ReadLatest(ctx context.Context, pointNames []string, max int64) (map[string]*entity.Point, error) {
	res := make(map[string]*entity.Point)
	if len(pointNames) == 0 {
		return res, nil
	}
	conditions := lo.Map(pointNames, func(pointName string, index int) string {
		return fmt.Sprintf("i='%s'", pointName)
	})
	query, err := f.influxCli.Query(ctx, influxdb.Query{
		Command: fmt.Sprintf(
			"SELECT time, q, v FROM %s WHERE time <= %d and (%s) and q = 0 group by i order by time desc limit 1",
			f.DataMeasurement, max*1000000000, strings.Join(conditions, " OR ")),
		Database:  f.InfluxDatabase,
		Precision: "s",
	})
	if err != nil || query.Err != "" {
		errorMsg := "request points influxdb fail"
		if err == nil {
			errorMsg = fmt.Sprintf("query points influxdb fail, query err: %s", query.Err)
		}
		return nil, errors.Wrapf(err, errorMsg)
	}
	for _, queryRes := range query.Results {
		if queryRes.Err != "" {
			continue
		}
		for _, row := range queryRes.Series {
			v := row.Values[0]
			if len(v) < 3 {
				continue
			}
			ts, _ := v[0].(json.Number).Int64()
			qlt, _ := v[1].(json.Number).Int64()
			val, _ := v[2].(json.Number).Float64()
			point := &entity.Point{
				Name:    row.Tags["i"],
				Time:    ts,
				Quality: qlt,
				Value:   val,
			}
			res[point.Name] = point
		}
	}
	return res, nil
}

// ReadChanged 查询测点最近变化的时间
func (f *InfluxRead) ReadChanged(ctx context.Context, pointNames []string, begin int64, end int64) (map[string]int64, error) {
	return nil, nil
}
