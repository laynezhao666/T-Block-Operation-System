// Package utils 工具类代码
package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"

	"agent/entity/consts"
	"agent/entity/definition"
	model2 "agent/entity/model/data"
	"agent/logic/cm"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
	"agent/repo/monitor"
	"agent/utils/flog"
)

const (
	pushPath          = "push_path"
	pushInterval      = "push_interval"
	pushPoints        = "push_points"
	pushVirtualPoints = "push_virtual_points"
	pushSendOk        = "push_send_ok"   // 推送成功测点数
	pushSendFail      = "push_send_fail" // 推送失败测点数
	pushCommOk        = "push_comm_ok"   // 通讯状态正常设备数
	pushCommErr       = "push_comm_err"  // 通讯状态异常设备数
	pushQuaBad        = "push_qua_bad"   // QUA 不为 0 的测点数
	pushMozuID        = "mozu_id"        // 模组ID维度
	pushType          = "type"           // 数据类型维度（collect/std）

	commSuffix      = consts.DefaultIDSep + definition.CommID // ".Comm"
	commInterruptPv = "1"                                     // pv=1 表示通讯中断
	qualityOkStr    = "0"                                     // QUA=0 表示质量正常
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
	Type          int                 `json:"type,omitempty"` // 1=采集点, 2=标准点
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
		Type:          k.Type,
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
	log.Debugf("push device data to %+v, interval: %+v, id: %+v, points: %v, virtual points: %v",
		dst, interval, deviceGid, len(k.Points), len(k.VirtualPoints),
	)
}

// dataTypeStr 将数据类型整数转换为字符串维度值
func dataTypeStr(dataType int) string {
	switch dataType {
	case definition.KafkaDataTypeCollector:
		return "collect"
	case definition.KafkaDataTypeStd:
		return "std"
	default:
		return fmt.Sprintf("%d", dataType)
	}
}

// Report 上报（不含 sendOk/sendFail 统计，兼容无法获取发送结果的调用方）
func (k *KafkaData) Report(path string, interval int, mozuID string, dataType int) {
	k.ReportWithStats(path, interval, mozuID, 0, 0, dataType)
}

// commStats 通讯状态统计结果
type commStats struct {
	commOk  int // 通讯正常设备数
	commErr int // 通讯异常设备数
	quaBad  int // QUA 不为 0 的测点数
}

// collectCommStats 从 KafkaData 中统计通讯状态和 QUA 异常数
// Comm 测点 pv="1" 表示通讯中断，pv="0" 表示通讯正常
// 注意：Comm 测点是虚拟测点，在 VirtualPoints 中，而非 Points 中
func collectCommStats(k *KafkaData) commStats {
	var s commStats
	// 从虚拟测点中统计通讯状态（Comm 测点在 GetVirtualPointsID 中）
	for _, p := range k.VirtualPoints {
		if strings.HasSuffix(p.ID, commSuffix) {
			if p.Value == commInterruptPv {
				s.commErr++
			} else {
				s.commOk++
			}
		}
	}
	// 从采集测点中统计 QUA 异常数
	for _, p := range k.Points {
		if p.Quality != qualityOkStr {
			s.quaBad++
		}
	}
	return s
}

// ReportWithStats 上报（含推送成功/失败数、通讯状态和 QUA 异常数）
// 所有指标统一上报，通过 type 维度区分数据类型
func (k *KafkaData) ReportWithStats(path string, interval int, mozuID string, sendOk, sendFail int, dataType int) {
	if k == nil {
		return
	}
	dimensions := []*metrics.Dimension{
		{Name: pushPath, Value: path},
		{Name: pushInterval, Value: fmt.Sprintf("%v", interval)},
		{Name: pushMozuID, Value: mozuID},
		{Name: pushType, Value: dataTypeStr(dataType)},
	}
	cs := collectCommStats(k)
	m := []*metrics.Metrics{
		metrics.NewMetrics(pushPoints, float64(len(k.Points)), metrics.PolicySUM),
		metrics.NewMetrics(pushVirtualPoints, float64(len(k.VirtualPoints)), metrics.PolicySUM),
		metrics.NewMetrics(pushSendOk, float64(sendOk), metrics.PolicySUM),
		metrics.NewMetrics(pushSendFail, float64(sendFail), metrics.PolicySUM),
		metrics.NewMetrics(pushCommOk, float64(cs.commOk), metrics.PolicySUM),
		metrics.NewMetrics(pushCommErr, float64(cs.commErr), metrics.PolicySUM),
		metrics.NewMetrics(pushQuaBad, float64(cs.quaBad), metrics.PolicySUM),
	}
	monitor.ReportMultiMetricsWithDimensions(m, dimensions)
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
		VirtualPoints: ConvertToKafkaPoints(virtualPoints, forVendor),
	}

	if isDefault {
		// 定时推送，同时推送 mozu 下所有设备的虚拟测点
		if len(d.DeviceGids) > 0 {
			for _, gid := range d.DeviceGids {
				k.VirtualPoints = append(k.VirtualPoints,
					ConvertToKafkaPoints(getVirtualPoints(gid), forVendor)...)
			}
		} else {
			// 兼容未设置 DeviceGids 的旧路径（如变化上报）
			k.VirtualPoints = append(k.VirtualPoints,
				ConvertToKafkaPoints(getVirtualPoints(d.DeviceGid), forVendor)...)
		}
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
