package dac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/collect/driver/xbrother/consts"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetDriverAllCards 获取指定控制器和通道下所有驱动卡片（按状态过滤）。
func (d *impl) GetDriverAllCards(ctx context.Context, controllerID db.IDType, channelID string, status []int) ([]db.DriverCard, error) {
	var driverCards = make([]db.DriverCard, 0)
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withStatuses(status)}
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).Find(&driverCards).Error
	return driverCards, err
}

// AddDriverCard 添加驱动卡片：若卡号已存在则更新，否则优先填充已删除的空位，最后才新建。
func (d *impl) AddDriverCard(ctx context.Context, controllerID db.IDType, channelID string, card driver.Card,
	addCardByNewCardFunc func(tx *gorm.DB, card driver.Card) error,
	addCardByUpdateOldCardFunc func(tx *gorm.DB, card driver.Card, oldCard db.DriverCard) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existDriverCard, err := getDriverCardByGormDB(tx, controllerID, channelID, card.CardNo)
		if err != nil {
			oldCard, err := GetFirstLogicDeleteDriverCard(tx, controllerID, channelID, consts.DriverCardStatusDelete)
			if err != nil {
				return addCardByNewCardFunc(tx, card)

			}
			// 协议规定，如果按顺序有删除的空位，优先填充
			return addCardByUpdateOldCardFunc(tx, card, oldCard)
		}
		return addCardByUpdateOldCardFunc(tx, card, existDriverCard)
	})
}

// AddDriverCard 创建或更新驱动卡片记录（冲突时全量更新）。
func AddDriverCard(tx *gorm.DB, driverCard db.DriverCard) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: db.ColumnControllerID}, {Name: db.ColumnChannelID}, {Name: db.ColumnCardNo}},
		UpdateAll: true,
	}).Create(&driverCard).Error
}

// GetDriverCard 获取指定控制器、通道和卡号的驱动卡片。
func (d *impl) GetDriverCard(ctx context.Context, controllerID db.IDType, channelID string, cardNo string) (db.DriverCard, error) {
	return getDriverCardByGormDB(d.db.WithContext(ctx), controllerID, channelID, cardNo)
}

// getDriverCardByGormDB 根据控制器ID、通道ID和卡号查询驱动卡片。
func getDriverCardByGormDB(tx *gorm.DB, controllerID db.IDType, channelID string, cardNo string) (db.DriverCard, error) {
	var driverCard db.DriverCard
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withCardNo(cardNo)}
	err := tgorm.WithOptions(tx, opts...).First(&driverCard).Error
	return driverCard, err
}

// LogicDeleteDriverCard 逻辑删除驱动卡片（将状态标记为已删除）。
func (d *impl) LogicDeleteDriverCard(ctx context.Context, controllerID db.IDType, channelID string, cardNo string, f func(driverCard db.DriverCard) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		driverCard, err := getDriverCardByGormDB(tx, controllerID, channelID, cardNo)
		if err != nil {
			return err
		}
		if f != nil {
			if err = f(driverCard); err != nil {
				return err
			}
		}
		opts := []tgorm.Option{
			withControllerIDOption(controllerID),
			withChannelID(channelID),
			withCardNo(cardNo),
		}
		return tgorm.WithOptions(
			tx.Model(&db.DriverCard{}), opts...,
		).Update(db.ColumnStatus, consts.DriverCardStatusDelete).Error
	})
}

// GetDriverCards 分页获取驱动卡片列表（按控制器、通道和状态过滤）。
func (d *impl) GetDriverCards(ctx context.Context, controllerID db.IDType, channelID string, offset int, limit int, status []int) (int64, []db.DriverCard, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("GetDriverCards not support by offset: %d, limit: %d", offset, limit)
	}

	var (
		totalCount  int64
		driverCards []db.DriverCard
	)
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withStatuses(status)}
	err := queryAndCountDriverCards(tgorm.WithOptions(d.db.WithContext(ctx).Model(&db.DriverCard{}), opts...),
		offset, limit, &driverCards, &totalCount)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, driverCards, nil
}

// AddDriverDoorParameter 添加单个驱动门参数（委托给批量方法）。
func (d *impl) AddDriverDoorParameter(ctx context.Context, controllerID db.IDType, channelID string,
	driverDoorParameter db.DriverDoorParameter) error {
	return d.AddDriverDoorParameters(ctx, controllerID, channelID, []db.DriverDoorParameter{driverDoorParameter})
}

// deleteDriverCards 删除指定控制器和通道下所有驱动卡片。
func deleteDriverCards(tx *gorm.DB, controllerID db.IDType, channelID string) error {
	if !tx.Migrator().HasTable(&db.DriverCard{}) {
		return nil
	}
	return tgorm.WithOptions(tx, withControllerIDOption(controllerID), withChannelID(channelID)).
		Delete(&db.DriverCard{}).Error
}

// deleteDriverTimeGroups 删除指定控制器和通道下所有驱动时间组。
func deleteDriverTimeGroups(tx *gorm.DB, controllerID db.IDType, channelID string) error {
	if !tx.Migrator().HasTable(&db.DriverTimeGroup{}) {
		return nil
	}
	return tgorm.WithOptions(tx, withControllerIDOption(controllerID), withChannelID(channelID)).
		Delete(&db.DriverTimeGroup{}).Error
}

// deleteDriverDoorParameter 删除指定控制器和通道下所有驱动门参数。
func deleteDriverDoorParameter(tx *gorm.DB, controllerID db.IDType, channelID string) error {
	if !tx.Migrator().HasTable(&db.DriverDoorParameter{}) {
		return nil
	}
	return tgorm.WithOptions(tx, withControllerIDOption(controllerID), withChannelID(channelID)).
		Delete(&db.DriverDoorParameter{}).Error
}

// SetDriverDoorParameters 先清除再重新设置驱动门参数。
func SetDriverDoorParameters(tx *gorm.DB, controllerID db.IDType, channelID string, driverDoorParameters []db.DriverDoorParameter) error {
	if len(driverDoorParameters) == 0 {
		return nil
	}
	var err error
	if err = deleteDriverDoorParameter(tx, controllerID, channelID); err != nil {
		return err
	}
	return AddDriverDoorParameters(tx, controllerID, channelID, driverDoorParameters)
}

// AddDriverDoorParameters 批量添加驱动门参数（冲突时全量更新）。
func AddDriverDoorParameters(tx *gorm.DB, controllerID db.IDType, channelID string, driverDoorParameters []db.DriverDoorParameter) error {
	if len(driverDoorParameters) == 0 {
		return nil
	}
	for i := range driverDoorParameters {
		driverDoorParameters[i].ControllerID = controllerID
		driverDoorParameters[i].ChannelID = channelID
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: db.ColumnControllerID}, {Name: db.ColumnChannelID}, {Name: db.ColumnNumber}},
		UpdateAll: true,
	}).Create(&driverDoorParameters).Error
}

// AddDriverDoorParameters 批量添加驱动门参数（impl 方法，委托给包级函数）。
func (d *impl) AddDriverDoorParameters(ctx context.Context, controllerID db.IDType, channelID string,
	driverDoorParameters []db.DriverDoorParameter) error {
	return AddDriverDoorParameters(d.db.WithContext(ctx), controllerID, channelID, driverDoorParameters)
}

// GetDriverDoorParameters 获取指定控制器和通道下所有驱动门参数。
func (d *impl) GetDriverDoorParameters(ctx context.Context, controllerID db.IDType, channelID string) ([]db.DriverDoorParameter, error) {
	driverDoorParams := make([]db.DriverDoorParameter, 0)
	err := tgorm.WithOptions(d.db.WithContext(ctx),
		withControllerIDOption(controllerID), withChannelID(channelID)).Find(&driverDoorParams).Error
	return driverDoorParams, err
}

// AddDriverTimeGroup 添加驱动时间组（冲突时全量更新）。
func (d *impl) AddDriverTimeGroup(ctx context.Context, driverTimeGroup db.DriverTimeGroup) error {
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: db.ColumnControllerID}, {Name: db.ColumnChannelID}, {Name: db.ColumnTimeGroupNo}},
		UpdateAll: true,
	}).Create(&driverTimeGroup).Error
}

// GetDriverTimeGroup 获取指定控制器、通道和编号的驱动时间组。
func (d *impl) GetDriverTimeGroup(ctx context.Context, controllerID db.IDType, channelID string, groupNo int) (
	db.DriverTimeGroup, error) {
	var driverTimeGroup db.DriverTimeGroup
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withGroupNo(groupNo)}
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).First(&driverTimeGroup).Error
	return driverTimeGroup, err
}

// ClearDriverTimeGroup 清除指定编号的驱动时间组（先回调处理其他编号的时间组，再删除目标编号）。
func (d *impl) ClearDriverTimeGroup(ctx context.Context, controllerID db.IDType, channelID string, timeGroupNo int, f func(dbTimeGroups []db.DriverTimeGroup) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbTimeGroups, err := GetDriverTimeGroupsExpectGroupNo(tx, controllerID, channelID, timeGroupNo)
		if err != nil {
			return err
		}
		if err = f(dbTimeGroups); err != nil {
			return err
		}
		return DeleteDriverTimeGroup(tx, controllerID, channelID, timeGroupNo)
	})
}

// GetDriverTimeGroupsExpectGroupNo 获取指定控制器和通道下除指定编号外的所有驱动时间组。
func GetDriverTimeGroupsExpectGroupNo(tx *gorm.DB, controllerID db.IDType, channelID string, groupNo int) ([]db.DriverTimeGroup, error) {
	var driverTimeGroup []db.DriverTimeGroup
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withoutGroupNo(groupNo)}
	err := tgorm.WithOptions(tx, opts...).Find(&driverTimeGroup).Error
	return driverTimeGroup, err
}

// DeleteDriverTimeGroup 删除指定编号的驱动时间组。
func DeleteDriverTimeGroup(tx *gorm.DB, controllerID db.IDType, channelID string, groupNo int) error {
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withGroupNo(groupNo)}
	return tgorm.WithOptions(tx, opts...).Delete(&db.DriverTimeGroup{}).Error
}

// SetDriverEvents 批量写入驱动事件（自动递增索引，在事务中执行）。
func (d *impl) SetDriverEvents(ctx context.Context, controllerID db.IDType, items []db.DriverEvent) error {
	if len(items) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var index int
		maxDriverEvent, err := GetMaxIndexDriverEvent(tx, controllerID, items[0].ChannelID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			index = 1
		} else {
			index = maxDriverEvent.Index + 1
		}
		for i := range items {
			items[i].Index = index
			index++
		}
		return tx.Create(&items).Error
	})

}

// SetDriverAlarms 批量写入驱动告警（自动递增索引，在事务中执行）。
func (d *impl) SetDriverAlarms(ctx context.Context, controllerID db.IDType, items []db.DriverAlarm) error {
	if len(items) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var index int
		maxDriverAlarm, err := GetMaxIndexDriverAlarm(tx, controllerID, items[0].ChannelID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			index = 1
		} else {
			index = maxDriverAlarm.Index + 1
		}
		for i := range items {
			items[i].Index = index
			index++
		}
		return tx.Create(&items).Error
	})
}

// GetDriverAlarms 分页获取驱动告警列表。
func (d *impl) GetDriverAlarms(ctx context.Context, controllerID db.IDType, channelID string,
	offset int, limit int) (int64, []db.DriverAlarm, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("GetDriverAlarms not support by offset: %d, limit: %d", offset, limit)
	}
	var (
		totalCount   int64
		driverAlarms []db.DriverAlarm
		opts         = make([]tgorm.Option, 0, 1)
	)
	opts = append(opts, withControllerIDOption(controllerID))
	opts = append(opts, withChannelID(channelID))
	err := queryAndCountDriverAlarms(tgorm.WithOptions(d.db.WithContext(ctx).Model(&db.DriverAlarm{}), opts...),
		offset, limit, &driverAlarms, &totalCount)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, driverAlarms, nil
}

// GetDriverEvents 分页获取驱动事件列表。
func (d *impl) GetDriverEvents(ctx context.Context, controllerID db.IDType, channelID string,
	offset int, limit int) (int64, []db.DriverEvent, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("GetDriverAlarms not support by offset: %d, limit: %d", offset, limit)
	}
	var (
		totalCount   int64
		driverEvents []db.DriverEvent
		opts         = make([]tgorm.Option, 0, 1)
	)
	opts = append(opts, withControllerIDOption(controllerID))
	opts = append(opts, withChannelID(channelID))
	err := queryAndCountDriverEvents(tgorm.WithOptions(d.db.WithContext(ctx).Model(&db.DriverEvent{}), opts...),
		offset, limit, &driverEvents, &totalCount)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, driverEvents, nil
}

// GetLastDriverCard 获取指定控制器和通道下最后一张驱动卡片。
func (d *impl) GetLastDriverCard(ctx context.Context, controllerID db.IDType, channelID string) (db.DriverCard, error) {
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID)}
	var lastCard db.DriverCard
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).Last(&lastCard).Error
	return lastCard, err
}

// GetFirstLogicDeleteDriverCard 获取指定控制器和通道下第一张已逻辑删除的驱动卡片。
func GetFirstLogicDeleteDriverCard(tx *gorm.DB, controllerID db.IDType, channelID string, status int) (db.DriverCard, error) {
	var card db.DriverCard
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withStatuses([]int{status})}
	err := tgorm.WithOptions(tx, opts...).First(&card).Error
	return card, err
}

// UpdateDriverCard 更新驱动卡片的卡号、标志、门列表、时间组等信息。
func UpdateDriverCard(tx *gorm.DB, controllerID db.IDType, oldCard db.DriverCard, card db.DriverCard) error {
	jsonDoorNos, err := json.Marshal(card.DoorNos)
	if err != nil {
		return err
	}
	opts := []tgorm.Option{
		withControllerIDOption(controllerID),
		withChannelID(oldCard.ChannelID),
		withCardNo(oldCard.CardNo),
	}
	return tgorm.WithOptions(
		tx.Model(&db.DriverCard{}), opts...,
	).Updates(map[string]interface{}{
		db.ColumnCardNo:      card.CardNo,
		db.ColumnCardFlag:    card.CardFlag,
		db.ColumnDoor:        string(jsonDoorNos),
		db.ColumnTimeGroupNo: card.TimeGroupNo,
		db.ColumnUserName:    card.UserName,
		db.ColumnPassword:    card.Password,
		db.ColumnCardIndex:   card.CardIndex,
		db.ColumnStatus:      card.Status,
	}).Error
}

// GetDriverAlarm 根据多条件精确匹配获取单条驱动告警。
func (d *impl) GetDriverAlarm(ctx context.Context, controllerID db.IDType, driverAlarm db.DriverAlarm) (db.DriverAlarm, error) {
	var res db.DriverAlarm
	opts := []tgorm.Option{withControllerIDOption(controllerID),
		withChannelID(driverAlarm.ChannelID),
		withTimestamp(driverAlarm.Timestamp),
		withDoorNumber(driverAlarm.DoorNumber),
		withType(int(driverAlarm.Type)),
		withState(driverAlarm.State)}
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).First(&res).Error
	return res, err
}

// GetDriverEvent 根据多条件精确匹配获取单条驱动事件。
func (d *impl) GetDriverEvent(ctx context.Context, controllerID db.IDType, driverEvent db.DriverEvent) (db.DriverEvent, error) {
	var res db.DriverEvent
	opts := []tgorm.Option{withControllerIDOption(controllerID),
		withChannelID(driverEvent.ChannelID),
		withTimestamp(driverEvent.Timestamp),
		withCardNumber(driverEvent.CardNumber),
		withDoorNumber(driverEvent.DoorNumber),
		withDirection(int(driverEvent.Direction)),
		withType(int(driverEvent.Type))}
	err := tgorm.WithOptions(d.db.WithContext(ctx), opts...).First(&res).Error
	return res, err
}

// GetMaxIndexDriverAlarm 获取指定控制器和通道下索引最大的驱动告警。
func GetMaxIndexDriverAlarm(tx *gorm.DB, controllerID db.IDType, channelID string) (db.DriverAlarm, error) {
	var res db.DriverAlarm
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withIndexDescOption()}
	err := tgorm.WithOptions(tx, opts...).First(&res).Error
	return res, err
}

// GetMaxIndexDriverEvent 获取指定控制器和通道下索引最大的驱动事件。
func GetMaxIndexDriverEvent(tx *gorm.DB, controllerID db.IDType, channelID string) (db.DriverEvent, error) {
	var res db.DriverEvent
	opts := []tgorm.Option{withControllerIDOption(controllerID), withChannelID(channelID), withIndexDescOption()}
	err := tgorm.WithOptions(tx, opts...).First(&res).Error
	return res, err
}
