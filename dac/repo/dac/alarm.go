package dac

import (
	"context"
	"fmt"

	"dac/entity/model/db"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetAlarms 获取告警列表（支持模组、控制器和时间范围过滤，按时间戳降序）。
func GetAlarms(tx *gorm.DB, mozuID string, controllerIDs []db.IDType,
	beginTime, endTime int64) ([]db.Alarm, error) {
	var (
		err    error
		alarms = make([]db.Alarm, 0)
		opts   = make([]tgorm.Option, 0, 3)
	)
	opts = addMozuOptionIfNotEmpty(opts, mozuID)
	opts = appendControllerIDsOptionIfNotEmpty(opts, controllerIDs)
	opts = append(opts, withTimestampBetweenOption(beginTime, endTime), withTimestampDescOption())
	err = tgorm.WithOptions(tx.Model(&db.Alarm{}), opts...).Find(&alarms).Error
	return alarms, err
}

// GetAlarmsNumber 获取符合条件的告警总数（支持后置回调）。
func (d *impl) GetAlarmsNumber(ctx context.Context, mozuID string, controllerIDs []db.IDType,
	beginTime, endTime int64, afterGet func(*gorm.DB, int64) error) (int64, error) {
	var total int64
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			e    error
			opts = make([]tgorm.Option, 0, 3)
		)
		opts = addMozuOptionIfNotEmpty(opts, mozuID)
		opts = appendControllerIDsOptionIfNotEmpty(opts, controllerIDs)
		opts = append(opts, withTimestampBetweenOption(beginTime, endTime))

		e = countRecord(tgorm.WithOptions(tx.Model(&db.Alarm{}), opts...), &total)
		if e != nil {
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

// GetAlarms 分页获取告警列表（impl 方法，支持多条件过滤和后置回调）。
func (d *impl) GetAlarms(ctx context.Context, mozuID string,
	controllerIDs []db.IDType, offset, limit int,
	beginTime, endTime int64,
	afterGet func(*gorm.DB, []db.Alarm) error,
) (int64, []db.Alarm, error) {
	if offset < 0 || limit <= 0 {
		return 0, nil, fmt.Errorf("invalid offset: %v, limit: %v", offset, limit)
	}
	if beginTime >= endTime {
		return 0, make([]db.Alarm, 0), nil
	}

	var (
		total  int64
		alarms []db.Alarm
		err    error
	)

	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			e    error
			opts = make([]tgorm.Option, 0, 4)
		)
		opts = addMozuOptionIfNotEmpty(opts, mozuID)
		opts = appendControllerIDsOptionIfNotEmpty(opts, controllerIDs)
		opts = append(opts, withTimestampBetweenOption(beginTime, endTime), withTimestampDescOption())

		e = queryAndCountRecords(tgorm.WithOptions(tx.Model(&db.Alarm{}), opts...), offset, limit, &alarms, &total)
		if e != nil {
			return e
		}

		if afterGet != nil {
			if e = afterGet(tx, alarms); e != nil {
				return e
			}
		}

		return nil
	})

	return total, alarms, err
}

// getAlarmsCheckExist 查询已存在的告警，返回需要新增的告警列表（去重）。
func getAlarmsCheckExist(tx *gorm.DB, controllerID db.IDType,
	alarms []db.Alarm, beginTime, endTime int64,
) ([]db.Alarm, error) {
	var existedAlarms []db.Alarm
	err := tgorm.WithOptions(tx,
		withControllerIDOption(controllerID),
		withTimestampBetweenOption(beginTime, endTime)).Find(&existedAlarms).Error
	if err != nil {
		return nil, fmt.Errorf("get existed alarms error: %v", err)
	}

	m := make(map[db.AlarmKey]struct{}, len(existedAlarms))
	for i := range existedAlarms {
		m[existedAlarms[i].GetKey()] = struct{}{}
	}

	toAddAlarms := make([]db.Alarm, 0, len(alarms))
	for i := range alarms {
		if _, ok := m[alarms[i].GetKey()]; ok {
			continue
		}
		toAddAlarms = append(toAddAlarms, alarms[i])
	}

	return toAddAlarms, nil
}

// createAlarms 批量创建告警记录（支持去重检查，冲突时全量更新）。
func createAlarms(tx *gorm.DB, controllerID db.IDType, alarms []db.Alarm,
	checkExist bool, beginTime, endTime int64, fillAlarm func(*db.Alarm)) error {
	if len(alarms) == 0 {
		return nil
	}

	for i := range alarms {
		alarms[i].ControllerID = controllerID
		fillAlarm(&alarms[i])
	}

	var err error
	if checkExist {
		alarms, err = getAlarmsCheckExist(tx, controllerID, alarms, beginTime, endTime)
		if err != nil {
			return fmt.Errorf("check existed alarms error: %w", err)
		}
		if len(alarms) == 0 {
			return nil
		}
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: db.ColumnControllerID}, {Name: db.ColumnIndex}, {Name: db.ColumnTimestamp}},
		UpdateAll: true,
	}).Create(&alarms).Error
}

// AddAlarms 批量添加告警记录（在事务中执行，支持后置回调）。
func (d *impl) AddAlarms(ctx context.Context, controllerID db.IDType, alarms []db.Alarm,
	checkExist bool, beginTime int64, endTime int64,
	fillAlarm func(*db.Alarm), afterAdd func(*gorm.DB) error) error {

	var err error
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = createAlarms(tx, controllerID, alarms, checkExist, beginTime, endTime, fillAlarm); err != nil {
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

// GetOrCreateAlarmIndex 获取或创建告警索引记录。
func (d *impl) GetOrCreateAlarmIndex(ctx context.Context, controllerID db.IDType) (db.AlarmIndexRecord, error) {
	var r db.AlarmIndexRecord
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return getOrCreateIndexer(tx, controllerID, &r)
	})
	return r, err
}

// GetOrCreateAlarmTimestampIndex 获取或创建告警时间戳索引记录。
func (d *impl) GetOrCreateAlarmTimestampIndex(ctx context.Context,
	controllerID db.IDType, mozuID string,
) (db.AlarmTimestampIndexRecord, error) {
	var r db.AlarmTimestampIndexRecord
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return getOrCreateTimestampIndexer(tx, controllerID, mozuID, &r)
	})
	return r, err
}

// UpdateAlarmIndex 更新告警索引记录。
func UpdateAlarmIndex(tx *gorm.DB, controllerID db.IDType, index, last int) error {
	return updateIndex(tx, controllerID, index, last, &db.AlarmIndexRecord{})
}

// UpdateAlarmIndex 更新告警索引记录（impl 方法）。
func (d *impl) UpdateAlarmIndex(ctx context.Context, controllerID db.IDType, index, last int) error {
	return UpdateAlarmIndex(d.db.WithContext(ctx), controllerID, index, last)
}

// UpdateAlarmCurrentTimestampIndex 更新告警当前同步时间戳索引。
func UpdateAlarmCurrentTimestampIndex(tx *gorm.DB, controllerID db.IDType, timestamp int64) error {
	return updateCurrentSyncedTimestampIndexer(tx, controllerID, timestamp, &db.AlarmTimestampIndexRecord{})
}

// UpdateAlarmHistoryTimestampIndex 更新告警历史同步时间戳索引。
func UpdateAlarmHistoryTimestampIndex(tx *gorm.DB, controllerID db.IDType, timestamp int64) error {
	return updateHistorySyncedTimestampIndexer(tx, controllerID, timestamp, &db.AlarmTimestampIndexRecord{})
}
