// Package strategy ...
package strategy

import (
	"context"
	"fmt"
	"time"

	"etrpc-go/log"

	"github.com/samber/lo"

	"alarm-server/entity/model"
	"alarm-server/logic/cache"
	"alarm-server/repo/dao/strategy"
	"alarm-server/utils/common"
	cmodel "common/entity/model"

	pb "trpcprotocol/alarm-server"
)

// GetAlarmName 查询告警名称
func (s *strategyLogicImpl) GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error) {
	var retList []string
	var cnt int64
	var err error
	retList, cnt, err = strategy.NewStrategyDao().GetAlarmName(ctx, int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}
	rsp := &pb.RspAlarmName{}
	rsp.List = retList
	rsp.Total = cnt
	return rsp, nil
}

// GetStrategyList 查询策略列表
func (s *strategyLogicImpl) GetStrategyList(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error) {
	cacheMap, ok := cache.GetLocalCache().GetStrategyCache(req.MozuId)
	ridList, _ := cache.GetLocalCache().GetStrategyKeyCache(req.MozuId)
	if !ok || len(ridList) == 0 {
		log.Errorf("GetStrategyList cacheMap failed, mozuId:%d", req.MozuId)
		return nil, fmt.Errorf("GetStrategyList cacheMap failed, mozuId:%d", req.MozuId)
	}
	totalResList := []*pb.RspStrategyList_Item{}
	rsp := &pb.RspStrategyList{}
	if req.Rid != 0 {
		gidMap, ok := cacheMap[req.Rid]
		if !ok {
			return rsp, nil
		}
		var strategyItem = s.geneIntegrateItem(gidMap, req)
		if strategyItem != nil && len(strategyItem.GidList) > 0 {
			totalResList = append(totalResList, strategyItem)
		}
	} else {
		lo.ForEach(ridList, func(rid int64, index int) {
			gidMap, ok := cacheMap[rid]
			if !ok {
				return
			}
			var strategyItem = s.geneIntegrateItem(gidMap, req)
			if strategyItem != nil && len(strategyItem.GidList) > 0 {
				totalResList = append(totalResList, strategyItem)
			}
		})
	}
	rsp.Total = int64(len(totalResList))
	if req.Page <= 0 || req.Size <= 0 {
		rsp.List = totalResList
	} else {
		start := (req.Page - 1) * req.Size
		if start < 0 {
			start = 0
		}
		end := req.Page * req.Size
		if end > int64(len(totalResList)) {
			end = int64(len(totalResList))
		}
		if start <= end {
			rsp.List = totalResList[start:end]
		}
	}
	return rsp, nil
}

// GetStrategyInstance 查询策略实例列表
func (s *strategyLogicImpl) GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error) {
	ridList, ok := cache.GetLocalCache().GetStrategyKeyCache(req.MozuId)
	cacheMap, _ := cache.GetLocalCache().GetStrategyCache(req.MozuId)
	if !ok || len(cacheMap) == 0 {
		log.Errorf("GetStrategyInstance cacheMap failed, mozuId:%d", req.MozuId)
		return nil, fmt.Errorf("GetStrategyInstance cacheMap failed, mozuId:%d", req.MozuId)
	}
	rsp := &pb.RspStrategyInstance{}
	totalResList := []*pb.RspStrategyInstance_Item{}
	if req.Rid != 0 {
		gidMap, ok := cacheMap[req.Rid]
		if !ok {
			return rsp, nil
		}
		totalResList = s.geneInstanceItemList(gidMap, req)
	} else {
		lo.ForEach(ridList, func(rid int64, index int) {
			gidMap, ok := cacheMap[rid]
			if !ok {
				return
			}
			totalResList = append(totalResList, s.geneInstanceItemList(gidMap, req)...)
		})
	}
	rsp.Total = int64(len(totalResList))
	if req.Page <= 0 || req.Size <= 0 {
		rsp.List = totalResList
		return rsp, nil
	}
	start := (req.Page - 1) * req.Size
	if start < 0 {
		start = 0
	}
	end := req.Page * req.Size
	if end > int64(len(totalResList)) {
		end = int64(len(totalResList))
	}
	if start > end {
		return rsp, nil
	}
	rsp.List = totalResList[start:end]
	return rsp, nil
}

// 聚合策略类型查询
func (s *strategyLogicImpl) geneIntegrateItem(gidMap *common.OrderedMap[string, model.StrategyCacheData],
	req *pb.ReqStrategyList) *pb.RspStrategyList_Item {
	var strategyItem *pb.RspStrategyList_Item
	gidMap.Range(func(gid string, ruleItem model.StrategyCacheData) {
		if strategyItem == nil {
			if (len(req.AlarmName) > 0 && !lo.Contains(req.AlarmName, ruleItem.AlarmName)) ||
				(len(req.Level) > 0 && !lo.Contains(req.Level, ruleItem.AlarmLevel)) {
				return
			}
			strategyItem = &pb.RspStrategyList_Item{
				Rid:        ruleItem.Rid,
				AlarmName:  ruleItem.AlarmName,
				Level:      ruleItem.AlarmLevel,
				Content:    ruleItem.ContentTemplate,
				Standard:   true,
				AlarmExp:   ruleItem.AlarmExpressionStr,
				RestoreExp: ruleItem.RestoreExpressionStr,
				Owner:      ruleItem.Owner,
				CreateAt:   ruleItem.CreateAt.Format(time.DateTime),
				UpdateAt:   ruleItem.UpdateAt.Format(time.DateTime),
				GidList:    []string{},
				DeviceList: []string{},
			}
		}
		deviceGid := ruleItem.DeviceGid
		deviceEntity, ok := cache.GetLocalCache().GetDeviceCache(deviceGid)
		if !ok {
			deviceEntity = &cmodel.DeviceEntity{}
		}
		if (len(req.ApplyType) > 0 && !lo.Contains(req.ApplyType, deviceEntity.ApplicationTypeZh)) ||
			(len(req.DeviceGid) > 0 && !lo.Contains(req.DeviceGid, deviceGid)) ||
			(len(req.DeviceNumber) > 0 && !lo.Contains(req.DeviceNumber, deviceEntity.DeviceNumber)) ||
			(len(req.DeviceType) > 0 && !lo.Contains(req.DeviceType, deviceEntity.DeviceTypeZh)) {
			return
		}
		strategyItem.GidList = append(strategyItem.GidList, deviceGid)
		strategyItem.DeviceList = append(strategyItem.DeviceList, deviceEntity.DeviceNumber)
		if len(strategyItem.ApplyType) == 0 {
			strategyItem.ApplyType = deviceEntity.ApplicationTypeZh
		}
		if len(strategyItem.DeviceType) == 0 {
			strategyItem.DeviceType = deviceEntity.DeviceTypeZh
		}
	})
	if strategyItem != nil && len(strategyItem.GidList) > 0 {
		return strategyItem
	}
	return nil
}

// 单个策略实例查询
func (s *strategyLogicImpl) geneInstanceItemList(gidMap *common.OrderedMap[string, model.StrategyCacheData],
	req *pb.ReqStrategyInstance) []*pb.RspStrategyInstance_Item {
	totalResList := []*pb.RspStrategyInstance_Item{}
	gidMap.Range(func(gid string, ruleItem model.StrategyCacheData) {
		if (len(req.AlarmName) > 0 && !lo.Contains(req.AlarmName, ruleItem.AlarmName)) ||
			(len(req.Level) > 0 && !lo.Contains(req.Level, ruleItem.AlarmLevel)) {
			return
		}
		deviceGid := ruleItem.DeviceGid
		deviceEntity, ok := cache.GetLocalCache().GetDeviceCache(deviceGid)
		if !ok {
			deviceEntity = &cmodel.DeviceEntity{}
		}
		if (len(req.DeviceGid) > 0 && !lo.Contains(req.DeviceGid, deviceGid)) ||
			(len(req.DeviceNumber) > 0 && !lo.Contains(req.DeviceNumber, deviceEntity.DeviceNumber)) ||
			(len(req.DeviceType) > 0 && !lo.Contains(req.DeviceType, deviceEntity.DeviceTypeZh)) {
			return
		}
		instanceItem := &pb.RspStrategyInstance_Item{
			MozuId:        req.MozuId,
			Rid:           ruleItem.Rid,
			Version:       ruleItem.RidVersion,
			RidType:       int64(ruleItem.RidType),
			ApplyType:     deviceEntity.ApplicationTypeZh,
			AlarmName:     ruleItem.AlarmName,
			Level:         ruleItem.AlarmLevel,
			Content:       ruleItem.ContentTemplate,
			AlarmExp:      ruleItem.AlarmExpressionStr,
			RestoreExp:    ruleItem.RestoreExpressionStr,
			ExpressionMap: ruleItem.GetExprMapStr(),
			Owner:         ruleItem.Owner,
			CreateAt:      ruleItem.CreateAt.Format(time.DateTime),
			UpdateAt:      ruleItem.UpdateAt.Format(time.DateTime),
			DeviceNumber:  deviceEntity.DeviceNumber,
			DeviceGid:     deviceGid,
			DeviceType:    deviceEntity.DeviceTypeZh,
			Points:        ruleItem.GetPointList(),
			AlarmExpStd:   ruleItem.AlarmExpression,
			RestoreExpStd: ruleItem.RestoreExpression,
		}
		totalResList = append(totalResList, instanceItem)
	})
	return totalResList
}
