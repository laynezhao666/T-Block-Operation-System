package opc

import (
	rtdbModel "agent/logic/collector/rtdb/model"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	dataCacheTimeout = 1 * time.Minute
)

// dataCache holds OPC UA subscription cache and mappings.
type dataCache struct {
	mu sync.Mutex

	// point values: reportId -> RTValue
	pointValue map[string]rtdbModel.RTValue

	// mappings between handles, nodeIDs and reportIDs
	handle2Point     map[uint32]string // handle -> nodeID
	handle2Report    map[uint32]string // handle -> reportId
	point2Handle     map[string]uint32 // nodeID -> handle
	reportId2PointId map[string]string // reportId -> nodeID
	handleIndex      uint32

	// subscription-level last ok time
	lastOkTime time.Time

	// 推送计数器，用于区分初始值推送和后续变化推送
	pushCount int
}

// NewDataCache creates a fresh per-device cache.
func NewDataCache() *dataCache {
	return &dataCache{
		pointValue:       make(map[string]rtdbModel.RTValue),
		handle2Point:     make(map[uint32]string),
		handle2Report:    make(map[uint32]string),
		point2Handle:     make(map[string]uint32),
		reportId2PointId: make(map[string]string),
		lastOkTime:       time.Now(),
		pushCount:        0,
	}
}

// SetLastOk updates subscription-level last ok time.
func (d *dataCache) SetLastOk() {
	d.mu.Lock()
	d.lastOkTime = time.Now()
	d.mu.Unlock()
}

// IncrPushCount 推送计数+1，返回新值（用于区分初始值推送和后续变化推送）
func (d *dataCache) IncrPushCount() int {
	d.mu.Lock()
	d.pushCount++
	c := d.pushCount
	d.mu.Unlock()
	return c
}

// ResetPushCount 重置推送计数（重新订阅时调用）
func (d *dataCache) ResetPushCount() {
	d.mu.Lock()
	d.pushCount = 0
	d.mu.Unlock()
}

// GetLastOkTime returns the last ok time (for logging/debug).
func (d *dataCache) GetLastOkTime() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastOkTime
}

// IsTimeout returns whether the subscription cache is stale globally.
func (d *dataCache) IsTimeout() bool {
	d.mu.Lock()
	t := d.lastOkTime
	d.mu.Unlock()
	return time.Since(t) > dataCacheTimeout
}

// GetPointValue returns the cached value by reportId.
func (d *dataCache) GetPointValue(reportId string) (rtdbModel.RTValue, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	v, ok := d.pointValue[reportId]
	return v, ok
}

// GetNodeIDFromReportID returns the nodeID for a reportId.
func (d *dataCache) GetNodeIDFromReportID(reportId string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	nid, ok := d.reportId2PointId[reportId]
	return nid, ok
}

// SetPointValue sets cached value for a point by handle.
// It resolves handle -> reportId, then writes pointValue[reportId] = v.
func (d *dataCache) SetPointValue(handle uint32, v rtdbModel.RTValue) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	rid, ok := d.handle2Report[handle]
	if !ok {
		return false
	}
	d.pointValue[rid] = v
	return true
}

// SetPointValueByReportID writes value by reportId directly.
func (d *dataCache) SetPointValueByReportID(reportId string, v rtdbModel.RTValue) {
	d.mu.Lock()
	d.pointValue[reportId] = v
	d.mu.Unlock()
}

// InvalidateByHandle deletes cached value for a monitored item by handle.
func (d *dataCache) InvalidateByHandle(handle uint32) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	rid, ok := d.handle2Report[handle]
	if !ok {
		return false
	}
	delete(d.pointValue, rid)
	return true
}

// InvalidateByReportID deletes cached value for a point by reportId.
func (d *dataCache) InvalidateByReportID(reportId string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.pointValue[reportId]; !ok {
		return false
	}
	delete(d.pointValue, reportId)
	return true
}

// InvalidateAllForSubscription clears all cached point values.
// Use on subscription-level failures.
func (d *dataCache) InvalidateAllForSubscription() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for rid := range d.pointValue {
		delete(d.pointValue, rid)
	}
}

// GenMonitoredItemRequest generates create request and maintains handle mappings.
func (d *dataCache) GenMonitoredItemRequest(nodeID string, reportId string) (*ua.MonitoredItemCreateRequest, error) {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		log.Errorf("ua.ParseNodeID err:%v", err)
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	handle, ok := d.point2Handle[nodeID]
	if !ok {
		// new handle
		d.handleIndex++
		handle = d.handleIndex
		d.point2Handle[nodeID] = handle
		d.handle2Point[handle] = nodeID
		d.handle2Report[handle] = reportId
		d.reportId2PointId[reportId] = nodeID
	}
	return opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, handle), nil
}

// GetNodeIDByHandle returns the nodeID for a given handle (for debug logging).
func (d *dataCache) GetNodeIDByHandle(handle uint32) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	nid, ok := d.handle2Point[handle]
	return nid, ok
}

// ListAllReportNodePairs returns a snapshot of (reportId, nodeID) pairs for re-subscription.
func (d *dataCache) ListAllReportNodePairs() [][2]string {
	d.mu.Lock()
	defer d.mu.Unlock()
	pairs := make([][2]string, 0, len(d.reportId2PointId))
	for rid, nid := range d.reportId2PointId {
		pairs = append(pairs, [2]string{rid, nid})
	}
	return pairs
}
