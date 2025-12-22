package vtpoint

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/conf"
	"alarm-compute/repo"
)

const (
	DefaultPointBatchSize = 100
)

var (
	pointCollector PointCollector
	pointOnce      sync.Once
)

// VirtualPoint 虚拟测点
type VirtualPoint struct {
	I string `json:"i"` // 测点名称
	V string `json:"v"` // 测点值
	Q string `json:"q"` // 质量
	T string `json:"t"` // 时间戳
}

// KafkaMessageKey 虚拟测点消息key
type KafkaMessageKey struct {
	MId   string `json:"mID"`
	DId   string `json:"dID"`
	WId   string `json:"wID"`
	Seq   int64  `json:"seq"`
	T     int64  `json:"t"`
	D     int64  `json:"d"`
	BKey  string `json:"bKey"`
	PubMs int64  `json:"pubMs"`
}

// VirtualPointMsg 虚拟测点消息
type VirtualPointMsg struct {
	Interval      int64          `json:"interval"`
	BoxID         string         `json:"box_id"`
	Points        []VirtualPoint `json:"points"` // 测点数据组
	VirtualPoints []VirtualPoint `json:"virtual_points"`
}

// PointCollector 任务生效收集器
type PointCollector struct {
	//有效性任务 chan 实时上报
	Collector chan *VirtualPoint
}

// GetPointCollector GetPointCollector
func GetPointCollector() *PointCollector {
	pointOnce.Do(func() {
		batchSize := conf.ServerConf.VirtualConfig.PointKafkaBatchSize
		if batchSize < 1000 {
			batchSize = DefaultPointBatchSize
		}
		pointCollector.Collector = make(chan *VirtualPoint, 10*batchSize)
	})
	return &pointCollector
}

// AddVirtualPointData AddVirtualPointData
func (pc *PointCollector) AddVirtualPointData(item *VirtualPoint) {
	pc.Collector <- item
}

func (pc *PointCollector) batchSendVtPointData(dataList []VirtualPoint) {
	if len(dataList) == 0 {
		return
	}
	timeUnix, err := strconv.ParseInt(dataList[0].T, 10, 64)
	if err != nil {
		log.Errorf("parse virtual point msg time failed, err: %v", err)
		return
	}
	newMsgKey := &KafkaMessageKey{
		DId:   "alarmVirtualPoints",
		T:     timeUnix,
		D:     1,
		PubMs: time.Now().UnixMilli(),
	}
	newPointMsg := &VirtualPointMsg{
		Interval: 1,
		BoxID:    "alarmVirtualPoints",
		Points:   dataList,
	}
	key, err := json.Marshal(newMsgKey)
	if err != nil {
		log.Errorf("marshal virtual point key msg failed, err: %v", err)
		return
	}
	data, err := json.Marshal(newPointMsg)
	if err != nil {
		log.Errorf("marshal virtual point data msg failed, err: %v", err)
		return
	}
	err = repo.GetCkafka().SendPointMsg([]byte(key), data)
	if err != nil {
		log.AlarmContextf(trpc.BackgroundContext(), "发送虚拟测点消息失败:%v", dataList)
	} else {
		go repo.GetCkafka().SendAdminMsg(data, timeUnix, time.Now())
	}
}

// ReportVtPointData ReportVtPointData
func (pc *PointCollector) ReportVtPointData(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	interval := conf.ServerConf.VirtualConfig.FlushInterval
	batchSize := conf.ServerConf.VirtualConfig.PointKafkaBatchSize
	if interval == 0 {
		interval = 100
		batchSize = 200
	}
	flushInterval := time.Duration(interval) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	var curTime string
	var results []VirtualPoint
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-pc.Collector:
			if item.T != curTime {
				if len(results) > 0 {
					pc.batchSendVtPointData(results)
					results = make([]VirtualPoint, 0)
				}
				results = append(results, *item)
			} else {
				results = append(results, *item)
				if len(results) >= int(batchSize) {
					//TODO: 发送kafka消息
					pc.batchSendVtPointData(results)
					results = make([]VirtualPoint, 0)
				}
			}
			curTime = item.T
		case <-ticker.C:
			if len(results) > 0 {
				pc.batchSendVtPointData(results)
				results = make([]VirtualPoint, 0)
				curTime = ""
			}
		}
	}
}
