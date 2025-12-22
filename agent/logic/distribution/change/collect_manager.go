package change

import (
	"agent/entity/consts"
	"agent/logic/collector/rtdb/model"
	"agent/logic/distribution/base"
	"agent/logic/distribution/distributor"
	utils2 "agent/logic/distribution/distributor/utils"
	"agent/repo/monitor"
	"agent/utils"
	"sync"
	"time"

	"agent/entity/definition"
	model2 "agent/entity/model/data"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	notChangedPoints = "not_changed_points"
	callbackPoints   = "callback_points"
)

var (
	collectInstance *CollectChangeManager
	pc              pointCount
)

type deviceDataType map[definition.DataPointIDType]model.DataPoint
type cacheDataType = map[definition.DeviceGidType]deviceDataType
type firstDataType map[definition.DataPointIDType]struct{}

// CollectChangeManager 采集测点变更管理器
type CollectChangeManager struct {
	cacheMutex   sync.RWMutex
	wg           sync.WaitGroup
	cacheData    cacheDataType
	first        firstDataType
	stopped      bool
	distributors distributor.Distributors
}

func newCollectChangeManager() *CollectChangeManager {
	collectInstance = &CollectChangeManager{
		cacheData:    make(cacheDataType),
		first:        make(firstDataType),
		distributors: base.GetDistributorList(consts.CollectChange),
	}
	return collectInstance
}

func callback(points model.DataPoints, _ interface{}) interface{} {
	collectInstance.cacheMutex.Lock()
	defer collectInstance.cacheMutex.Unlock()

	notChangedPointsNum := uint64(0)
	for i := range points {
		p := &points[i]
		if !p.IsValueChanged {
			if _, has := collectInstance.first[p.ID]; has {
				notChangedPointsNum++
				continue
			}
			collectInstance.first[p.ID] = struct{}{}
		}

		device := p.DeviceGiD
		deviceData, ok := collectInstance.cacheData[device]
		if !ok {
			deviceData = make(deviceDataType)
			collectInstance.cacheData[device] = deviceData
		}
		deviceData[p.ID] = *p
	}
	pc.Add(notChangedPointsNum, uint64(len(points)))
	return nil
}

func (w *CollectChangeManager) start() {
	w.wg.Add(1)
	w.stopped = false
	go w.loop()
	go w.report()
}

func (w *CollectChangeManager) stop() {
	w.stopped = true
	w.wg.Wait()
}

func (w *CollectChangeManager) processDataPoints(d definition.DeviceGidType, points deviceDataType) {
	ps := make(model.DataPoints, 0, len(points))
	for i := range points {
		ps = append(ps, points[i])
	}

	data := &model2.DataUnit{
		DeviceGid: d,
		Points:    ps,
	}
	arg := utils2.DistributorArgs{
		Time:     utils.GetNowUTCTime(),
		Interval: 1,
	}
	w.distributors.BatchDistribute(data, &arg)

	//log.Debugf("push changed collect points, num: %v", len(points))
}

func (w *CollectChangeManager) process() {
	w.cacheMutex.Lock()
	defer w.cacheMutex.Unlock()

	l := len(w.cacheData)
	if l == 0 {
		return
	}

	removeDevice := make([]definition.DeviceGidType, 0, len(w.cacheData))
	for device, deviceData := range w.cacheData {
		if len(deviceData) == 0 {
			continue
		}
		removeDevice = append(removeDevice, device)
		w.processDataPoints(device, deviceData)
	}

	for _, device := range removeDevice {
		w.cacheData[device] = make(deviceDataType)
	}
}

func (w *CollectChangeManager) loop() {
	defer w.wg.Done()

	for {
		if w.stopped {
			return
		}
		select {
		case <-time.After(time.Second):
			if w.stopped {
				return
			}
			w.process()
		}
	}
}

func (w *CollectChangeManager) report() {
	for {
		if w.stopped {
			return
		}
		n, m := pc.Get()

		monitor.ReportMultiMetricsWithDimensions(
			[]*metrics.Metrics{
				metrics.NewMetrics(notChangedPoints, float64(n), metrics.PolicySUM),
				metrics.NewMetrics(callbackPoints, float64(m), metrics.PolicySUM),
			},
			getCollectDimensions(),
		)
		time.Sleep(time.Second)
	}
}

func getCollectDimensions() (dimensions []*metrics.Dimension) {
	attrsList := []*metrics.Dimension{
		{
			Name:  consts.AttrDeviceName,
			Value: "collect",
		},
	}

	return attrsList
}
