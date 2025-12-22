// Package api 告警相关接口实现逻辑
package api

import (
	"context"

	"etrpc-go/util/httputil"
	alarmPb "trpcprotocol/alarm-server"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	pb "trpcprotocol/cgi"
)

const (
	// AlarmNameCacheKey 告警名称缓存key
	AlarmNameCacheKey = "alarm:name"
)

// IAlarmApi 告警相关逻辑接口
type IAlarmApi interface {
	GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error)
	GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error)
	GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error)
	GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error)
	GetStrategy(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error)
	GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error)
	GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error)
	GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error)
}

// alarmApi 告警相关逻辑接口实现类
type alarmApi struct {
	alarmClientProxy alarmPb.AlarmServerClientProxy
}

// NewAlarmApi 创建Alarm相关逻辑接口实现类
func NewAlarmApi() IAlarmApi {
	return &alarmApi{
		alarmClientProxy: alarmPb.NewAlarmServerClientProxy(),
	}
}

/*
GetAlarmCnt 获取告警数量
@param mozuId 模组ID
@param begin 开始时间
@param end 结束时间
@param alarmType 告警类型
@param eventStatus 建单/转单状态
@param level 告警等级
@return 告警数量
*/
func (a *alarmApi) GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error) {
	alarmReq := &alarmPb.ReqAlarmCnt{
		MozuId:      int64(req.MozuId),
		Begin:       int64(req.Begin),
		End:         int64(req.End),
		AlarmType:   int64(req.AlarmType),
		EventStatus: int64(req.EventStatus),
		Level:       req.Level,
	}
	alarmRsp, err := a.alarmClientProxy.GetAlarmCnt(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm cnt api fail")
	}
	rsp := &pb.RspAlarmCnt{
		Begin: alarmRsp.Begin,
		End:   alarmRsp.End,
		Count: alarmRsp.Count,
	}
	return rsp, nil
}

/*
GetAlarmCntTrend 获取告警数量趋势
@param mozuId 模组ID
@return 告警数量趋势
*/
func (a *alarmApi) GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error) {
	alarmReq := &alarmPb.ReqAlarmCntTrend{
		MozuId: int64(req.MozuId),
	}
	alarmRsp, err := a.alarmClientProxy.GetAlarmCntTrend(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm cnt trend api fail")
	}
	rsp := &pb.RspAlarmCntTrend{}
	rsp.List = lo.Map(alarmRsp.List, func(item *alarmPb.RspAlarmCntTrend_AlarmCount, index int) *pb.RspAlarmCntTrend_AlarmCount {
		return &pb.RspAlarmCntTrend_AlarmCount{
			UTime: item.UTime,
			Count: item.Count,
		}
	})
	return rsp, nil
}

/*
GetAlarmName 获取告警类型名称
@param alarmType 告警类型
@param page 分页
@param size 分页大小
@return 告警类型名称列表
*/
func (a *alarmApi) GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error) {
	alarmReq := &alarmPb.ReqAlarmName{
		AlarmType: int64(req.AlarmType),
		Page:      int64(req.Page),
		Size:      int64(req.Size),
	}
	alarmRsp, err := a.alarmClientProxy.GetAlarmName(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm name api fail")
	}
	rsp := &pb.RspAlarmName{
		Total: int32(alarmRsp.Total),
		List:  alarmRsp.List,
	}
	return rsp, nil
}

/*
geneMetricMap 将告警类型列表转换为pb格式
*/
func (a *alarmApi) geneMetricMap(metrics map[string]*alarmPb.RspAlarmList_VList) map[string]*pb.RspAlarmList_VList {
	metricsMap := map[string]*pb.RspAlarmList_VList{}
	for metric, vList := range metrics {
		if metric != "level" {
			metricsMap[metric] = &pb.RspAlarmList_VList{
				List: lo.Map(vList.List, func(item *alarmPb.RspAlarmList_VItem,
					index int) *pb.RspAlarmList_VItem {
					return &pb.RspAlarmList_VItem{
						Name:  item.Name,
						Count: int32(item.Count),
					}
				}),
			}
		} else {
			metricsMap[metric] = &pb.RspAlarmList_VList{}
			levelMap := map[string]int64{}
			for _, item := range vList.List {
				levelMap[item.Name] = item.Count
			}
			for _, level := range []string{"L0", "L1", "L2", "L3", "L4"} {
				_, ok := levelMap[level]
				metricsMap[metric].List = append(metricsMap[metric].List, &pb.RspAlarmList_VItem{
					Name:  level,
					Count: lo.Ternary(ok, int32(levelMap[level]), 0),
				})
			}
		}
	}
	return metricsMap
}

/*
GetAlarmList 获取告警列表
1. 返回活动告警/历史告警列表
2. 告警数量统计
*/
func (a *alarmApi) GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error) {
	alarmReq := &alarmPb.ReqAlarmList{
		MozuId:        int64(req.MozuId),
		AlarmType:     int64(req.AlarmType),
		AlarmId:       req.AlarmId,
		Rid:           int64(req.Rid),
		OccurBegin:    req.OccurBegin,
		OccurEnd:      req.OccurEnd,
		Level:         req.Level,
		EventStatus:   int64(req.EventStatus),
		DeviceGid:     req.DeviceGid,
		DeviceNumber:  req.DeviceNumber,
		AlarmName:     req.AlarmName,
		Content:       req.Content,
		RestoreBegin:  req.RestoreBegin,
		RestoreEnd:    req.RestoreEnd,
		MaxDuration:   int64(req.MaxDuration),
		MinDuration:   int64(req.MinDuration),
		SortType:      int64(req.SortType),
		CountByMetric: req.CountByMetric,
		Page:          int64(req.Page),
		Size:          int64(req.Size),
	}
	alarmRsp, err := a.alarmClientProxy.GetAlarmList(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm list api fail")
	}
	rsp := &pb.RspAlarmList{
		List: lo.Map(alarmRsp.List, func(item *alarmPb.RspAlarmList_Item, index int) *pb.RspAlarmList_Item {
			return &pb.RspAlarmList_Item{
				AlarmId:      item.AlarmId,
				Level:        item.Level,
				AlarmName:    item.AlarmName,
				Rid:          int32(item.Rid),
				DeviceGid:    item.DeviceGid,
				DeviceNumber: item.DeviceNumber,
				DeviceTypeZh: item.DeviceTypeZh,
				Box:          item.Box,
				Room:         item.Room,
				MozuId:       int32(item.MozuId),
				MozuName:     item.MozuName,
				AlarmContent: item.AlarmContent,
				AlarmStatus:  int32(item.AlarmStatus),
				EventStatus:  int32(item.EventStatus),
				Points:       item.Points,
				OccurTime:    item.OccurTime,
				RestoreTime:  item.RestoreTime,
				RestoreType:  item.RestoreType,
				HangupReason: item.HangupReason,
			}
		}),
		Total: int32(alarmRsp.Total),
	}
	rsp.Metrics = a.geneMetricMap(alarmRsp.Metrics)
	return rsp, nil
}

/*
GetStrategy 获取告警策略
按照rid聚合，一个策略有多个device
*/
func (a *alarmApi) GetStrategy(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error) {
	alarmReq := &alarmPb.ReqStrategyList{
		MozuId:       int64(req.MozuId),
		Rid:          int64(req.Rid),
		DeviceGid:    req.DeviceGid,
		DeviceNumber: req.DeviceNumber,
		DeviceType:   req.DeviceType,
		ApplyType:    req.ApplyType,
		AlarmName:    req.AlarmName,
		Level:        req.Level,
		Page:         int64(req.Page),
		Size:         int64(req.Size),
	}
	alarmRsp, err := a.alarmClientProxy.GetStrategy(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm strategy api fail")
	}
	rsp := &pb.RspStrategyList{
		Total: int32(alarmRsp.Total),
		List: lo.Map(alarmRsp.List, func(item *alarmPb.RspStrategyList_Item,
			index int) *pb.RspStrategyList_Item {
			return &pb.RspStrategyList_Item{
				Rid:        int32(item.Rid),
				ApplyType:  item.ApplyType,
				DeviceType: item.DeviceType,
				AlarmName:  item.AlarmName,
				Level:      item.Level,
				Content:    item.Content,
				Standard:   item.Standard,
				AlarmExp:   item.AlarmExp,
				RestoreExp: item.RestoreExp,
				Owner:      item.Owner,
				CreateAt:   item.CreateAt,
				UpdateAt:   item.UpdateAt,
				DeviceList: item.DeviceList,
				GidList:    item.GidList,
			}
		}),
	}
	return rsp, nil
}

/*
GetStrategyInstance 获取告警策略实例
rid - device_gid
*/
func (a *alarmApi) GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error) {
	alarmReq := &alarmPb.ReqStrategyInstance{
		MozuId:       int64(req.MozuId),
		Rid:          int64(req.Rid),
		DeviceGid:    req.DeviceGid,
		DeviceNumber: req.DeviceNumber,
		DeviceType:   req.DeviceType,
		ApplyType:    req.ApplyType,
		AlarmName:    req.AlarmName,
		Level:        req.Level,
		Page:         int64(req.Page),
		Size:         int64(req.Size),
	}
	alarmRsp, err := a.alarmClientProxy.GetStrategyInstance(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm strategy instance api fail")
	}
	rsp := &pb.RspStrategyInstance{
		Total: int32(alarmRsp.Total),
		List: lo.Map(alarmRsp.List, func(item *alarmPb.RspStrategyInstance_Item,
			index int) *pb.RspStrategyInstance_Item {
			return &pb.RspStrategyInstance_Item{
				MozuId:        int32(item.MozuId),
				Rid:           int32(item.Rid),
				Version:       item.Version,
				RidType:       int32(item.RidType),
				ApplyType:     item.ApplyType,
				AlarmName:     item.AlarmName,
				Level:         item.Level,
				Content:       item.Content,
				AlarmExp:      item.AlarmExp,
				RestoreExp:    item.RestoreExp,
				ExpressionMap: item.ExpressionMap,
				Owner:         item.Owner,
				CreateAt:      item.CreateAt,
				UpdateAt:      item.UpdateAt,
				DeviceNumber:  item.DeviceNumber,
				DeviceGid:     item.DeviceGid,
				DeviceType:    item.DeviceType,
				Points:        item.Points,
			}
		}),
	}
	return rsp, nil
}

/*
GetValidate 获取告警策略生效详情
策略实例生效数量 / 策略实例总数
*/
func (a *alarmApi) GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error) {
	alarmReq := &alarmPb.ReqValidateList{
		MozuId:       int64(req.MozuId),
		ValidType:    int64(req.ValidType),
		Begin:        req.Begin,
		End:          req.End,
		RuleType:     int64(req.RuleType),
		Level:        req.Level,
		DeviceGid:    req.DeviceGid,
		DeviceNumber: req.DeviceNumber,
		AlarmName:    req.AlarmName,
		ErrorName:    req.ErrorName,
		Page:         int64(req.Page),
		Size:         int64(req.Size),
	}
	alarmRsp, err := a.alarmClientProxy.GetValidate(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm validate api fail")
	}
	rsp := &pb.RspValidateList{
		Metrics:   alarmRsp.Metrics,
		MetricMap: alarmRsp.MetricMap,
		Total:     int32(alarmRsp.Total),
		List: lo.Map(alarmRsp.List, func(item *alarmPb.RspValidateList_Item,
			index int) *pb.RspValidateList_Item {
			return &pb.RspValidateList_Item{
				Rid:          int32(item.Rid),
				DeviceGid:    item.DeviceGid,
				DeviceNumber: item.DeviceNumber,
				Level:        item.Level,
				AlarmName:    item.AlarmName,
				Content:      item.Content,
				AlarmExp:     item.AlarmExp,
				RestoreExp:   item.RestoreExp,
				Points:       item.Points,
				Standard:     item.Standard,
				EvalTime:     item.EvalTime,
				Succeed:      item.Succeed,
				Fired:        item.Fired,
				ErrorCode:    int32(item.ErrorCode),
				ErrorName:    item.ErrorName,
				ErrorDetail:  item.ErrorDetail,
			}
		}),
	}
	return rsp, nil
}

/*
GetVirtualPoint 获取虚拟测点
特殊场景下，存在告警虚拟点
*/
func (a *alarmApi) GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error) {
	alarmReq := &alarmPb.ReqGetVirtualPoint{
		MozuId:    int64(req.MozuId),
		DeviceGid: req.DeviceGid,
		PointId:   req.PointId,
	}
	alarmRsp, err := a.alarmClientProxy.GetVirtualPoint(ctx, alarmReq, httputil.GetPbCallOption())
	if err != nil {
		return nil, errors.Wrapf(err, "request alarm get virtual point api fail")
	}
	rsp := &pb.RspGetVirtualPoint{
		List: lo.Map(alarmRsp.List, func(item *alarmPb.RspGetVirtualPoint_Item,
			index int) *pb.RspGetVirtualPoint_Item {
			return &pb.RspGetVirtualPoint_Item{
				MozuId:        int32(item.MozuId),
				PointName:     item.PointName,
				ComputeCost:   int32(item.ComputeCost),
				Expression:    item.Expression,
				ExpressionMap: item.ExpressionMap,
				CreateAt:      item.CreateAt,
				UpdateAt:      item.UpdateAt,
			}
		}),
	}
	return rsp, nil
}
