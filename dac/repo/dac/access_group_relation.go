package dac

import (
	"context"
	"fmt"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/utils"

	"gorm.io/gorm"
)

// deleteAccessGroupRelation 删除指定权限组的门关联关系。
func deleteAccessGroupRelation(tx *gorm.DB, id db.IDType) error {
	return withAccessGroupID(tx, id).Delete(&db.AccessGroupRelation{}).Error
}

// getAccessGroupRelationDoorIDs 从权限组关联关系中提取去重后的门ID列表。
func getAccessGroupRelationDoorIDs(relations []db.AccessGroupRelation) []db.IDType {
	doorIDMap := make(map[db.IDType]struct{})
	for i := range relations {
		doorIDMap[relations[i].DoorID] = struct{}{}
	}

	doorIDs := make([]db.IDType, len(doorIDMap))
	for id := range doorIDMap {
		doorIDs = append(doorIDs, id)
	}
	return doorIDs
}

// GetAccessGroupIDsByDoors 根据门ID列表获取关联的权限组ID列表。
func GetAccessGroupIDsByDoors(tx *gorm.DB, doorIDs []db.IDType) ([]db.IDType, error) {
	if len(doorIDs) == 0 {
		return make([]db.IDType, 0), nil
	}

	var relations []db.AccessGroupRelation
	if err := withDoorIDs(tx, doorIDs).Find(&relations).Error; err != nil {
		return nil, err
	}

	accessGroupIDMap := make(map[db.IDType]struct{}, len(relations))
	for i := range relations {
		accessGroupID := relations[i].AccessGroupID
		accessGroupIDMap[accessGroupID] = struct{}{}
	}
	return utils.IntMapToSlice(accessGroupIDMap), nil
}

// getGroupRelationByGroups 根据权限组ID列表获取门关联关系。
func getGroupRelationByGroups(tx *gorm.DB, accessGroupIDs []db.IDType) ([]db.AccessGroupRelation, error) {
	if len(accessGroupIDs) == 0 {
		return make([]db.AccessGroupRelation, 0), nil
	}
	var relations []db.AccessGroupRelation
	err := withAccessGroupIDs(tx, accessGroupIDs).Find(&relations).Error
	return relations, err
}

// updateAccessGroupDoorRelation 更新权限组与门的关联关系（先删后增）。
func updateAccessGroupDoorRelation(tx *gorm.DB, accessGroupID db.IDType, doorIDs []db.IDType) error {
	// 删除原有记录
	if err := deleteAccessGroupRelation(tx, accessGroupID); err != nil {
		return err
	}

	if len(doorIDs) == 0 { // 当门信息为空时，只做删除
		return nil
	}

	// 新增
	accessGroupRelations := make([]db.AccessGroupRelation, len(doorIDs))
	for i := range accessGroupRelations {
		accessGroupRelations[i] = db.AccessGroupRelation{
			AccessGroupID: accessGroupID,
			DoorID:        doorIDs[i],
		}
	}
	return tx.Create(&accessGroupRelations).Error
}

// GetAccessGroupsDoors 根据权限组ID列表获取每个权限组关联的门列表。
func GetAccessGroupsDoors(tx *gorm.DB, ids []db.IDType) (map[db.IDType][]db.Door, error) {
	results := make(map[db.IDType][]db.Door)
	if len(ids) == 0 {
		return results, nil
	}

	var (
		relations []db.AccessGroupRelation
		err       error
	)
	if relations, err = getGroupRelationByGroups(tx, ids); err != nil {
		return nil, fmt.Errorf("get access group relations error: %w", err)
	}

	doorIDs := getAccessGroupRelationDoorIDs(relations)
	doorMap, err := getDoorsMap(tx, doorIDs)
	if err != nil {
		return nil, fmt.Errorf("get doors error: %w", err)
	}

	for i := range relations {
		groupID := relations[i].AccessGroupID
		doorID := relations[i].DoorID

		d, ok := doorMap[doorID]
		if !ok {
			config.Log.Warnf("not find door %v", doorID)
			continue
		}

		results[groupID] = append(results[groupID], d)
	}

	return results, nil
}

// GetAccessGroupDoors 获取权限组关联的门列表（impl 方法）。
func (d *impl) GetAccessGroupDoors(ctx context.Context, ids []db.IDType) (map[db.IDType][]db.Door, error) {
	return GetAccessGroupsDoors(d.db.WithContext(ctx), ids)
}
