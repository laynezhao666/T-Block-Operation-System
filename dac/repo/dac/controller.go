package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"
	"dac/entity/model/rt"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// GetAllDoorControllers 获取指定模组下所有门禁控制器（impl 方法）。
func (d *impl) GetAllDoorControllers(ctx context.Context, mozuID string) ([]db.DoorController, error) {
	return GetAllControllerRecords(d.db.WithContext(ctx), mozuID)
}

// GetControllerRecord 根据ID获取单个门禁控制器记录。
func GetControllerRecord(tx *gorm.DB, id db.IDType) (db.DoorController, error) {
	var controller db.DoorController
	err := queryRecordByID(tx, id, &controller)
	return controller, err
}

// GetDoorControllerRecord 根据ID获取单个门禁控制器记录（impl 方法）。
func (d *impl) GetDoorControllerRecord(ctx context.Context, id db.IDType) (db.DoorController, error) {
	return GetControllerRecord(d.db.WithContext(ctx), id)
}

// updateDoorController 更新门禁控制器信息。
func updateDoorController(tx *gorm.DB, id db.IDType, controller *db.DoorController) error {
	return withID(tx.Model(&db.DoorController{}), id).Updates(*controller).Error
}

// UpdateDoorController 更新门禁控制器信息（impl 方法，支持前置和后置回调）。
func (d *impl) UpdateDoorController(ctx context.Context, id db.IDType,
	controller *db.DoorController,
	beforeUpdate func(tx *gorm.DB) error,
	afterUpdate func(tx *gorm.DB) error,
) error {
	if controller == nil {
		return nil
	}

	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		controller.ID = id
		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		if err = updateDoorController(tx, id, controller); err != nil {
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

// AddDoorControllers 批量添加门禁控制器及其关联的门（impl 方法，支持前置和后置回调）。
func (d *impl) AddDoorControllers(ctx context.Context, controllers []rt.DoorController,
	beforeAdd func(*gorm.DB) error, afterAdd func(*gorm.DB) error,
) error {
	l := len(controllers)
	if l == 0 {
		return nil
	}

	doors := make([]db.Door, 0, l<<1)
	controllerRecords := make([]db.DoorController, 0, l)
	for i := range controllers {
		c := &controllers[i]

		doors = append(doors, c.Doors...)
		controllerRecords = append(controllerRecords, c.DoorController)
	}

	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if beforeAdd != nil {
			if err = beforeAdd(tx); err != nil {
				return err
			}
		}

		if err = createDoors(tx, doors); err != nil {
			return fmt.Errorf("create dac doors error: %w", err)
		}
		if err = createControllers(tx, controllerRecords); err != nil {
			return fmt.Errorf("create dac controllers error: %w", err)
		}

		if afterAdd != nil {
			if err = afterAdd(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetControllerNames 根据ID列表获取控制器名称映射。
func (d *impl) GetControllerNames(ctx context.Context, ids []db.IDType) (map[db.IDType]string, error) {
	controllerMap := make(map[db.IDType]string, len(ids))
	var err error
	var controllerRecords []db.DoorController
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		controllerRecords, err = getDoorControllers(tx, ids)
		if err != nil {
			return fmt.Errorf("get dac controllers error: %w", err)
		}

		for i := range controllerRecords {
			r := &controllerRecords[i]
			controllerMap[r.ID] = r.Name
		}
		return nil
	})

	return controllerMap, err
}

// GetDoorControllers 根据ID列表获取门禁控制器及其关联的门。
func (d *impl) GetDoorControllers(ctx context.Context, ids []db.IDType) (map[db.IDType]rt.DoorController, error) {
	controllers := make(map[db.IDType]rt.DoorController, len(ids))
	var controllerRecords []db.DoorController
	var err error
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		controllerRecords, err = getDoorControllers(tx, ids)
		if err != nil {
			return fmt.Errorf("get dac controllers error: %w", err)
		}

		doors, err := getDoorsWithControllerID(tx, ids)
		if err != nil {
			return fmt.Errorf("get dac doors error: %w", err)
		}

		for i := range controllerRecords {
			r := &controllerRecords[i]

			c := rt.DoorController{
				DoorController: *r,
				Doors:          doors[r.ID],
			}
			if c.Doors == nil {
				c.Doors = make([]db.Door, 0)
			}
			controllers[r.ID] = c
		}
		return nil
	})
	return controllers, err
}

// DeleteDoorControllers 批量删除门禁控制器及其关联的驱动数据、门和请求。
func DeleteDoorControllers(tx *gorm.DB, ids []db.IDType) error {
	if len(ids) == 0 {
		return nil
	}

	var err error

	var temp []db.DoorController
	if err = tgorm.WithOptions(tx, withIDsOption(ids)).Find(&temp).Error; err != nil {
		return fmt.Errorf("get dac controllers error: %w", err)
	}
	for i := range temp {
		t := temp[i]
		if err = deleteDriverCards(tx, t.ID, t.Channel.ID); err != nil {
			return fmt.Errorf("delete driver cards of controller %+v error: %w", t, err)
		}
		if err = deleteDriverTimeGroups(tx, t.ID, t.Channel.ID); err != nil {
			return fmt.Errorf("delete driver time groups of controller %+v error: %w", t, err)
		}
		if err = deleteDriverDoorParameter(tx, t.ID, t.Channel.ID); err != nil {
			return fmt.Errorf("delete driver door parameter of controller %+v error: %w", t, err)
		}
	}

	if err = deleteRecordsByID(tx, ids, &db.DoorController{}); err != nil {
		return fmt.Errorf("delete dac controllers error: %w", err)
	}
	if err = deleteDoorsWithControllerID(tx, ids); err != nil {
		return fmt.Errorf("delete dac doors error: %w", err)
	}
	if err = deleteRequests(tx, ids); err != nil {
		return fmt.Errorf("delete requests error: %w", err)
	}
	return nil
}

// DeleteDoorControllers 批量删除门禁控制器（impl 方法，在事务中执行）。
func (d *impl) DeleteDoorControllers(ctx context.Context, ids []db.IDType) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return DeleteDoorControllers(tx, ids)
	})
}

// GetAllControllerRecords 获取指定模组下所有门禁控制器记录。
func GetAllControllerRecords(tx *gorm.DB, mozuID string) ([]db.DoorController, error) {
	controllers := make([]db.DoorController, 0)
	opts := make([]tgorm.Option, 0, 1)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	err := tgorm.WithOptions(tx, opts...).Find(&controllers).Error
	return controllers, err
}

// GetAllDoors 获取所有门记录并按控制器ID分组。
func GetAllDoors(tx *gorm.DB) (map[db.IDType][]db.Door, error) {
	doors, err := getAllDoors(tx)
	if err != nil {
		return nil, err
	}

	doorMap := make(map[db.IDType][]db.Door, len(doors))
	for i := range doors {
		d := &doors[i]
		doorMap[d.ControllerID] = append(doorMap[d.ControllerID], *d)
	}

	return doorMap, err
}

// GetAllDoorControllersAndDoors 获取指定模组下所有门禁控制器和门信息（impl 方法）。
func (d *impl) GetAllDoorControllersAndDoors(ctx context.Context,
	mozuID string,
) ([]db.DoorController, map[db.IDType][]db.Door, error) {
	return GetAllDoorControllersAndDoors(d.db.WithContext(ctx), mozuID)
}

// GetAllDoorControllersAndDoors 获取目标模组下所有门禁控制器和所有的门信息
func GetAllDoorControllersAndDoors(tx *gorm.DB, mozuID string) ([]db.DoorController, map[db.IDType][]db.Door, error) {
	controllers, err := GetAllControllerRecords(tx, mozuID)
	if err != nil {
		return nil, nil, err
	}

	doors, err := GetAllDoors(tx)
	if err != nil {
		return nil, nil, err
	}

	return controllers, doors, err
}
