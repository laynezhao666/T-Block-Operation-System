// Package utils 工具类代码
package utils

import (
	"fmt"
	"math"
	"strconv"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	model2 "agent/entity/model/data"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"agent/repo/monitor"
	"agent/utils/flog"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	pushPath          = "push_path"
	pushInterval      = "push_interval"
	pushPoints        = "push_points"
	pushVirtualPoints = "push_virtual_points"
)

var (
	filterLog *flog.Filter
)

// BoxInfo T-BOX相关信息
type BoxInfo struct {
	ClientId string `json:"client_id"`
}

// KafkaData 写入 kafka 的数据
type KafkaData struct {
	Interval      int                 `json:"interval"`
	Box           BoxInfo             `json:"box"`
	Points        []model2.KafkaPoint `json:"points"`
	VirtualPoints []model2.KafkaPoint `json:"virtual_points"`
}

// Copy 复制一份
func (k *KafkaData) Copy() *KafkaData {
	if k == nil {
		return nil
	}
	copy := &KafkaData{
		Interval:      k.Interval,
		Box:           k.Box,
		Points:        make([]model2.KafkaPoint, 0, len(k.Points)),
		VirtualPoints: make([]model2.KafkaPoint, 0, len(k.VirtualPoints)),
	}
	copy.Points = append(copy.Points, k.Points...)
	copy.VirtualPoints = append(copy.VirtualPoints, k.VirtualPoints...)
	return copy
}

// Log 日志
func (k *KafkaData) Log(dst string, deviceGid interface{}, interval int) {
	if k == nil {
		return
	}
	log.Infof("push device data to %+v, interval: %+v, id: %+v, points: %v, virtual points: %v",
		dst, interval, deviceGid, len(k.Points), len(k.VirtualPoints),
	)
}

// Report 上报
func (k *KafkaData) Report(path string, interval int) {
	if k == nil {
		return
	}
	monitor.ReportMetricsWithDimensions(
		pushPoints, float64(len(k.Points)), metrics.PolicySUM, []*metrics.Dimension{
			{Name: pushPath, Value: path},
			{Name: pushInterval, Value: fmt.Sprintf("%v", interval)},
		},
	)
	monitor.ReportMetricsWithDimensions(
		pushVirtualPoints, float64(len(k.VirtualPoints)), metrics.PolicySUM, []*metrics.Dimension{
			{Name: pushPath, Value: path},
			{Name: pushInterval, Value: fmt.Sprintf("%v", interval)},
		},
	)
}

func getVirtualPoints(deviceGiD definition.DeviceGidType) model.DataPoints {
	ids := definition.GetVirtualPointsID(deviceGiD)
	points := make(model.DataPoints, len(ids))
	for i, id := range ids {
		points[i].ID = id
	}
	rtdb.GetVirtualDataPoints(points)
	return points
}

// ConvertToKafkaPoints 转换为 kafka 点位
func ConvertToKafkaPoints(points model.DataPoints, forVendor bool) []model2.KafkaPoint {
	kPoints := make([]model2.KafkaPoint, 0, len(points))
	var err error
	var pv string
	for i, point := range points {
		if point.Rtd.Val.NotCollected() {
			filterLog.Debugf(point.ID, "not collected: %v", point.ID)
			continue
		}
		val := &points[i].Rtd.Val
		if pv, err = val.Pv.AsString(); err != nil {
			log.Errorf("ToKafkaData: AsString error: %v", err)
		}
		p := model2.KafkaPoint{
			ID:        string(points[i].ID),
			Value:     pv,
			Quality:   strconv.FormatInt(int64(val.Qua), 10),
			Timestamp: strconv.FormatInt(val.Tms, 10),
		}
		if forVendor {
			deviceGid, pointId, err := definition.SplitDataPointID(points[i].ID)
			if err == nil {
				deviceId, ok := cm.Worker().GetDeviceIdByGid(deviceGid)
				if ok {
					p.ID = string(definition.GenerateDataPointID(deviceId, pointId))
				}
			}
		}
		kPoints = append(kPoints, p)
	}
	return kPoints
}

// ToKafkaData 转换为 kafka 数据
func ToKafkaData(d *model2.DataUnit, interval int, forVendor bool) (*KafkaData, []*KafkaData, error) {
	if d == nil {
		return nil, nil, fmt.Errorf("ToKafkaData: DataUnit is nil")
	}

	isDefault := IsDefaultInterval(interval)
	realPoints, virtualPoints := rtdb.SegmentRealVirtual(d.Points)

	k := &KafkaData{
		Interval:      interval,
		Points:        ConvertToKafkaPoints(realPoints, forVendor),
		VirtualPoints: FilterInvalidVirtualPoints(ConvertToKafkaPoints(virtualPoints, forVendor), forVendor),
	}

	if isDefault {
		// 定时推送，同时推送方仓的虚拟测点
		k.VirtualPoints = append(k.VirtualPoints, ConvertToKafkaPoints(getVirtualPoints(d.DeviceGid), forVendor)...)
		// return k, []*KafkaData{k}, nil
	}

	emptyPoints := make([]model2.KafkaPoint, 0)
	if forVendor {
		k.Points = append(k.Points, k.VirtualPoints...)
		k.VirtualPoints = emptyPoints
	}

	l := int(math.Ceil(float64(len(k.Points)) * 1.0 / float64(definition.PointNumberPerMessage)))
	dataList := make([]*KafkaData, 0, l)
	begin := 0
	end := 0

	for i := 0; i < l; i++ {
		if end = begin + definition.PointNumberPerMessage; end > len(k.Points) {
			end = len(k.Points)
		}
		dataList = append(dataList, &KafkaData{
			Interval:      interval,
			Points:        k.Points[begin:end],
			VirtualPoints: emptyPoints,
		})
		begin += definition.PointNumberPerMessage
	}
	if len(k.VirtualPoints) > 0 {
		dataList = append(dataList, &KafkaData{
			Interval:      interval,
			Points:        emptyPoints,
			VirtualPoints: k.VirtualPoints,
		})
	}
	return k, dataList, nil
}
