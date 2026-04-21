package dac

import (
	"context"
	"errors"
	"fmt"

	"dac/entity/model/db"
	"dac/entity/utils"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// createDoors 批量创建门记录。
func createDoors(tx *gorm.DB, doors []db.Door) error {
	if len(doors) == 0 {
		return nil
	}

	return tx.Create(&doors).Error
}

// getAllDoors 获取所有门记录。
func getAllDoors(tx *gorm.DB) ([]db.Door, error) {
	doors := make([]db.Door, 0)
	err := tx.Find(&doors).Error
	return doors, err
}

// GetDoor 根据ID获取单个门记录。
func GetDoor(tx *gorm.DB, id db.IDType) (db.Door, error) {
	var door db.Door
	err := queryRecordByID(tx, id, &door)
	return door, err
}

// GetDoors 根据ID列表获取门记录。
func GetDoors(tx *gorm.DB, ids []db.IDType) ([]db.Door, error) {
	if len(ids) == 0 {
		return make([]db.Door, 0), nil
	}

	var doors []db.Door
	err := queryRecordsByIDs(tx, ids, &doors)
	return doors, err
}

// getDoorsMap 根据ID列表获取门记录并转换为以门ID为key的map。
func getDoorsMap(tx *gorm.DB, ids []db.IDType) (map[db.IDType]db.Door, error) {
	doors, err := GetDoors(tx, ids)
	if err != nil {
		return nil, err
	}

	return utils.GetDoorsMap(doors), nil
}

// GetControllerDoors 获取指定控制器下的所有门（支持后置回调）。
func (d *impl) GetControllerDoors(ctx context.Context, controllerID db.IDType,
	afterGet func(*gorm.DB, []db.Door) error) ([]db.Door, error) {
	var (
		err   error
		doors = make([]db.Door, 0)
	)
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		results, e := getDoorsWithControllerID(tx, []db.IDType{controllerID})
		if e != nil {
			return e
		}
		doors = results[controllerID]

		if afterGet != nil {
			if e = afterGet(tx, doors); e != nil {
				return e
			}
		}

		return nil
	})

	return doors, err
}

// getDoorsWithControllerID 根据控制器ID列表获取门记录（按控制器ID分组）。
func getDoorsWithControllerID(tx *gorm.DB, controllerIDs []db.IDType) (map[db.IDType][]db.Door, error) {
	if len(controllerIDs) == 0 {
		return make(map[db.IDType][]db.Door), nil
	}

	var doors []db.Door
	err := withControllerIDs(tx, controllerIDs).Find(&doors).Error
	if err != nil {
		return nil, err
	}

	results := make(map[db.IDType][]db.Door, len(controllerIDs))
	for i := range doors {
		d := &doors[i]
		results[d.ControllerID] = append(results[d.ControllerID], *d)
	}
	return results, err
}

// setDoorsGroup 批量设置门的分组ID。
func setDoorsGroup(tx *gorm.DB, ids []db.IDType, groupID db.IDType) error {
	if len(ids) == 0 {
		return nil
	}

	return withIDs(tx.Model(&db.Door{}), ids).Updates(map[string]interface{}{
		"group_id": groupID,
	}).Error
}

// deleteDoors 批量删除门及其权限组关联关系。
func deleteDoors(tx *gorm.DB, ids []db.IDType) error {
	if len(ids) == 0 {
		return nil
	}

	var err error
	if err = deleteRecordsByID(tx, ids, &db.Door{}); err != nil {
		return err
	}
	if err = withIn(tx, "door_id", ids).Delete(&db.AccessGroupRelation{}).Error; err != nil {
		return err
	}
	return nil
}

// deleteDoorsWithControllerID 根据控制器ID列表删除关联的门（存在权限组关联时拒绝删除）。
func deleteDoorsWithControllerID(tx *gorm.DB, controllerIDs []db.IDType) error {
	if len(controllerIDs) == 0 {
		return nil
	}

	var doors []db.Door

	err := withControllerIDs(tx.Select("id"), controllerIDs).Find(&doors).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if len(doors) == 0 {
		return nil
	}

	doorIDs := make([]db.IDType, 0, len(doors))
	for i := range doors {
		doorIDs = append(doorIDs, doors[i].ID)
	}
	accessGroupCount := int64(0)
	withIn(tx.Model(&db.AccessGroupRelation{}), "door_id", doorIDs).Count(&accessGroupCount)
	if accessGroupCount > 0 {
		return fmt.Errorf("删除门 %+v 失败: 存在已关联的权限组", doorIDs)
	}

	return tx.Delete(&doors).Error
}

// GetGroupDoors 获取指定分组下的所有门。
func (d *impl) GetGroupDoors(ctx context.Context, group db.IDType) ([]db.Door, error) {
	doors := make([]db.Door, 0)
	err := withEqual(d.db.WithContext(ctx), "group_id", group).Find(&doors).Error
	return doors, err
}

// GetDoor 根据ID获取单个门（impl 方法）。
func (d *impl) GetDoor(ctx context.Context, id db.IDType) (db.Door, error) {
	return GetDoor(d.db.WithContext(ctx), id)
}

// GetDoors 根据ID列表获取门（impl 方法）。
func (d *impl) GetDoors(ctx context.Context, ids []db.IDType) ([]db.Door, error) {
	return GetDoors(d.db.WithContext(ctx), ids)
}

// GetAllDoors 获取所有门（impl 方法）。
func (d *impl) GetAllDoors(ctx context.Context) ([]db.Door, error) {
	return getAllDoors(d.db.WithContext(ctx))
}

// SetDoors 设置控制器下的门列表（自动处理新增和删除多余的门，冲突时更新参数）。
func SetDoors(tx *gorm.DB, controllerID db.IDType, doors []db.Door, afterSave func(*gorm.DB) error) error {
	if len(doors) == 0 {
		return nil
	}

	var err error
	if existedDoors, err := getDoorsWithControllerID(tx, []db.IDType{controllerID}); err == nil {
		// 删除编号不存在的门
		oldDoors, ok := existedDoors[controllerID]
		if ok && len(oldDoors) > 0 {
			newDoorNumbers := make(map[int]struct{}, len(doors))
			for i := range doors {
				newDoorNumbers[doors[i].Number] = struct{}{}
			}

			toDeleteDoors := make([]db.IDType, 0)
			for i := range oldDoors {
				if _, ok = newDoorNumbers[oldDoors[i].Number]; !ok {
					toDeleteDoors = append(toDeleteDoors, oldDoors[i].ID)
				}
			}
			if err = deleteDoors(tx, toDeleteDoors); err != nil {
				return err
			}
		}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: db.ColumnNumber}, {Name: db.ColumnControllerID}},
		// 其它列不需要更新
		DoUpdates: clause.AssignmentColumns([]string{db.ColumnParameters}),
	}).Create(&doors).Error
	if err != nil {
		return err
	}

	if afterSave != nil {
		if err = afterSave(tx); err != nil {
			return err
		}
	}
	return nil
}

// SetDoors 设置控制器下的门列表（impl 方法）。
func (d *impl) SetDoors(ctx context.Context, controllerID db.IDType,
	doors []db.Door, afterSave func(*gorm.DB) error,
) error {
	return SetDoors(d.db.WithContext(ctx), controllerID, doors, afterSave)
}

// UpdateDoorCode 更新门的采集编码（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateDoorCode(ctx context.Context, id db.IDType, code string,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error,
) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		if err = UpdateDoorCode(tx, id, code); err != nil {
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

// UpdateDoorCode 更新门的采集编码。
func UpdateDoorCode(tx *gorm.DB, id db.IDType, code string) error {
	return withID(tx.Model(&db.Door{}), id).Updates(map[string]interface{}{
		db.ColumnCode: code,
	}).Error
}

// updateDoorName 更新门的名称。
func updateDoorName(tx *gorm.DB, id db.IDType, name string) error {
	return withID(tx.Model(&db.Door{}), id).Updates(map[string]interface{}{
		"name": name,
	}).Error
}

// updateDoor 更新门信息（同时处理 IDC 编码）。
func updateDoor(tx *gorm.DB, id db.IDType, door *db.Door) error {
	var err error
	if err = tgorm.WithOptions(tx.Model(&db.Door{}), withIDOption(id)).Updates(*door).Error; err != nil {
		return err
	}
	code := door.GetIDCDBCode()
	if len(code) == 0 {
		if err = UpdateDoorCode(tx, id, code); err != nil {
			return err
		}

	}
	return nil
}

// updateDoorExtend 更新门的扩展信息。
func updateDoorExtend(tx *gorm.DB, id db.IDType, extend map[string]interface{}) error {
	if len(extend) == 0 {
		return nil
	}
	return withID(tx.Model(&db.Door{}), id).Updates(map[string]interface{}{
		"extend": extend,
	}).Error
}

// updateDoorNames 批量更新门的名称。
func updateDoorNames(tx *gorm.DB, ids []db.IDType, names map[db.IDType]string) error {
	if len(names) == 0 {
		return nil

	}

	var err error
	for _, id := range ids {
		name := names[id]
		if len(name) == 0 {
			continue
		}

		if err = updateDoorName(tx, id, name); err != nil {
			return err
		}
	}

	return nil
}

// updateDoorsParams 批量更新门的参数。
func updateDoorsParams(tx *gorm.DB, ids []db.IDType, params map[db.IDType]*db.DoorParameter) error {
	var err error
	for _, id := range ids {
		p, ok := params[id]
		if !ok || p == nil {
			continue
		}
		// struct 更新不能使用 map[string]interface{}
		if err = withID(tx.Model(&db.Door{}), id).Updates(db.Door{Parameters: *p}).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateDoorsNameAndParams 批量更新门的名称和参数（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateDoorsNameAndParams(ctx context.Context, ids []db.IDType,
	names map[db.IDType]string, params map[db.IDType]*db.DoorParameter,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error {
	if len(ids) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error

		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		if err = updateDoorNames(tx, ids, names); err != nil {
			return err
		}

		if err = updateDoorsParams(tx, ids, params); err != nil {
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

// UpdateDoor 更新单个门信息（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateDoor(ctx context.Context, id db.IDType, door *db.Door,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error

		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		if err = updateDoor(tx, id, door); err != nil {
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

// getCardNosFromAccessGroupRelation 获取卡号，同时去掉重复的
func getCardNosFromAccessGroupRelation(caRelation []db.CardAccessRelation) []string {
	var cardNos []string
	cardNoMap := make(map[string]bool)
	for i := range caRelation {
		r := &caRelation[i]
		_, exist := cardNoMap[r.CardNo]
		if !exist {
			cardNos = append(cardNos, r.CardNo)
		}
	}
	return cardNos
}

// UpdateDoors 批量更新门信息（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateDoors(ctx context.Context, doors map[db.IDType]*db.Door,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error,
) error {
	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		for id, door := range doors {
			if err = updateDoor(tx, id, door); err != nil {
				return err
			}
		}

		if afterUpdate != nil {
			if err = afterUpdate(tx); err != nil {
				return err
			}
		}

		return nil
	})
}
