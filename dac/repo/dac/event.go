package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetEventsByDoors 获取指定控制器和门编号组合下的刷卡事件（分页，按时间戳降序）。
func (d *impl) GetEventsByDoors(ctx context.Context, controllerDoors map[db.IDType][]int,
	offset, limit int, beginTime, endTime int64,
	afterGet func(*gorm.DB, []db.Event) error) (int64, []db.Event, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("invalid offset: %v, limit: %v", offset, limit)
	}
	if beginTime >= endTime {
		return 0, make([]db.Event, 0), nil
	}

	columnConditions := make([][]interface{}, 0, len(controllerDoors))
	for controllerID, doorNumbers := range controllerDoors {
		for _, doorNumber := range doorNumbers {
			columnConditions = append(columnConditions, []interface{}{controllerID, doorNumber})
		}
	}

	if len(columnConditions) == 0 {
		return 0, make([]db.Event, 0), nil
	}

	var (
		total  int64
		events []db.Event
		err    error
	)

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e error
		e = queryAndCountRecords(
			withTimestampDesc(
				withTimestampBetween(tx.Model(&db.Event{}), beginTime, endTime).
					Where("(controller_id, door_number) IN ?", columnConditions)),
			offset, limit, &events, &total)
		if e != nil {
			return e
		}
		if afterGet != nil {
			if e = afterGet(tx, events); e != nil {
				return e
			}
		}
		return nil
	})

	return total, events, err
}

// GetEvents 获取事件列表（支持模组、门名称、控制器、时间范围和模糊查询过滤）。
func GetEvents(tx *gorm.DB, mozuID string, doorName string, controllerDoors map[int][]int,
	beginTime, endTime int64, query string) ([]db.Event, error) {
	var (
		err    error
		events = make([]db.Event, 0)
		opts   = make([]tgorm.Option, 0, 3)
		orOpts = make([]tgorm.Option, 0, 2)
	)

	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = addDoorNameOptionIfNotEmpty(opts, doorName)
	opts = appendCtrlDoorsOptIfNotEmpty(opts, controllerDoors)
	opts = append(opts, withTimestampBetweenOption(beginTime, endTime), withTimestampDescOption())

	if len(query) > 0 {
		orOpts = append(orOpts, withCardNumberLike(query))
		orOpts = append(orOpts, withUsernameLike(query))
	}
	opts = append(opts, tgorm.WithOr(tx, orOpts...))

	err = tgorm.WithOptions(tx.Model(&db.Event{}), opts...).Find(&events).Error

	return events, err
}

// GetEventsNumber 获取符合条件的事件总数（支持后置回调）。
func (d *impl) GetEventsNumber(ctx context.Context, mozuID string, doorName string, controllerDoors map[int][]int,
	beginTime, endTime int64, query string, afterGet func(*gorm.DB, int64) error) (int64, error) {
	var total int64
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			e      error
			opts   = make([]tgorm.Option, 0, 3)
			orOpts = make([]tgorm.Option, 0, 2)
		)
		opts = addMozuOptionIfNotEmpty(opts, mozuID)
		opts = addDoorNameOptionIfNotEmpty(opts, doorName)
		opts = appendCtrlDoorsOptIfNotEmpty(opts, controllerDoors)
		opts = append(opts, withTimestampBetweenOption(beginTime, endTime))

		if len(query) > 0 {
			orOpts = append(orOpts, withCardNumberLike(query))
			orOpts = append(orOpts, withUsernameLike(query))
		}

		opts = append(opts, tgorm.WithOr(tx, orOpts...))

		if e = countRecord(tgorm.WithOptions(tx.Model(&db.Event{}), opts...), &total); e != nil {
			return e
		}

		if afterGet != nil {
			if e = afterGet(tx, total); e != nil {
				return e
			}
		}

		return nil
	})
	return total, err
}

// GetEvents 分页获取事件列表（impl 方法，支持多条件过滤和后置回调）。
func (d *impl) GetEvents(ctx context.Context, mozuID string, controllerIDs []db.IDType, query string, doorName string,
	offset, limit int, beginTime, endTime int64, afterGet func(*gorm.DB, []db.Event) error) (int64, []db.Event, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("invalid offset: %v, limit: %v", offset, limit)
	}
	if beginTime >= endTime {
		return 0, make([]db.Event, 0), nil
	}

	var (
		total  int64
		events []db.Event
		err    error
	)

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			e      error
			opts   = make([]tgorm.Option, 0, 4)
			orOpts = make([]tgorm.Option, 0, 2)
		)

		if len(query) > 0 {
			orOpts = append(orOpts, withCardNumberLike(query))
			orOpts = append(orOpts, withUsernameLike(query))
		}

		opts = addMozuOptionIfNotEmpty(opts, mozuID)
		opts = appendControllerIDsOptionIfNotEmpty(opts, controllerIDs)
		opts = append(opts, withTimestampBetweenOption(beginTime, endTime), withTimestampDescOption())
		if doorName != "" {
			opts = append(opts, withDoorName(doorName))
		}
		opts = append(opts, tgorm.WithOr(tx, orOpts...))

		e = queryAndCountRecords(tgorm.WithOptions(tx.Model(&db.Event{}), opts...), offset, limit, &events, &total)
		if e != nil {
			return e
		}
		if afterGet != nil {
			if e = afterGet(tx, events); e != nil {
				return e
			}
		}
		return nil
	})

	return total, events, err
}

// UpdateEventIndex 更新事件索引记录。
func UpdateEventIndex(tx *gorm.DB, controllerID db.IDType, index, last int) error {
	return updateIndex(tx, controllerID, index, last, &db.EventIndexRecord{})
}

// UpdateEventIndex 更新事件索引记录（impl 方法）。
func (d *impl) UpdateEventIndex(ctx context.Context, controllerID db.IDType, index, last int) error {
	return UpdateEventIndex(d.db.WithContext(ctx), controllerID, index, last)
}

// GetOrCreateEventIndex 获取或创建事件索引记录。
func (d *impl) GetOrCreateEventIndex(ctx context.Context, controllerID db.IDType) (db.EventIndexRecord, error) {
	var r db.EventIndexRecord
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return getOrCreateIndexer(tx, controllerID, &r)
	})
	return r, err
}

// GetOrCreateEventTimestampIndex 获取或创建事件时间戳索引记录。
func (d *impl) GetOrCreateEventTimestampIndex(ctx context.Context,
	controllerID db.IDType, mozuID string,
) (db.EventTimestampIndexRecord, error) {
	var r db.EventTimestampIndexRecord
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return getOrCreateTimestampIndexer(tx, controllerID, mozuID, &r)
	})
	return r, err
}

// getEventsCheckExist 查询已存在的事件，返回需要新增的事件列表（去重）。
func getEventsCheckExist(tx *gorm.DB, controllerID db.IDType,
	events []db.Event, beginTime, endTime int64,
) ([]db.Event, error) {
	var existedEvents []db.Event
	err := tgorm.WithOptions(tx,
		withControllerIDOption(controllerID),
		withTimestampBetweenOption(beginTime, endTime)).Find(&existedEvents).Error
	if err != nil {
		return nil, fmt.Errorf("get existed events error: %v", err)
	}

	m := make(map[db.EventKey]struct{}, len(existedEvents))
	for i := range existedEvents {
		m[existedEvents[i].GetKey()] = struct{}{}
	}

	toAddEvents := make([]db.Event, 0, len(events))
	for i := range events {
		if _, ok := m[events[i].GetKey()]; ok {
			continue
		}
		toAddEvents = append(toAddEvents, events[i])
	}

	return toAddEvents, nil
}

// createEvents 批量创建事件记录（支持去重检查，冲突时全量更新）。
func createEvents(tx *gorm.DB, controllerID db.IDType, events []db.Event,
	checkExist bool, beginTime int64, endTime int64,
	fillEvent func(e *db.Event),
) error {
	if len(events) == 0 {
		return nil
	}

	for i := range events {
		events[i].ControllerID = controllerID
		fillEvent(&events[i])
	}

	var err error
	if checkExist {
		events, err = getEventsCheckExist(tx, controllerID, events, beginTime, endTime)
		if err != nil {
			return fmt.Errorf("get existed events error: %w", err)
		}
		if len(events) == 0 {
			return nil
		}
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: db.ColumnControllerID}, {Name: db.ColumnIndex}, {Name: db.ColumnTimestamp}},
		UpdateAll: true,
	}).Create(&events).Error
}

// AddEvents 批量添加事件记录（在事务中执行，支持后置回调）。
func (d *impl) AddEvents(ctx context.Context, controllerID db.IDType, events []db.Event,
	checkExist bool, beginTime int64, endTime int64,
	fillEvent func(e *db.Event), afterAdd func(*gorm.DB) error) error {

	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = createEvents(tx, controllerID, events, checkExist, beginTime, endTime, fillEvent); err != nil {
			return err
		}
		if afterAdd != nil {
			if err = afterAdd(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateEventCurrentTimestampIndex 更新事件当前同步时间戳索引。
func UpdateEventCurrentTimestampIndex(tx *gorm.DB, controllerID db.IDType, timestamp int64) error {
	return updateCurrentSyncedTimestampIndexer(tx, controllerID, timestamp, &db.EventTimestampIndexRecord{})
}

// UpdateEventHistoryTimestampIndex 更新事件历史同步时间戳索引。
func UpdateEventHistoryTimestampIndex(tx *gorm.DB, controllerID db.IDType, timestamp int64) error {
	return updateHistorySyncedTimestampIndexer(tx, controllerID, timestamp, &db.EventTimestampIndexRecord{})
}
