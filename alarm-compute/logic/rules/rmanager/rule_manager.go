package rmanager

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"

	"alarm-compute/entity"
	"alarm-compute/entity/taskcode"
	"alarm-compute/logic/rules/rtask"
	"alarm-compute/utils/common"
)

const (
	// rid-gid-version
	ErrParseConfigKeyTemplate = "%d-%s-%s"
)

var (
	gRuleManager *RuleManager
	once         sync.Once
)

// GetGlobalRuleManager GetGlobalRuleManager
func GetGlobalRuleManager() *RuleManager {
	once.Do(func() {
		gRuleManager = &RuleManager{
			RtTotalExecChan: make(chan struct{}, 10),
			DtTotalExecChan: make(chan struct{}, 10),
			VtTotalExecChan: make(chan struct{}, 10),
		}
	})
	return gRuleManager
}

// RuleMap RuleMap
type RuleMap struct {
	sync.Map // map[string]*rtask.RuleTask   rule_key -> rule
}

// RDelete RDelete
func (rm *RuleMap) RDelete(key string) {
	r, ok := rm.Load(key)
	if ok {
		rule := r.(*rtask.RuleTask)
		rule.Stop()
		rm.Delete(key)
		log.Infof("delete rule success, key: %v, rule: %+v", key, rule)
	}
}

// RGetAll RGetAll
func (rm *RuleMap) RGetAll() map[string]*rtask.RuleTask {
	rules := make(map[string]*rtask.RuleTask)
	rm.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		rules[ruleKey] = value.(*rtask.RuleTask)
		return true
	})
	return rules
}

// PointRuleMap PointRuleMap
type PointRuleMap struct {
	// 使用slice代替Set 存储ruleKey
	// 删除频率较低，运行时间频繁遍历，因此采用slice
	sync.Map // map[string] []ruleKey
}

// PRDelete PRDelete
func (prm *PointRuleMap) PRDelete(rule *rtask.RuleTask) {
	pointList := rule.GetPointList()
	for _, point := range pointList {
		pointRuleList, ok := prm.Load(point)
		if ok {
			newRuleList := common.RemoveEleFromSlice(pointRuleList.([]string), rule.GetKey())
			if len(newRuleList) == 0 {
				prm.Delete(point)
			} else {
				prm.Store(point, newRuleList)
			}
		}
	}
}

// PRStore PRStore
func (prm *PointRuleMap) PRStore(rule *rtask.RuleTask) {
	pointList := rule.GetPointList()
	for _, point := range pointList {
		pointRuleList, ok := prm.Load(point)
		if ok {
			pointRuleList = append(pointRuleList.([]string), rule.GetKey())
		} else {
			pointRuleList = []string{rule.GetKey()}
		}
		prm.Store(point, pointRuleList)
	}
}

// GetTotalPointList 获取全部测点
func (prm *PointRuleMap) GetTotalPointList() []string {
	pointList := []string{}
	prm.Range(func(key, value interface{}) bool {
		pointList = append(pointList, key.(string))
		return true
	})
	return pointList
}

// RuleManager RuleManager
type RuleManager struct {

	// 实时策略 -> 单测点 多测点 跨设备
	RealTimeRule RuleMap
	// 实时策略 测点名称 -> ruleKey映射
	RtPointRuleMap PointRuleMap // map[string][rule_key]string
	// 通道信号，控制下个周期执行全量策略
	RtTotalExecChan chan struct{}

	// 延时策略 -> 单测点 多测点 跨设备
	DelayTimeRule RuleMap
	// 延时策略 测点名称 -> ruleKey映射
	DtPointRuleMap PointRuleMap // map[string][rule_key]string
	// 通道信号，控制下个周期执行全量策略
	DtTotalExecChan chan struct{}

	// 虚拟点计算策略
	VirtualRule           RuleMap
	VtPointRuleMap        PointRuleMap // map[string][rule_key]string
	VtTotalExecChan       chan struct{}
	EmptyVaryVtPointCount int
}

// GetRealtimeRules 获取全部实时策略
func (m *RuleManager) GetRealtimeRules() map[string]*rtask.RuleTask {
	rules := make(map[string]*rtask.RuleTask)
	m.RealTimeRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		rules[ruleKey] = value.(*rtask.RuleTask)
		return true
	})
	return rules
}

// GetDelaytimeRules 获取全部延时策略
func (m *RuleManager) GetDelaytimeRules() map[string]*rtask.RuleTask {
	rules := make(map[string]*rtask.RuleTask)
	m.DelayTimeRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		rules[ruleKey] = value.(*rtask.RuleTask)
		return true
	})
	return rules
}

// GetVirtualRules 获取全部虚拟策略
func (m *RuleManager) GetVirtualRules() map[string]*rtask.RuleTask {
	rules := make(map[string]*rtask.RuleTask)
	m.VirtualRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		rules[ruleKey] = value.(*rtask.RuleTask)
		return true
	})
	return rules
}

func (m *RuleManager) getTotalRuleKeySet() map[string]struct{} {
	keySet := map[string]struct{}{}
	m.RealTimeRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		keySet[ruleKey] = struct{}{}
		return true
	})
	m.DelayTimeRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		keySet[ruleKey] = struct{}{}
		return true
	})
	m.VirtualRule.Range(func(key, value interface{}) bool {
		ruleKey := key.(string)
		keySet[ruleKey] = struct{}{}
		return true
	})
	return keySet
}

func (m *RuleManager) getRuleByRuleKey(ruleKey string) (*rtask.RuleTask, bool) {
	if r, ok := m.RealTimeRule.Load(ruleKey); ok {
		return r.(*rtask.RuleTask), true
	}
	if r, ok := m.DelayTimeRule.Load(ruleKey); ok {
		return r.(*rtask.RuleTask), true
	}
	if r, ok := m.VirtualRule.Load(ruleKey); ok {
		return r.(*rtask.RuleTask), true
	}
	return nil, false
}

func (m *RuleManager) deleteByRuleKey(ruleKey string) {
	delTask, exist := m.getRuleByRuleKey(ruleKey)
	if !exist {
		return
	}
	if delTask.AnalyzeTimeType == taskcode.RuleTaskRealtime {
		m.RtPointRuleMap.PRDelete(delTask)
		m.RealTimeRule.RDelete(ruleKey)
	} else if delTask.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
		m.DtPointRuleMap.PRDelete(delTask)
		m.DelayTimeRule.RDelete(ruleKey)
	} else if delTask.AnalyzeTimeType == taskcode.RuleTaskVirtual {
		m.VtPointRuleMap.PRDelete(delTask)
		m.VirtualRule.RDelete(ruleKey)
	}
}

func (m *RuleManager) addRuleTask(task *rtask.RuleTask) {
	// 如果有老版本的任务，则先删除，避免测点列表版本干扰
	delTask, exist := m.getRuleByRuleKey(task.GetKey())
	if exist {
		m.deleteByRuleKey(delTask.GetKey())
	}
	// 更新任务控制器
	// 先更新策略实例map，再更新测点列表
	if task.AnalyzeTimeType == taskcode.RuleTaskRealtime {
		m.RealTimeRule.Store(task.GetKey(), task)
		m.RtPointRuleMap.PRStore(task)
	} else if task.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
		m.DelayTimeRule.Store(task.GetKey(), task)
		m.DtPointRuleMap.PRStore(task)
	} else if task.AnalyzeTimeType == taskcode.RuleTaskVirtual {
		m.VirtualRule.Store(task.GetKey(), task)
		m.VtPointRuleMap.PRStore(task)
	}
}

// DelRuleTaskByKey 通过传入的key列表在任务管理器中删除任务
// @param keyList 元素形如： rid;gid;version
func (m *RuleManager) DelRuleTaskByKey(keyList []string) error {
	if len(keyList) == 0 {
		return nil
	}
	// 解析失败的key列表
	errFormatKeyList := []string{}
	// 多余删除的key列表（要删除的key在原任务管理器中不存在） / version不同
	invalidKeyList := []string{}
	for _, key := range keyList {
		eles := strings.Split(key, ";")
		if len(eles) != 3 {
			errFormatKeyList = append(errFormatKeyList, key)
			continue
		}
		ridStr, gidStr, versionStr := eles[0], eles[1], eles[2]
		ridNum, err := strconv.Atoi(ridStr)
		if err != nil {
			errFormatKeyList = append(errFormatKeyList, key)
			continue
		}
		ruleKey := fmt.Sprintf(rtask.RuleTaskKeyTemplate, ridNum, gidStr)
		delRT, exist := m.getRuleByRuleKey(ruleKey)
		if !exist || delRT.Version != versionStr {
			invalidKeyList = append(invalidKeyList, key)
			continue
		}
		if delRT.AnalyzeTimeType == taskcode.RuleTaskRealtime {
			m.RtPointRuleMap.PRDelete(delRT)
			m.RealTimeRule.RDelete(ruleKey)
		} else if delRT.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
			m.DtPointRuleMap.PRDelete(delRT)
			m.DelayTimeRule.RDelete(ruleKey)
		} else if delRT.AnalyzeTimeType == taskcode.RuleTaskVirtual {
			m.VtPointRuleMap.PRDelete(delRT)
			m.VirtualRule.RDelete(ruleKey)
		}
	}
	errMsg := ""
	if len(errFormatKeyList) > 0 {
		errMsg = fmt.Sprintf("format error key: %v", errFormatKeyList)
	}
	if len(invalidKeyList) > 0 {
		errMsg = fmt.Sprintf("%s, invalid key: %v", errMsg, invalidKeyList)
	}
	if len(errMsg) > 0 {
		return fmt.Errorf("DelRuleTaskByKey wrong: %s", errMsg)
	}
	return nil
}

// AddRuleTasks 更新策略执行任务
func (m *RuleManager) AddRuleTasks(alarmConfigs []*entity.AlarmConfig, updateType pb.ReqStrategyRecv_PublishType) error {
	errParseConfigKey := []string{}
	// 存量告警策略任务Key集合，用于在全量更新时，删除多余的rulekey
	delRuleKeySet := m.getTotalRuleKeySet()
	for _, config := range alarmConfigs {
		newRule, err := TransferConfig2Rules(config, config.Gid)
		if err != nil {
			log.Errorf("parse config error, rid:%d,gid:%s,version:%s,err: %s",
				config.Rid, config.Gid, config.RidVersion, err.Error())
			errParseConfigKey = append(errParseConfigKey, fmt.Sprintf(
				ErrParseConfigKeyTemplate, config.Rid, config.Gid, config.RidVersion))
			continue
		}
		// 增量更新
		if updateType == pb.ReqStrategyRecv_INCREMENT {
			m.addRuleTask(newRule)
		} else if updateType == pb.ReqStrategyRecv_UPDATEALL {
			// 全量更新
			// 当前为全量任务，最后需要将多余的删除
			m.addRuleTask(newRule)
			delete(delRuleKeySet, newRule.GetKey())
		} else {
			return fmt.Errorf("invalid update type: %v", updateType)
		}
	}
	// 全量更新时，删除原有的多余策略
	// 增量更新不用删除
	if updateType == pb.ReqStrategyRecv_UPDATEALL {
		for ruleKey := range delRuleKeySet {
			m.deleteByRuleKey(ruleKey)
		}
	}
	if len(errParseConfigKey) > 0 {
		return fmt.Errorf("AddRuleTasks parse configs error, key: %v", errParseConfigKey)
	}
	return nil
}

// TransferConfig2Rules transfer config to ruleTask
func TransferConfig2Rules(alarmConfig *entity.AlarmConfig, gid string) (*rtask.RuleTask, error) {

	rule, err := InitRuleTask(gid, alarmConfig)
	if err != nil {
		log.Errorf("init ruleTask error, gid: %v, rid: %v, error: %v", gid, alarmConfig.Rid, err)
		return nil, err
	}
	return rule, nil
}

// InitRuleTask init ruleTask
func InitRuleTask(gid string, config *entity.AlarmConfig) (*rtask.RuleTask, error) {
	var rt = &rtask.RuleTask{
		Version:         config.RidVersion,
		Rid:             config.Rid,
		MozuId:          config.MozuId,
		Gid:             gid,
		Level:           config.AlarmLevel,
		Alert:           rtask.NewAlarmTask(rtask.AlarmTaskType),
		Restore:         rtask.NewAlarmTask(rtask.RestoreTaskType),
		AlarmName:       config.AlarmName,
		ContentTemplate: config.ContentTemplate,
	}
	rt.SetKey()
	err := rt.SetAnalyzeTimeType(config.RidType)
	if err != nil {
		return nil, err
	}
	rt.Alert.RuleTask = rt
	rt.Restore.RuleTask = rt
	// 初始化告警任务和恢复任务 exp
	err = rt.SetAlarmTaskExp(config)
	if err != nil {
		return nil, err
	}
	if rt.AnalyzeTimeType == taskcode.RuleTaskDelaytime {
		err := rt.Alert.UpdatePointFetchList()
		if err != nil {
			log.Errorf("failed parse alarm delay expression, gid:%s, rid:%v", gid, rt.Rid, err)
			return nil, err
		}
		err = rt.Restore.UpdatePointFetchList()
		if err != nil {
			// 恢复表达式计算出错，打印错误日志，不中断告警计算流程
			log.Errorf("failed parse restore delay expression, gid:%s, rid:%v", gid, rt.Rid, err)
		}
	} else if rt.AnalyzeTimeType == taskcode.RuleTaskVirtual {
		err := rt.Alert.UpdatePointFetchList()
		if err != nil {
			log.Errorf("failed parse alarm virtual expression, gid:%s, rid:%v", gid, rt.Rid, err)
			return nil, err
		}
	}
	rt.SetPointDelayMap()
	return rt, nil
}
