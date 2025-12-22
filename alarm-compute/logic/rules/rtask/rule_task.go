package rtask

import (
	"fmt"
	"strings"
	"time"

	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"

	"github.com/samber/lo"

	"alarm-compute/entity"
	"alarm-compute/entity/taskcode"
	"alarm-compute/logic/collector/validate"
	"alarm-compute/logic/lcache"
	"alarm-compute/utils/modcall"
)

const (
	// rid:gid
	RuleTaskKeyTemplate = "%d;%s"
)

// RuleTask 每一个gid 对应的一条告警计算策略
type RuleTask struct {
	Version         string // 任务版本
	Key             string
	Rid             int64
	MozuId          int64
	Gid             string
	Level           string
	AnalyzeTimeType taskcode.RuleTaskTimeType // 任务类型 实时/延时
	AlarmName       string
	ContentTemplate string
	Alert           *AlarmTask       // 告警计算任务
	Restore         *AlarmTask       // 恢复计算任务
	PointDelayMap   map[string]int32 // 任务所依赖的测点最大延迟map
}

// RuleTaskAnalyzeType 任务告警分析类型
type RuleTaskAnalyzeType int64

// RuleTaskAnalyzeTypeAlarm 告警
const (
	RuleTaskAnalyzeTypeAlarm = RuleTaskAnalyzeType(0)

	// RuleTaskAnalyzeTypeRestore  恢复
	RuleTaskAnalyzeTypeRestore = RuleTaskAnalyzeType(1)
)

// ServiceName ServiceName
func (rt *RuleTask) ServiceName() string {
	if rt.AnalyzeTimeType == taskcode.RuleTaskRealtime {
		return "RealtimeRT"
	} else if rt.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
		return "DelaytimeRT"
	} else {
		return "VirtualRT"
	}
}

// GetKey GetKey
func (rt *RuleTask) GetKey() string {
	return rt.Key
}

// SetKey SetKey
func (rt *RuleTask) SetKey() {
	rt.Key = fmt.Sprintf(RuleTaskKeyTemplate, rt.Rid, rt.Gid)
}

// SetAnalyzeTimeType 设置分析时间类型
func (rt *RuleTask) SetAnalyzeTimeType(ridType int64) error {
	if ridType == 0 {
		rt.AnalyzeTimeType = taskcode.RuleTaskRealtime
	} else if ridType == 1 {
		rt.AnalyzeTimeType = taskcode.RuleTaskDelaytime
	} else if ridType == 2 {
		rt.AnalyzeTimeType = taskcode.RuleTaskVirtual
	} else {
		return fmt.Errorf("unknown rid type")
	}
	return nil
}

// SetPointDelayMap 设置测点最大延迟秒数
func (rt *RuleTask) SetPointDelayMap() {
	pointDelayMap := make(map[string]int32)
	defer func() {
		rt.PointDelayMap = pointDelayMap
	}()
	if rt.AnalyzeTimeType == taskcode.RuleTaskRealtime {
		return
	}
	alertDelayMap := rt.Alert.GetMaxPointDelayMap()
	for k, v := range alertDelayMap {
		pointDelayMap[k] = v
	}
	// 虚拟策略不关心恢复计算任务
	if rt.AnalyzeTimeType != taskcode.RuleTaskVirtual {
		restoreDelayMap := rt.Restore.GetMaxPointDelayMap()
		for k, v := range restoreDelayMap {
			if _, ok := pointDelayMap[k]; !ok {
				pointDelayMap[k] = v
			} else {
				pointDelayMap[k] = max(v, pointDelayMap[k])
			}
		}
	}
}

// GetPointDelayMap 设置测点最大延迟秒数
func (rt *RuleTask) GetPointDelayMap() map[string]int32 {
	return rt.PointDelayMap
}

// Stop Stop
func (rt *RuleTask) Stop() {}

// GetPointList GetPointList
func (rt *RuleTask) GetPointList() []string {
	return lo.Union(rt.Alert.GetPointList(), rt.Restore.GetPointList())
}

// SetAlarmTaskExp SetAlarmTaskExp
func (rt *RuleTask) SetAlarmTaskExp(config *entity.AlarmConfig) error {
	exprMap := config.ExpressionMap
	err := rt.Alert.SetExp(config.AlarmExpression, &exprMap.Fire)
	if err != nil {
		return err
	}
	err = rt.Restore.SetExp(config.RestoreExpression, &exprMap.Restore)
	if err != nil {
		return err
	}
	return nil
}

// SetValidateRuleMsg 设置策略生效信息
func (rt *RuleTask) SetValidateRuleMsg(r *RuleTask, rTime int64, fired bool, err *taskcode.TaskStatusErr) *pb.ValidateTaskItem {
	ret := &pb.ValidateTaskItem{
		MozuId:     r.MozuId,
		Rid:        r.Rid,
		AlarmLevel: r.Level,
		Version:    r.Version,
		RidType:    int64(r.AnalyzeTimeType),
		Gid:        r.Gid,
		RunTime:    rTime,
		Fired:      fired,
	}
	if err != nil {
		ret.ErrorCode = int64(err.GetErrCode())
		ret.ErrorName = err.Error()
		ret.ErrorDetail = err.GetErrDetail()
		if len(ret.ErrorDetail) > 100 {
			ret.ErrorDetail = ret.ErrorDetail[:100] + "..."
		}
	} else {
		ret.Successed = true
	}
	return ret
}

// RecordAnalyzeProcess 记录策略分析进度
func (rt *RuleTask) RecordAnalyzeProcess(ts int64, fired bool, restored bool, err error) {
	var taskErr *taskcode.TaskStatusErr
	if err != nil {
		if e, ok := err.(*taskcode.TaskStatusErr); ok {
			taskErr = e
		} else {
			taskErr = &taskcode.UnKnownErr
		}
	}
	if taskErr != nil && taskErr.JudgeErrType(&taskcode.PointDataLackErr) {
		if rt.AnalyzeTimeType == taskcode.RuleTaskRealtime {
			validate.GetFailRuleCollector().AddFailedRt(rt.GetKey())
		} else if rt.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
			validate.GetFailRuleCollector().AddFailedDt(rt.GetKey())
		}
	}
	validMsgPb := rt.SetValidateRuleMsg(rt, ts, fired, taskErr)
	validate.GetTaskCollector().AddValidateRecord(validMsgPb)
}

// CheckMissRealPointList 检查实时分析中测点缺失情况
func (rt *RuleTask) CheckMissRealPointList(pointValue map[string]float64) (missVarStr string, err error) {
	missList := []string{}

	_, restoreMissList := rt.Restore.CheckMissRealPointList(pointValue)
	if len(restoreMissList) > 0 {
		// 恢复缺测点数据，但不影响触发，可以继续走
		missList = append(missList, restoreMissList...)
	}
	missVarList, alertMissList := rt.Alert.CheckMissRealPointList(pointValue)
	if len(alertMissList) > 0 {
		// 触发缺测点数据，直接报错，影响生效率，把恢复缺的测点也一同上报
		missList = append(missList, alertMissList...)
		uniqueMissList := lo.Uniq(missList)
		vMsg := strings.Join(uniqueMissList, ",")
		err = fmt.Errorf("miss points: %v", vMsg)
		missVarStr = strings.Join(missVarList, ",")
		return
	}
	if len(missList) > 0 {
		log.Warnf("restore miss points: %v", missList)
	}
	return
}

// StartRealtimeByData StartRealtimeByData
func (rt *RuleTask) StartRealtimeByData(pointValue map[string]float64, ts int64) (success bool) {
	var fired bool
	var restored bool
	var err error
	startTime := time.Now()
	defer func() {
		modcall.RecordAlarmComputeTime(rt.ServiceName(), float64(time.Since(startTime).Milliseconds()))
		rt.RecordAnalyzeProcess(ts, fired, restored, err)
	}()
	missVarStr, err := rt.CheckMissRealPointList(pointValue)
	if err != nil {
		err = taskcode.NewErr(&taskcode.PointDataLackErr, missVarStr)
		return
	}
	alertExecPV, fired, err := rt.Alert.StartRealTimeAnalyze(pointValue, ts)
	if err != nil {
		// 失败处理
		log.Warnf("mozu:%d, rt task not running:%s, exp err: %v", rt.MozuId, rt.GetKey(), err)
		err = taskcode.NewErr(&taskcode.ExprAnalyzeErr, err.Error())
		return
	}
	if fired {
		// 实时策略 发送告警前的判断逻辑
		fileAlert, err := rt.Alert.GeneFireAlert(ts, alertExecPV, nil)
		if err != nil {
			log.Errorf("mozu:%d, rt Gene FireAlert failed, key %s", rt.MozuId, rt.GetKey(), err.Error())
		}
		_ = rt.SendAlert(ts, fileAlert)
	} else {
		// 告警恢复分析
		restored, err = rt.processRealTimeRestore(pointValue, ts)
		if err != nil {
			log.Warnf("mozu:%d, rt processRestore failed, rule task %s restore expr: %s, err: %s",
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

func (rt *RuleTask) processRealTimeRestore(checkValue map[string]float64, ts int64) (restored bool, err error) {
	hasActive := lcache.GetLocalCache().CheckActiveAlarmCache(rt.GetKey())
	if !hasActive {
		return
	}
	restoreExecPV, restored, err := rt.Restore.StartRealTimeAnalyze(checkValue, ts)
	if err != nil {
		err = fmt.Errorf("告警恢复计算任务失败: %s", err.Error())
		return
	}
	if restored {
		restoreAlert, geneErr := rt.Restore.GeneFireAlert(ts, restoreExecPV, nil)
		if geneErr != nil {
			err = fmt.Errorf("告警恢复消息产生失败: %s", geneErr.Error())
			return
		}
		sendErr := rt.SendRestoreAlert(ts, restoreAlert)
		if sendErr != nil {
			err = fmt.Errorf("告警恢复消息发送失败: %s", sendErr.Error())
		}
	}
	return
}
