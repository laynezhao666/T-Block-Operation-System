package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// GetAccessGroups 分页获取权限组列表。
func (d *impl) GetAccessGroups(ctx context.Context, mozuID string,
	offset int, limit int,
) (int64, []db.AccessGroup, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("invalid offset: %v, limit: %v", offset, limit)
	}

	var (
		totalCount   int64
		accessGroups []db.AccessGroup
		opts         = make([]tgorm.Option, 0, 1)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)

	err := queryAndCountRecords(tgorm.WithOptions(d.db.WithContext(ctx).Model(&db.AccessGroup{}), opts...),
		offset, limit, &accessGroups, &totalCount)
	return totalCount, accessGroups, err
}

// GetAllAccessGroups 获取全量权限组数据
func (d *impl) GetAllAccessGroups(ctx context.Context, mozuID string) ([]db.AccessGroup, error) {
	var (
		accessGroups []db.AccessGroup
		opts         = make([]tgorm.Option, 0, 1)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)

	err := tgorm.WithOptions(d.db.WithContext(ctx).Model(&db.AccessGroup{}), opts...).
		Find(&accessGroups).Error

	return accessGroups, err
}

// AddAccessGroup 添加权限组及其关联的门和卡片（支持前置和后置回调）。
func (d *impl) AddAccessGroup(ctx context.Context, accessGroupWrapper db.AccessGroupInfoWrapper, mozuID string,
	beforeAdd func(*gorm.DB) error, afterAdd func(*gorm.DB) error) (db.IDType, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		if beforeAdd != nil {
			if err = beforeAdd(tx); err != nil {
				return err
			}
		}
		if err = tx.Create(&accessGroupWrapper.AccessGroup).Error; err != nil {
			return err
		}
		if err = updateGroupDoorAndCardRelation(tx, accessGroupWrapper.ID, accessGroupWrapper, mozuID); err != nil {
			return err
		}
		if afterAdd != nil {
			if err = afterAdd(tx); err != nil {
				return err
			}
		}
		return nil
	})

	return accessGroupWrapper.ID, err
}

// UpdateAccessGroup 更新权限组及其关联关系（支持前置和后置回调）。
func (d *impl) UpdateAccessGroup(ctx context.Context, id db.IDType, accessGroupWrapper db.AccessGroupInfoWrapper,
	mozuID string, beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}
		if err = withID(tx, id).Updates(&accessGroupWrapper.AccessGroup).Error; err != nil {
			return err
		}
		// 更新时间组序号，可能为 0
		if err = withID(tx.Model(&db.AccessGroup{}), id).Updates(map[string]interface{}{
			"time_group_no": accessGroupWrapper.TimeGroupNo,
		}).Error; err != nil {
			return err
		}
		if err = updateGroupDoorAndCardRelation(tx, id, accessGroupWrapper, mozuID); err != nil {
			return err
		}
		if afterUpdate != nil {
			if err = afterUpdate(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteAccessGroup 删除权限组及其关联的门和卡片关系（支持前置和后置回调）。
func (d *impl) DeleteAccessGroup(ctx context.Context, id db.IDType,
	beforeDelete func(*gorm.DB) error, afterDelete func(*gorm.DB) error) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if beforeDelete != nil {
			if err = beforeDelete(tx); err != nil {
				return err
			}
		}

		if err = withID(tx, id).Delete(&db.AccessGroup{}).Error; err != nil {
			return err
		}
		if err = deleteAccessGroupRelation(tx, id); err != nil {
			return err
		}
		if err = delCardAccessRelByGroup(tx, id); err != nil {
			return err
		}
		if afterDelete != nil {
			if err = afterDelete(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAccessGroupsByID 根据ID列表获取权限组列表。
func GetAccessGroupsByID(tx *gorm.DB, ids []db.IDType) ([]db.AccessGroup, error) {
	if len(ids) == 0 {
		return make([]db.AccessGroup, 0), nil
	}
	var groups []db.AccessGroup
	err := queryRecordsByIDs(tx, ids, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// GetAccessGroupMapByID 根据ID列表获取权限组映射（以ID为key）。
func GetAccessGroupMapByID(tx *gorm.DB, ids []db.IDType) (map[db.IDType]db.AccessGroup, error) {
	groups, err := GetAccessGroupsByID(tx, ids)
	if err != nil {
		return nil, err
	}

	results := make(map[db.IDType]db.AccessGroup)
	for i := range groups {
		results[groups[i].ID] = groups[i]
	}
	return results, nil
}

// GetAccessGroupsByID 根据ID列表获取权限组映射（impl 方法）。
func (d *impl) GetAccessGroupsByID(ctx context.Context, ids []db.IDType) (map[db.IDType]db.AccessGroup, error) {
	return GetAccessGroupMapByID(d.db.WithContext(ctx), ids)
}

// GetAccessGroupsBaseInfo 根据ID列表获取权限组基本信息映射。
func (d *impl) GetAccessGroupsBaseInfo(ctx context.Context,
	ids []db.IDType,
) (map[db.IDType]db.AccessGroupBaseInfo, error) {
	groups, err := getAccessGroupBaseInfo(d.db.WithContext(ctx), ids)
	if err != nil {
		return nil, err
	}

	results := make(map[db.IDType]db.AccessGroupBaseInfo, len(groups))
	for i := range groups {
		g := &groups[i]
		results[g.ID] = *g
	}
	return results, nil
}

// GetAllCardAccessGroups 获取指定模组下所有卡片关联的权限组。
func (d *impl) GetAllCardAccessGroups(ctx context.Context, mozuID string) ([]db.AccessGroup, error) {
	var groups []db.AccessGroup
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cards, e := GetAllCards(tx, mozuID, nil)
		if e != nil {
			return e
		}
		cardNumbers := make([]string, 0, len(cards))
		for i := range cards {
			cardNumbers = append(cardNumbers, cards[i].CardNo)
		}

		relations, e := GetCardAccessRelationByCards(tx, cardNumbers, mozuID)
		if e != nil {
			return e
		}

		groupIDs := getAccessGroupIDFromCardRelations(relations)
		if groups, e = GetAccessGroupsByID(tx, groupIDs); e != nil {
			return e
		}
		return nil
	})
	return groups, err
}
