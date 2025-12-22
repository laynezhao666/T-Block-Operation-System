package manager

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/avast/retry-go"
	"github.com/panjf2000/ants/v2"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/conf"
	"alarm-manage/logic/notification"
	"alarm-manage/repo/db"
	"alarm-manage/utils/batch"
	"alarm-manage/utils/common"
	cmodel "common/entity/model"
)

func (m *Manager) processRestoreAlert(ctx context.Context) {
	var doJobTask = func(i interface{}) error {
		activeList := i.([]*cmodel.AlarmHistory)
		if len(activeList) == 0 {
			return nil
		}
		m.processEdgeRestoreAlert(ctx, activeList)
		return nil
	}
	var poolWg sync.WaitGroup
	wpSize := conf.ServerConf.RestoreManageConfig.PoolSize
	wp, _ := ants.NewPoolWithFunc(int(wpSize), func(i interface{}) {
		doJobTask(i)
		poolWg.Done()
	}, ants.WithNonblocking(false))
	defer wp.Release()
	size := conf.ServerConf.RestoreManageConfig.BatchChannelSize
	interval := time.Duration(conf.ServerConf.RestoreManageConfig.BatchFetchIntervalMS) * time.Millisecond
	historyChannel := batch.BatchChannel(ctx, m.restoringCh, int(size), interval)
	for {
		select {
		case <-ctx.Done():
			poolWg.Wait()
			return
		case list := <-historyChannel:
			edgeRestores := m.splitHistoryAlarms(list)
			poolWg.Add(1)
			wp.Invoke(edgeRestores)
		}
	}
}

func (m *Manager) splitHistoryAlarms(list *batch.GroupData) (edgeAlarms []*cmodel.AlarmHistory) {
	for _, item := range list.Data {
		ac := item.(*cmodel.AlarmHistory)
		if !isAlarmHistoryValid(ac) {
			continue
		}
		edgeAlarms = append(edgeAlarms, ac)
	}

	return
}

func isAlarmHistoryValid(ac *cmodel.AlarmHistory) bool {
	if ac.DeviceGid == "" || ac.Rid == 0 {
		return false
	}
	return true
}

func (m *Manager) processEdgeRestoreAlert(ctx context.Context, historys []*cmodel.AlarmHistory) (err error) {
	dbRestoreList, err := BatchAutoRestoreAlert(historys)
	log.Infof("receive restore alerts, historys len: %d, dbRestoreList len: %d",
		len(historys), len(dbRestoreList))

	if err != nil {
		bjson, _ := json.Marshal(historys)
		log.AlarmContextf(trpc.BackgroundContext(), "告警恢复操作数据库失败, alert: %v, error: %s", string(bjson), err.Error())
	}
	// TODO 告警恢复后续处理，上报恢复告警等
	if len(dbRestoreList) > 0 {
		go notification.GetSpeaker().ReportRestore(dbRestoreList)
	}
	return
}

// BatchAutoRestoreAlert 批量自动恢复告警
func BatchAutoRestoreAlert(historys []*cmodel.AlarmHistory) ([]cmodel.AlarmHistory, error) {
	restoreList, err := BatchFillRestoreAlert(historys)
	if err != nil {
		// retry
		log.Warnf("BatchFillRestoreAlert failed, retry, err: %v", err)
		time.Sleep(100 * time.Microsecond)
		restoreList, err = BatchFillRestoreAlert(historys)
		if err != nil {
			return nil, err
		}
	}
	if len(restoreList) == 0 {
		// 没有活动告警，已经都恢复了
		log.Infof("all history no active, history: %v", common.JSONMarshalNoErr(historys))
		return restoreList, nil
	}
	var successRestoreList []cmodel.AlarmHistory
	err = retry.Do(func() error {
		var retryErr error
		successRestoreList, retryErr = db.GetAlarmDBImpl().RestoreAlerts(restoreList)
		return retryErr
	}, retry.Attempts(3), retry.RetryIf(func(retryErr error) bool {
		return retryErr != nil
	}))
	if err != nil {
		return successRestoreList, err
	}
	return successRestoreList, nil
}

// BatchFillRestoreAlert 填充恢复告警
func BatchFillRestoreAlert(historys []*cmodel.AlarmHistory) ([]cmodel.AlarmHistory, error) {
	var err error
	returnHistoryList := make([]cmodel.AlarmHistory, 0)
	// 去重本批
	uniqueFpList, uniqueList := uniqueHistoryList(historys)
	activeList, err := db.GetAlarmDBImpl().GetActiveListByFp(uniqueFpList)
	if err != nil {
		log.Errorf("select active error:%s, alarmids:%v",
			err.Error(), uniqueFpList)
		return nil, err
	}
	activeFpMap := make(map[string]cmodel.AlarmActive)
	for _, item := range activeList {
		activeFpMap[item.FingerPrint] = item
	}
	log.Infof("activeList len: %v, activeFpMap len: %v",
		len(activeList), len(activeFpMap))
	if len(activeFpMap) == 0 {
		return returnHistoryList, nil
	}
	for _, historyItem := range uniqueList {
		historyFp := historyItem.FingerPrint
		active, ok := activeFpMap[historyFp]
		if !ok {
			continue
		}
		if active.OccurTime.After(historyItem.RestoreTime) {
			// 触发时间大于恢复时间，过滤
			// 采集侧在方仓中断时会把最后一次的数据保留，下一次继续采集方仓时会重发这次数据
			log.Infof("active OccurTime is after history RestoreTime (%v, %v), alarmID: %v",
				active.OccurTime, historyItem.RestoreTime, active.AlarmID)
			continue
		}
		historyAlert := cmodel.ActiveAlert2History(&active)
		historyAlert.RestoreTime = historyItem.RestoreTime
		historyAlert.RestoreAnalyzeResult = historyItem.AnalyzeResult
		historyAlert.CreateAt = time.Now()
		returnHistoryList = append(returnHistoryList, *historyAlert)
	}
	return returnHistoryList, nil
}

func uniqueHistoryList(historys []*cmodel.AlarmHistory) ([]string, []*cmodel.AlarmHistory) {
	list := make(map[string]*cmodel.AlarmHistory)
	for _, item := range historys {
		fp := item.FingerPrint
		if v, ok := list[fp]; ok {
			log.Infof("fp duplicated, v: %v, item: %v", *v, *item)
			if item.RestoreTime.Before(v.RestoreTime) {
				// 使用更早的恢复时间
				list[fp] = item
			}
		} else {
			// 不存在，保存起来
			list[fp] = item
		}
	}

	// 去重本批
	uniqueFpList := make([]string, 0)
	uniqueList := make([]*cmodel.AlarmHistory, 0)
	for _, item := range list {
		uniqueList = append(uniqueList, item)
		uniqueFpList = append(uniqueFpList, item.FingerPrint)
	}

	return uniqueFpList, uniqueList
}
