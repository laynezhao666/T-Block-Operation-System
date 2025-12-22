package strategy

import (
	"context"
	"fmt"
	"time"

	"etrpc-go/log"

	"github.com/samber/lo"

	"alarm-server/entity/model"
	"alarm-server/logic/cache"
	"alarm-server/repo/store"
	"alarm-server/utils/common"
	cmodel "common/entity/model"

	pb "trpcprotocol/alarm-server"

	"google.golang.org/protobuf/types/known/structpb"
)

// initMetricMap 初始化指标数据
// return metricMap[string][]int32 -> [execCnt, invalidCnt, fireCnt, unFireCnt]
func initMetricMap(cntMap map[string]int32) map[string][]int32 {
	return map[string][]int32{
		"all": {0, cntMap["all"], 0, 0},
		"L0":  {0, cntMap["L0"], 0, 0},
		"L1":  {0, cntMap["L1"], 0, 0},
		"L2":  {0, cntMap["L2"], 0, 0},
		"L3":  {0, cntMap["L3"], 0, 0},
		"L4":  {0, cntMap["L4"], 0, 0},
	}
}

// updateMetricMap 更新指标数据
// opType 操作类型 1: +1  -1: -1
func updateMetricMap(metricMap map[string][]int32, level string, index int32, opNum int32) {
	if _, ok := metricMap[level]; ok {
		metricMap[level][index] += opNum
	}
	metricMap["all"][index] += opNum
}

// FilterRuleRecord 过滤策略执行记录
// 0: 无效策略 1: 产生告警策略 2: 未产生告警策略 -1: 不返回策略列表
// @ param mozuId int64
// @param keyList []int64
// @param cntMap map[string]int32  当前模组下，策略总数
// @param validType int64 0: 无效策略 1: 产生告警策略 2: 未产生告警策略 -1: 返回全量
// @return recordMap [rid][gid]record, metricMap[string][]int32 -> [execCnt, invalidCnt, fireCnt, unFireCnt], err
// execCnt 执行策略总数
// invalidCnt 无效策略，包含： 未执行的策略，执行失败策略
func FilterRuleRecord(mozuId int64, keyList []int64,
	strategyCache map[int64]*common.OrderedMap[string, model.StrategyCacheData],
	cntMap map[string]int32, validType int64) (
	map[int64]map[string]*model.ValidStoreData, map[string][]int32, error) {
	chunkValuesMap, err := store.GetRedisStoreApi().BatchGetRuleRecord(mozuId, keyList, strategyCache)
	if err != nil || len(chunkValuesMap) <= 0 {
		log.Errorf("FilterRuleRecord BatchGetRuleRecord failed, mozuId:%d, err:%v", mozuId, err)
		return nil, nil, fmt.Errorf("FilterRuleRecord加载本地缓存失败, mozuId:%d, err:%v", mozuId, err)
	}
	metricMap := initMetricMap(cntMap)
	resultMap := map[int64]map[string]*model.ValidStoreData{}
	for rid, gidMap := range chunkValuesMap {
		if gidMap == nil {
			continue
		}
		resGidMap := make(map[string]*model.ValidStoreData)
		for gid, v := range gidMap {
			if v.ErrorCode != -1 {
				updateMetricMap(metricMap, v.AlarmLevel, 0, 1)
			}
			if v.Success {
				updateMetricMap(metricMap, v.AlarmLevel, 1, -1)
				if v.Fired {
					updateMetricMap(metricMap, v.AlarmLevel, 2, 1)
				} else {
					updateMetricMap(metricMap, v.AlarmLevel, 3, 1)
				}
			}
			if validType == -1 {
				resGidMap[gid] = v
			} else {
				if validType == 0 {
					if !v.Success {
						resGidMap[gid] = v
						continue
					}
				} else if validType == 1 {
					if v.Success && v.Fired {
						resGidMap[gid] = v
						continue
					}
				} else if validType == 2 {
					if v.Success && !v.Fired {
						resGidMap[gid] = v
						continue
					}
				}
			}
		}
		if len(resGidMap) > 0 {
			resultMap[rid] = resGidMap
		}
	}
	return resultMap, metricMap, nil
}

func geneValidListItem(item *model.ValidStoreData,
	strategyItem *model.StrategyCacheData, deviceEntity *cmodel.DeviceEntity) *pb.RspValidateList_Item {
	deviceNumber := ""
	if deviceEntity != nil {
		deviceNumber = deviceEntity.DeviceNumber
	}
	return &pb.RspValidateList_Item{
		Rid:          item.Rid,
		DeviceGid:    item.Gid,
		DeviceNumber: deviceNumber,
		Level:        strategyItem.AlarmLevel,
		AlarmName:    strategyItem.AlarmName,
		Content:      strategyItem.ContentTemplate,
		AlarmExp:     strategyItem.AlarmExpressionStr,
		RestoreExp:   strategyItem.RestoreExpressionStr,
		Points:       strategyItem.GetPointList(),
		Standard:     true,
		EvalTime:     time.Unix(item.EvalTime, 0).Format(time.DateTime),
		Succeed:      item.Success,
		Fired:        item.Fired,
		ErrorCode:    int64(item.ErrorCode),
		ErrorName:    item.ErrorName,
		ErrorDetail:  item.GetErrDetail(strategyItem),
	}
}

// cntMap map[level]cnt, metricMap[string][]int32 -> [execCnt, invalidCnt, fireCnt, unFireCnt], err
func geneRspMetricMap(cntMap map[string]int32, metricMap map[string][]int32) map[string]map[string]float64 {
	levelKeys := []string{"all", "L0", "L1", "L2", "L3", "L4"}
	ret := make(map[string]map[string]float64)
	for _, level := range levelKeys {
		if cntMap[level] == 0 {
			continue
		}
		lMap := map[string]float64{
			"all":        float64(cntMap[level]),
			"valid":      float64(int64(cntMap[level] - metricMap[level][1])),
			"invalid":    float64(metricMap[level][1]),
			"efficiency": float64(cntMap[level]-metricMap[level][1]) / float64(cntMap[level]),
			"fired":      float64(metricMap[level][2]),
			"unfired":    float64(metricMap[level][3]),
		}
		ret[level] = lMap
	}
	return ret
}

func convertRspMetricToStruct(metricMap map[string]map[string]float64) (map[string]*structpb.Struct, error) {
	resMap := make(map[string]*structpb.Struct)
	for level, vMap := range metricMap {
		valuesMap := make(map[string]any)
		for key, value := range vMap {
			valuesMap[key] = value
		}
		structMap, err := structpb.NewStruct(valuesMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert struct to structpb.Value: %v", err)
		}
		resMap[level] = structMap
	}
	return resMap, nil
}

// GetValidate 查询策略校验结果
func (s *strategyLogicImpl) GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	cacheMap, _ := cache.GetLocalCache().GetStrategyCache(req.MozuId)
	ruleList, _ := cache.GetLocalCache().GetStrategyKeyCache(req.MozuId)
	cntMap, ok := cache.GetLocalCache().GetMozuStrategyCnt(req.MozuId)
	if !ok || len(ruleList) == 0 {
		log.Errorf("GetValidate get strategy cache is empty, mozuId:%d", req.MozuId)
		return nil, fmt.Errorf("GetValidate get strategy cache is empty, mozuId:%d", req.MozuId)
	}
	runRecordMap, metricMap, err :=
		FilterRuleRecord(req.MozuId, ruleList, cacheMap, cntMap, req.ValidType)
	if err != nil {
		log.Errorf("GetValidate get rule run record failed, mozuId:%d, err:%v", req.MozuId, err)
		return nil, fmt.Errorf("GetValidate get rule run record failed, mozuId:%d, err:%v", req.MozuId, err)
	}
	metricsMap := geneRspMetricMap(cntMap, metricMap)
	retList := []*pb.RspValidateList_Item{}
	for rid, gidMap := range runRecordMap {
		for gid, item := range gidMap {
			strategyItem, ok := cacheMap[rid].Get(gid)
			if !ok {
				continue
			}
			deviceEntity, ok := cache.GetLocalCache().GetDeviceCache(gid)
			deviceNumber := ""
			if ok {
				deviceNumber = deviceEntity.DeviceNumber
			}
			if len(req.Begin) > 0 {
				t, err := time.ParseInLocation(time.DateTime, req.Begin, location)
				if err != nil || item.EvalTime < t.Unix() {
					continue
				}
			}
			if len(req.End) > 0 {
				t, err := time.ParseInLocation(time.DateTime, req.End, location)
				if err != nil || item.EvalTime > t.Unix() {
					continue
				}
			}
			if (len(req.Level) > 0 && !lo.Contains(req.Level, strategyItem.AlarmLevel)) ||
				(len(req.AlarmName) > 0 && !lo.Contains(req.AlarmName, strategyItem.AlarmName)) ||
				(len(req.DeviceGid) > 0 && !lo.Contains(req.DeviceGid, item.Gid)) ||
				(len(req.DeviceNumber) > 0 && !lo.Contains(req.DeviceNumber, deviceNumber)) ||
				(len(req.ErrorName) > 0 && !lo.Contains(req.ErrorName, item.ErrorName)) {
				continue
			}
			retList = append(retList, geneValidListItem(item, &strategyItem, deviceEntity))
		}
	}
	structMap, convertErr := convertRspMetricToStruct(metricsMap)
	if convertErr != nil {
		log.Errorf("convertRspMetricToStruct failed, err:%v", convertErr)
		structMap = map[string]*structpb.Struct{}
	}
	rsp := &pb.RspValidateList{
		Metrics:   metricsMap["all"],
		Total:     int64(len(retList)),
		MetricMap: structMap,
	}
	if req.Page <= 0 || req.Size <= 0 {
		rsp.List = retList
		return rsp, nil
	}
	start := (req.Page - 1) * req.Size
	end := req.Page * req.Size
	if end > int64(len(retList)) {
		end = int64(len(retList))
	}
	if start > end {
		rsp.List = []*pb.RspValidateList_Item{}
	} else {
		rsp.List = retList[start:end]
	}
	return rsp, nil
}
