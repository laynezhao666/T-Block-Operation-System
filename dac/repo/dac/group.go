package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// clearGroup 清除指定分组下所有门的分组ID（重置为默认分组）。
func clearGroup(tx *gorm.DB, id db.IDType) error {
	return withEqual(tx.Model(&db.Door{}), "group_id", id).Updates(map[string]interface{}{
		"group_id": db.DefaultGroupID,
	}).Error
}

// AddGroup 添加门分组并关联指定的门。
func (d *impl) AddGroup(ctx context.Context, group db.DoorGroup, doorIDs []db.IDType) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := tx.Create(&group)
		if err = r.Error; err != nil {
			return err
		}
		return setDoorsGroup(tx, doorIDs, group.ID)
	})
}

// DeleteGroup 删除门分组（先清除门的分组关联，再删除分组记录）。
func (d *impl) DeleteGroup(ctx context.Context, id db.IDType) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = clearGroup(tx, id); err != nil {
			return err
		}
		return tx.Delete(&db.DoorGroup{}, id).Error
	})
}

// UpdateGroup 更新门分组名称和关联的门列表。
func (d *impl) UpdateGroup(ctx context.Context, group db.DoorGroup, doorIDs []db.IDType) error {
	var err error
	id := group.ID
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = withID(tx.Model(&db.DoorGroup{}), id).Updates(map[string]interface{}{
			"name": group.Name,
		}).Error; err != nil {
			return fmt.Errorf("update group %v name to %v error: %w", id, group.Name, err)
		}

		if err = clearGroup(tx, id); err != nil {
			return fmt.Errorf("update group %v doors to default group error: %w", id, err)
		}

		if err = setDoorsGroup(tx, doorIDs, id); err != nil {
			return fmt.Errorf("update doors %v to group %v error: %w", doorIDs, id, err)
		}
		return nil
	})
}

// GetAllDoorGroups 获取指定模组下所有门分组。
func (d *impl) GetAllDoorGroups(ctx context.Context, mozuID string) ([]db.DoorGroup, error) {
	groups := make([]db.DoorGroup, 0)
	err := tgorm.WithOptions(d.db.WithContext(ctx), withMozuID(mozuID)).Find(&groups).Error
	return groups, err
}
