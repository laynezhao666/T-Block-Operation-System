// Package manager manager
package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "trpcprotocol/alarm-manage"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"

	cmodel "common/entity/model"
)

var gManager *Manager
var once sync.Once

const (
	FingerPrintTemplate = "%d;%s"
)

// GetGlobalManager GetGlobalManager
func GetGlobalManager() *Manager {
	once.Do(func() {
		gManager = &Manager{
			alertingCh:  make(chan interface{}, 20000),
			restoringCh: make(chan interface{}, 5000),
		}
	})

	return gManager
}

// Manager Manager
type Manager struct {
	alertingCh  chan interface{}
	restoringCh chan interface{}
}

// Run Run
func (m *Manager) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	go m.processRestoreAlert(ctx)
	m.processFireAlert(ctx)
}

// GetAlertingCh GetAlertingCh
func (m *Manager) GetAlertingCh() chan interface{} {
	return m.alertingCh
}

// GetRestoringCh GetRestoringCh
func (m *Manager) GetRestoringCh() chan interface{} {
	return m.restoringCh
}

// GeneFingerPrint GeneFingerPrint
func (m *Manager) GeneFingerPrint(alertMsg *pb.AlarmMsgPb) string {
	ret := fmt.Sprintf(FingerPrintTemplate, alertMsg.Rid, alertMsg.Gid)
	return ret
}

// AddAlertToCh AddAlertToCh
// 判断告警kafka是否有消息堆积
func (m *Manager) AddAlertToCh(alertMsg *pb.AlarmMsgPb) {
	now := time.Now()
	dbAlert := &cmodel.AlarmActive{
		OccurTime:     time.Unix(alertMsg.StartAt, 0),
		Content:       alertMsg.Content,
		FingerPrint:   m.GeneFingerPrint(alertMsg),
		AnalyzeResult: alertMsg.AnalyzeResult,
		Rid:           alertMsg.Rid,
		DeviceGid:     alertMsg.Gid,
		Level:         alertMsg.Level,
		AlarmName:     alertMsg.AlarmName,
		MozuId:        int64(alertMsg.MozuId),
		UpdateTime:    time.Unix(alertMsg.StartAt, 0),
		EventStatus:   1,
	}
	if dbAlert.OccurTime.Before(now.Add(-3 * time.Minute)) {
		log.AlarmContextf(trpc.BackgroundContext(), "告警消费速度过慢, 当前处理时间:%s, 告警消息%+v", now.Format(time.DateTime), alertMsg)
		return
	}
	if dbAlert.OccurTime.Before(now.Add(-1 * time.Minute)) {
		log.Errorf("告警消费速度较慢:当前处理时间:%s, 告警消息%+v", now.Format(time.DateTime), alertMsg)
	}
	m.alertingCh <- dbAlert
}

// AddRestoreToCh AddRestoreToCh
// 判断告警恢复kafka是否有消息堆积
func (m *Manager) AddRestoreToCh(restoreMsg *pb.AlarmMsgPb) {
	now := time.Now()
	dbRestore := &cmodel.AlarmHistory{
		FingerPrint:   m.GeneFingerPrint(restoreMsg),
		AnalyzeResult: restoreMsg.AnalyzeResult,
		MozuId:        int64(restoreMsg.MozuId),
		Rid:           restoreMsg.Rid,
		DeviceGid:     restoreMsg.Gid,
		RestoreTime:   time.Unix(restoreMsg.EndAt, 0),
	}
	if dbRestore.RestoreTime.Before(now.Add(-3 * time.Minute)) {
		log.AlarmContextf(trpc.BackgroundContext(), "告警恢复消费速度过慢, 当前处理时间:%s, 告警消息%+v", now.Format(time.DateTime), restoreMsg)
		return
	}
	if dbRestore.RestoreTime.Before(now.Add(-1 * time.Minute)) {
		log.Errorf("告警恢复消费速度较慢:当前处理时间:%s, 告警消息%+v", now.Format(time.DateTime), restoreMsg)
	}
	m.restoringCh <- dbRestore
}
