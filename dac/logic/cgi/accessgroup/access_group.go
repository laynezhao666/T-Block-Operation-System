// Package accessgroup 提供门禁权限组的增删改查业务逻辑。
package accessgroup

import (
	"context"
	"fmt"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/card"
	"dac/repo/dac"
)

// GetAllCardGroups 获取模组下所有卡权限组
func GetAllCardGroups(ctx context.Context, mozuID string) ([]db.AccessGroup, error) {
	return dac.GetRW().GetAllCardAccessGroups(ctx, mozuID)
}

// Get 分页获取权限组及其关联的门、卡、时间组信息
func Get(ctx context.Context, mozuID string, offset int, limit int) (cgi.AccessGroupAndRelationInfos, error) {
	var result cgi.AccessGroupAndRelationInfos
	var err error

	total, accessGroups, err := dac.GetRW().GetAccessGroups(ctx, mozuID, offset, limit)
	if err != nil {
		return result, fmt.Errorf("GetAccessGroups error, %w", err)
	}
	accessGroupIDs := make([]db.IDType, len(accessGroups))
	timeGroupNos := make([]db.IDType, len(accessGroups))
	for i := range accessGroups {
		accessGroupIDs[i] = accessGroups[i].ID
		timeGroupNos[i] = accessGroups[i].TimeGroupNo
	}

	// 获取权限组相关的门和卡信息, 并发查询
	var doors map[db.IDType][]db.Door
	var cardAndStaffs map[db.IDType][]db.CardAndStaffBase
	var timeGroups map[int]db.TimeGroup
	err = utils.AsyncCall(
		func() error {
			doors, err = dac.GetRW().GetAccessGroupDoors(ctx, accessGroupIDs)
			return err
		},
		func() error {
			cardAndStaffs, err = dac.GetRW().GetCardAndStaffsByAccessGroupIDs(ctx, accessGroupIDs, mozuID)
			return err
		},
		func() error {
			timeGroups, err = dac.GetRW().GetTimeGroupsByNos(ctx, timeGroupNos)
			return err
		},
	)
	if err != nil {
		return result, fmt.Errorf("get doors or cards or timegroups error, %w", err)
	}

	accessInfos := make([]cgi.AccessGroupAndRelationInfo, len(accessGroups))
	for i := range accessGroups {
		info := &accessInfos[i]
		a := &accessGroups[i]
		info.AccessGroup = *a
		info.Door = utils.GetDoorsBaseInfo(doors[a.ID])
		info.Card = cardAndStaffs[a.ID]
		info.TimeGroup.GroupNo = a.TimeGroupNo
		info.TimeGroup.GroupName = timeGroups[a.TimeGroupNo].GroupName
	}

	result = cgi.AccessGroupAndRelationInfos{
		Total: total,
		List:  accessInfos,
	}

	return result, nil
}

// Add 新增权限组，按控制器合并门信息
func Add(ctx context.Context, mozuID string, wrapper db.AccessGroupInfoWrapper) (db.IDType, error) {
	return card.MergeByDoorsInController(ctx, mozuID, wrapper)
}

// Update 更新权限组信息
func Update(ctx context.Context, id db.IDType, mozuID string, wrapper db.AccessGroupInfoWrapper) error {
	return card.UpdateByAccessGroups(ctx, id, mozuID, wrapper)
}

// Delete 删除权限组并清理控制器中的关联数据
func Delete(ctx context.Context, id db.IDType, mozuID string) error {
	return card.PruneAccessGroupInController(ctx, id, mozuID)
}
