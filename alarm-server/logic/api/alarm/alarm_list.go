package alarm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"etrpc-go/log"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"

	"alarm-server/logic/cache"
	"alarm-server/repo/dao/alarm"
	cmodel "common/entity/model"

	pb "trpcprotocol/alarm-server"
)

const (
	// 本地动环 按告警设备 & 告警实例 类型筛选
	AlarmDeviceTemplate = "%s %s"
)

// AnalyzeRet 告警计算结果，用于接口展示
type AnalyzeRet struct {
	PointValueMap        map[string]float64         `json:"pointValueMap,omitempty"`
	HistoryPointValueMap map[string]map[int]float64 `json:"historyPointValueMap,omitempty"`
	PointMap             map[string][]string        `json:"pointMap,omitempty"`
	StartRunAt           int64                      `json:"startRunAt,omitempty"`
}

/*
modifyMetricsMap 数量统计
维度：告警设备 告警实例
*/
func modifyMetricsMap(mozuId int64, metricsMap map[string]*pb.RspAlarmList_VList) {
	// 告警设备
	alarmGidMetricList := metricsMap["device_gid"].List
	newList := []*pb.RspAlarmList_VItem{}
	for _, item := range alarmGidMetricList {
		entity, ok := cache.GetLocalCache().GetDeviceCache(item.Name)
		if ok {
			item.Name = fmt.Sprintf(AlarmDeviceTemplate, entity.DeviceTypeZh, entity.DeviceNumber)
		}
		newList = append(newList, item)
	}
	metricsMap["device_gid"] = &pb.RspAlarmList_VList{
		List: newList,
	}

	strategyMap, ok := cache.GetLocalCache().GetStrategyCache(mozuId)
	if !ok {
		log.Errorf("GetStrategyCache failed when modifyMetricsMap, mozuId: %d", mozuId)
		return
	}
	// 告警实例 fingerprint -> rid;gid
	alarmInstanceMetricList := metricsMap["fingerprint"].List
	newList = []*pb.RspAlarmList_VItem{}
	for _, item := range alarmInstanceMetricList {
		parts := strings.Split(item.Name, ";")
		if len(parts) < 2 {
			continue
		}
		ridStr, gid := parts[0], parts[1]
		rid, err := strconv.ParseInt(ridStr, 10, 64)
		if err != nil {
			continue
		}
		if _, ok := strategyMap[rid]; !ok {
			continue
		}
		strategyItem, ok := strategyMap[rid].Get(gid)
		if !ok {
			continue
		}
		deviceNumber := ""
		deviceEntity, ok := cache.GetLocalCache().GetDeviceCache(strategyItem.DeviceGid)
		if ok {
			deviceNumber = deviceEntity.DeviceNumber
		}
		item.Name = fmt.Sprintf(AlarmDeviceTemplate, strategyItem.AlarmName, deviceNumber)
		newList = append(newList, item)
	}
	metricsMap["fingerprint"] = &pb.RspAlarmList_VList{
		List: newList,
	}
}

/*
getActiveAlarm 获取活动告警
*/
func getActiveAlarm(req *pb.ReqAlarmList) ([]*cmodel.AlarmActive, int64, map[string]*pb.RspAlarmList_VList, error) {
	// 构建查询条件
	con := &alarm.ActiveAlarmFilter{
		MozuId:        req.GetMozuId(),
		Level:         req.GetLevel(),
		OccurBegin:    req.GetOccurBegin(),
		OccurEnd:      req.GetOccurEnd(),
		AlarmName:     req.GetAlarmName(),
		AlarmId:       req.GetAlarmId(),
		Rid:           req.GetRid(),
		Content:       req.GetContent(),
		DeviceGid:     req.GetDeviceGid(),
		DeviceNumber:  req.GetDeviceNumber(),
		SortType:      req.GetSortType(),
		CountByMetric: req.GetCountByMetric(),
		Page:          req.GetPage(),
		Size:          req.GetSize(),
	}
	// eventStatus 1:未转单 2: 已转单 3: 已结单
	switch req.GetEventStatus() {
	case 1:
		con.EventStatus = []int64{1}
	case 2:
		con.EventStatus = []int64{2}
	case 3:
		con.EventStatus = []int64{3}
	default:
		con.EventStatus = []int64{}
	}
	// status 0:正常 1:挂起
	switch req.GetAlarmType() {
	case 1:
		con.Status = []int64{0}
	case 2:
		con.Status = []int64{1}
	default:
		con.Status = []int64{}
	}
	// 获取活动告警
	retList, cnt, metricsMap, err := alarm.NewAlarmDao().GetActiveAlarmList(trpc.BackgroundContext(), con)
	if req.GetCountByMetric() && len(retList) > 0 {
		// 对告警 分类统计
		modifyMetricsMap(req.GetMozuId(), metricsMap)
	}
	return retList, cnt, metricsMap, err
}

// getHistoryAlarm 获取历史告警
func getHistoryAlarm(req *pb.ReqAlarmList) ([]*cmodel.AlarmHistory, int64, map[string]*pb.RspAlarmList_VList, error) {
	// 构建查询条件
	con := &alarm.HistoryAlarmFilter{
		MozuId:        req.GetMozuId(),
		AlarmId:       req.GetAlarmId(),
		OccurBegin:    req.GetOccurBegin(),
		OccurEnd:      req.GetOccurEnd(),
		DeviceGid:     req.GetDeviceGid(),
		DeviceNumber:  req.GetDeviceNumber(),
		Level:         req.GetLevel(),
		AlarmName:     req.GetAlarmName(),
		Content:       req.GetContent(),
		Rid:           req.GetRid(),
		RestoreBegin:  req.GetRestoreBegin(),
		RestoreEnd:    req.GetRestoreEnd(),
		MaxDuration:   req.GetMaxDuration(),
		MinDuration:   req.GetMinDuration(),
		SortType:      req.GetSortType(),
		CountByMetric: req.GetCountByMetric(),
		Page:          req.GetPage(),
		Size:          req.GetSize(),
	}
	retList, cnt, metricsMap, err := alarm.NewAlarmDao().GetHistoryAlarmList(trpc.BackgroundContext(), con)
	if req.GetCountByMetric() && len(retList) > 0 {
		// 对告警 分类统计
		modifyMetricsMap(req.GetMozuId(), metricsMap)
	}
	return retList, cnt, metricsMap, err
}

// getPointsViaAnalyzeResult 解析AnalyzeResult，获取测点列表
func getPointsViaAnalyzeResult(analyzeResult string) []string {
	if len(analyzeResult) == 0 {
		return []string{}
	}
	analyzeRet := AnalyzeRet{}
	err := json.Unmarshal([]byte(analyzeResult), &analyzeRet)
	if err != nil {
		log.Errorf("Unmarshal analyzeResult err:%s", err.Error())
		return []string{}
	}
	retList := []string{}
	// 获取测点列表
	for _, pointName := range analyzeRet.PointMap {
		retList = append(retList, pointName...)
	}
	return retList
}

// GetAlarmList 查询告警列表
func (a *alarmLogicImpl) GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error) {
	alarmType := req.GetAlarmType()
	rsp := &pb.RspAlarmList{
		List: []*pb.RspAlarmList_Item{},
	}
	var getMozuNameFun = func(gid string) string {
		entity, ok := cache.GetLocalCache().GetDeviceCache(gid)
		if !ok {
			return ""
		}
		return entity.MozuName
	}
	if alarmType == 3 {
		retList, cnt, metricsMap, err := getHistoryAlarm(req)
		if err != nil {
			log.Errorf("getHistoryAlarm err:%s", err.Error())
			return nil, err
		}
		for _, item := range retList {
			// 历史告警Item构建返回Item
			rsp.List = append(rsp.List, &pb.RspAlarmList_Item{
				AlarmId:   item.AlarmID,
				Level:     item.Level,
				AlarmName: item.AlarmName,
				Box:       item.BoxName,
				Room:      item.RoomName,
				Rid:       item.Rid,
				MozuId:    int64(item.MozuId),
				// 每个设备，查询模组Name
				MozuName:     getMozuNameFun(item.DeviceGid),
				AlarmContent: item.Content,
				OccurTime:    item.OccurTime.Format(time.DateTime),
				RestoreTime:  item.RestoreTime.Format(time.DateTime),
				RestoreType:  "",
				// 此处的Points，解析AnalyzeResult构建
				Points:       getPointsViaAnalyzeResult(item.AnalyzeResult),
				DeviceGid:    item.DeviceGid,
				DeviceNumber: item.DeviceNumber,
				DeviceTypeZh: item.DeviceTypeZh,
			})
		}
		// 构建返回的告警总数
		rsp.Total = cnt
		rsp.Metrics = metricsMap
		return rsp, nil
	} else {
		retList, cnt, metricsMap, err := getActiveAlarm(req)
		if err != nil {
			log.Errorf("getActiveAlarm err:%s", err.Error())
			return nil, err
		}
		for _, item := range retList {
			// 活动告警Item构建返回Item —— 挂起/非挂起
			rsp.List = append(rsp.List, &pb.RspAlarmList_Item{
				AlarmId:      item.AlarmID,
				Level:        item.Level,
				AlarmName:    item.AlarmName,
				DeviceGid:    item.DeviceGid,
				DeviceNumber: item.DeviceNumber,
				DeviceTypeZh: item.DeviceTypeZh,
				Box:          item.BoxName,
				Room:         item.RoomName,
				Rid:          item.Rid,
				MozuId:       int64(item.MozuId),
				MozuName:     getMozuNameFun(item.DeviceGid),
				AlarmContent: item.Content,
				AlarmStatus:  int64(item.Status),
				EventStatus:  int64(item.EventStatus),
				OccurTime:    item.OccurTime.Format(time.DateTime),
				HangupReason: item.OpReason,
				Points:       getPointsViaAnalyzeResult(item.AnalyzeResult),
			})
		}
		// 构建返回的告警总数
		rsp.Total = cnt
		rsp.Metrics = metricsMap
		return rsp, nil
	}
}

// DelHistoryAlarm 删除历史告警
func (a *alarmLogicImpl) DelHistoryAlarm(ctx context.Context, req *pb.ReqDelHistoryAlarm) (*emptypb.Empty, error) {
	// 删除历史告警条件
	con := &alarm.DelHistoryAlarmCon{
		MozuId:    req.GetMozuId(),
		EndTime:   time.Unix(req.GetEndTime(), 10).Format(time.DateTime),
		Rid:       req.GetRid(),
		DeviceGid: req.GetDeviceGid(),
		Level:     req.GetLevel(),
	}
	// 删除历史告警
	err := alarm.NewAlarmDao().DelHistoryAlarm(ctx, con)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
