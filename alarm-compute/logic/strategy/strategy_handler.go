package strategy

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"

	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/conf"
	"alarm-compute/entity"
	"alarm-compute/logic/lcache"
	"alarm-compute/logic/rules/rmanager"
	"alarm-compute/repo"
	"alarm-compute/utils/common"
)

var (
	strategyHandler *StrategyHandler
	once            sync.Once
)

// StrategyHandler 策略处理器
type StrategyHandler struct {
	StrategyCh chan *pb.ReqStrategyRecv
}

// GetStrategyHandler 获取策略处理器
func GetStrategyHandler() *StrategyHandler {
	once.Do(func() {
		strategyHandler = &StrategyHandler{
			StrategyCh: make(chan *pb.ReqStrategyRecv, 10),
		}
	})
	return strategyHandler
}

// Run Run
func (s *StrategyHandler) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	go s.RegularSyncAlarm2Cache(trpc.BackgroundContext())
	for {
		select {
		case req := <-s.StrategyCh:
			s.HandleStrategyReq(req)
		case <-ctx.Done():
			return
		}
	}
}

// AddStrategyReq 添加策略请求
func (s *StrategyHandler) AddStrategyReq(req *pb.ReqStrategyRecv) bool {
	select {
	case s.StrategyCh <- req:
		return true
	default:
		log.Warnf("strategy channel is full")
		return false
	}
}

// HandleStrategyReq handle strategy request
func (s *StrategyHandler) HandleStrategyReq(req *pb.ReqStrategyRecv) {
	log.Infof("strategy req, timeStamp:%s, pushType:%d, add len:%d, del len:%d",
		req.RecvTimestamp, req.PublishType, len(req.AddTask), len(req.DelTaskKey))
	delKeys := req.DelTaskKey
	// 删除策略任务
	err := rmanager.GetGlobalRuleManager().DelRuleTaskByKey(delKeys)
	if err != nil {
		log.Errorf("err occurred when del task, req timeStamp:%d, err:%s", req.RecvTimestamp, err.Error())
	}
	addTask := req.AddTask
	alarmConfigList := []*entity.AlarmConfig{}
	for _, item := range addTask {
		alarmConfig, err := s.parseStrategyPb2Alarm(item)
		if err != nil {
			continue
		}
		alarmConfigList = append(alarmConfigList, alarmConfig)
	}
	if len(alarmConfigList) != 0 {
		err = rmanager.GetGlobalRuleManager().AddRuleTasks(alarmConfigList, req.PublishType)
		if err != nil {
			log.Errorf("err occurred when add task, req timeStamp:%d, err:%s", req.RecvTimestamp, err.Error())
		}
	}
	log.Infof("realtimeCount: %d, delaytimeCount: %d, virtualCount: %d",
		len(rmanager.GetGlobalRuleManager().GetRealtimeRules()),
		len(rmanager.GetGlobalRuleManager().GetDelaytimeRules()),
		len(rmanager.GetGlobalRuleManager().GetVirtualRules()))
	// 策略变化，数据库中同步活动告警
	go s.readActiveAlarmFromDB()
}

func (s *StrategyHandler) parseStrategyPb2Alarm(item *pb.ReqStrategyRecv_AddItem) (*entity.AlarmConfig, error) {
	var expressionsMap entity.ExpressionsMap
	err := json.Unmarshal([]byte(item.ExpressionMap), &expressionsMap)
	if err != nil {
		log.Errorf("Error parsing JSON, rid:%d, gid: %s, version:%d, mozuId:%d, err:%s",
			item.Rid, item.Gid, item.Version, item.MozuId, err)
		return nil, err
	}
	alarmConfig := &entity.AlarmConfig{
		Rid:               item.Rid,
		Gid:               item.Gid,
		RidVersion:        item.Version,
		RidType:           item.RidType,
		MozuId:            item.MozuId,
		AlarmExpression:   item.AlarmExpression,
		RestoreExpression: item.RestoreExpression,
		ExpressionMap:     expressionsMap,
		AlarmLevel:        item.AlarmLevel,
		AlarmName:         item.AlarmName,
		ContentTemplate:   item.ContentTemplate,
	}
	return alarmConfig, nil
}

// RegularSyncAlarm2Cache ...
// 1. 定期更新本地缓存
func (s *StrategyHandler) RegularSyncAlarm2Cache(ctx context.Context) {
	interval := time.Duration(conf.ServerConf.ActiveAlarmCache.ActiveNormalSyncInterval) * time.Second
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.readActiveAlarmFromDB()
		}
	}
}

func (s *StrategyHandler) readActiveAlarmFromDB() {
	rule_key_list := []string{}
	for _, rule := range rmanager.GetGlobalRuleManager().GetRealtimeRules() {
		rule_key_list = append(rule_key_list, rule.GetKey())
	}
	for _, rule := range rmanager.GetGlobalRuleManager().GetDelaytimeRules() {
		rule_key_list = append(rule_key_list, rule.GetKey())
	}
	if len(rule_key_list) == 0 {
		return
	}
	chunkList, err := common.ChunkStringList(rule_key_list, int(conf.ServerConf.ActiveAlarmCache.ActiveRequestBatchSize))
	if err != nil {
		log.Errorf("ActiveCache ChunkList failed: %v", err)
		return
	}
	for _, fingerList := range chunkList {
		active_list, err := repo.GetActiveFingerprints(fingerList)
		if err != nil {
			log.Errorf("GetActiveFingerprints failed: %v", err)
			return
		}
		for _, active_item := range active_list {
			// 将活动告警写入缓存
			lcache.GetLocalCache().SetActiveAlarmCache(active_item, time.Now().Unix(),
				int64(conf.ServerConf.ActiveAlarmCache.CacheKeyTimeDuration))
		}
	}
}
