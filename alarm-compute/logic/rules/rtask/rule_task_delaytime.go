package rtask

import (
	"fmt"
	"strings"
	"time"

	"etrpc-go/log"

	"github.com/samber/lo"

	"alarm-compute/entity/epoint"
	"alarm-compute/entity/taskcode"
	"alarm-compute/logic/lcache"
	"alarm-compute/utils/modcall"
)

// StartDelayTimeRuleTask StartDelayTimeRuleTask
func (rt *RuleTask) StartDelayTimeRuleTask(pointValue epoint.HistoryValueMap, ts int64) (success bool) {
	var fired bool
	var restored bool
	var err error
	startTime := time.Now()
	defer func() {
		modcall.RecordAlarmComputeTime(rt.ServiceName(), float64(time.Since(startTime).Milliseconds()))
		rt.RecordAnalyzeProcess(ts, fired, restored, err)
	}()
	missVarStr, err := rt.CheckMissDelayPointList(ts, pointValue, false)
	if err != nil {
		err = taskcode.NewErr(&taskcode.PointDataLackErr, missVarStr)
		return
	}
	historyAlertPV, fired, err := rt.Alert.StartDelayTimeAnalyze(pointValue, ts)
	if err != nil {
		// 失败处理
		log.Warnf("mozu:%d, dt task not running:%s, exp err: %v", rt.MozuId, rt.GetKey(), err)
		err = taskcode.NewErr(&taskcode.ExprAnalyzeErr, err.Error())
		return
	}
	if fired {
		// 告警逻辑
		fileAlert, err := rt.Alert.GeneFireAlert(ts, nil, historyAlertPV)
		if err != nil {
			log.Errorf("mozu:%d, dt Gene FireAlert failed, key %s", rt.MozuId, rt.GetKey(), err.Error())
		}
		_ = rt.SendAlert(ts, fileAlert)
	} else {
		// 告警恢复分析
		restored, err = rt.processDelayTimeRestore(pointValue, ts)
		if err != nil {
			log.Warnf("mozu:%d, dt processRestore failed, rule task %s restore expr: %s, err: %s",
				rt.MozuId, rt.GetKey(), rt.Restore.Exp.Express, err.Error())
			err = taskcode.NewErr(&taskcode.RestoreAnalyzeErr, err.Error())
			return
		}
	}
	if err != nil {
		err = taskcode.NewErr(&taskcode.UnKnownErr, err.Error())
	}
	success = err == nil
	return
}

func (rt *RuleTask) processDelayTimeRestore(checkValue epoint.HistoryValueMap, ts int64) (restored bool, err error) {
	hasActive := lcache.GetLocalCache().CheckActiveAlarmCache(rt.GetKey())
	if !hasActive {
		return
	}
	historyRestorePV, restored, err := rt.Restore.StartDelayTimeAnalyze(checkValue, ts)
	if err != nil {
		err = fmt.Errorf("延时告警恢复计算失败; %s", err.Error())
		return
	}
	if restored {
		restoreAlert, geneErr := rt.Restore.GeneFireAlert(ts, nil, historyRestorePV)
		if geneErr != nil {
			err = fmt.Errorf("生成延时告警恢复信息失败 %s", geneErr.Error())
			return
		}
		sendErr := rt.SendRestoreAlert(ts, restoreAlert)
		if sendErr != nil {
			err = fmt.Errorf("发送延时告警恢复消息失败 %s", sendErr.Error())
		}
	}
	return
}

// CheckMissDelayPointList CheckMissDelayPointList
func (rt *RuleTask) CheckMissDelayPointList(ts int64, pointValue epoint.HistoryValueMap, isVirtual bool) (missVarStr string, err error) {
	missList := []string{}
	if !isVirtual {
		_, restoreMissList := rt.Restore.CheckMissDelayPointList(ts, pointValue)
		if len(restoreMissList) > 0 {
			// 恢复缺测点数据，但不影响触发，可以继续走
			missList = append(missList, restoreMissList...)
		}
	}
	alertMissVarList, alertMissList := rt.Alert.CheckMissDelayPointList(ts, pointValue)
	if len(alertMissList) > 0 {
		// 触发缺测点数据，直接报错，影响生效率，把恢复缺的测点也一同上报
		missList = append(missList, alertMissList...)
		uniqueMissList := lo.Uniq(missList)
		vMsg := strings.Join(uniqueMissList, ",")
		err = fmt.Errorf("miss points: %v", vMsg)
		missVarStr = strings.Join(alertMissVarList, ",")
		return
	}
	if len(missList) > 0 {
		// 到这里有缺测点数据，说明触发没缺，恢复缺，打印 warning 日志
		log.Warnf("restore miss points: %v", missList)
	}
	return
}
