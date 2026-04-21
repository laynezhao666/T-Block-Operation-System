package dac

import (
	"context"
	"fmt"
	"time"

	"dac/entity/model/db"

	"gorm.io/gorm"
)

const (
	// timeGroupCount 默认时间组数量。
	timeGroupCount = 12
	// WeekDefault 默认星期配置（周一到周日）。
	WeekDefault = "[1,2,3,4,5,6,7]"
	// TimeZoneDefault 默认时区配置（全天 00:00-23:59）。
	TimeZoneDefault = "[{\"begin\": \"00:00\",\"end\": \"23:59\"}]"
)

// AddDefaultTimeGroups 添加默认时间组（仅当数据库中无时间组数据时执行）。
func (d *impl) AddDefaultTimeGroups(ctx context.Context) error {
	timeGroups := make([]db.TimeGroup, 0, timeGroupCount)
	t := time.Now().UTC()
	for i := 0; i < timeGroupCount; i++ {
		timeGroups = append(timeGroups, db.TimeGroup{
			GroupNo:    i,
			GroupName:  fmt.Sprintf("时间组%v", i),
			Week:       WeekDefault,
			TimeZone:   TimeZoneDefault,
			UpdateTime: t.Unix(),
		})
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		temp, err := GetAllTimeGroups(tx)
		if err != nil {
			return err
		}
		// 仅当时间组数据为空时添加默认时间组数据
		if len(temp) > 0 {
			return nil
		}
		return tx.Create(&timeGroups).Error
	})
}

// GetAllTimeGroups 获取所有时间组记录。
func GetAllTimeGroups(tx *gorm.DB) ([]db.TimeGroup, error) {
	var timeGroups []db.TimeGroup
	err := tx.Find(&timeGroups).Error
	return timeGroups, err
}

// GetAllTimeGroups 获取所有时间组记录（impl 方法）。
func (d *impl) GetAllTimeGroups(ctx context.Context) ([]db.TimeGroup, error) {
	return GetAllTimeGroups(d.db.WithContext(ctx))
}

// GetTimeGroup 根据时间组编号获取单个时间组。
func GetTimeGroup(tx *gorm.DB, groupNumber int) (db.TimeGroup, error) {
	var tg db.TimeGroup
	err := withEqual(tx, "group_no", groupNumber).First(&tg).Error
	return tg, err
}

// GetTimeGroupsByNos 根据时间组编号列表获取时间组映射。
func (d *impl) GetTimeGroupsByNos(ctx context.Context, groupNos []int) (map[int]db.TimeGroup, error) {
	result := make(map[int]db.TimeGroup)
	if len(groupNos) == 0 {
		return result, nil
	}

	var timeGroups []db.TimeGroup
	if err := withIn(d.db.WithContext(ctx), "group_no", groupNos).Find(&timeGroups).Error; err != nil {
		return result, err
	}
	for i := range timeGroups {
		t := &timeGroups[i]
		result[t.GroupNo] = *t
	}
	return result, nil
}

// UpdateTimeGroup 更新时间组（不更新时间组序号，支持前置和后置回调）。
func (d *impl) UpdateTimeGroup(ctx context.Context, timeGroup db.TimeGroup,
	beforeUpdate func(*gorm.DB) error, afterUpdate func(*gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error

		if beforeUpdate != nil {
			if err = beforeUpdate(tx); err != nil {
				return err
			}
		}

		temp := timeGroup
		temp.GroupNo = 0
		// 不会更新时间组序号
		if err = withEqual(tx, "group_no", timeGroup.GroupNo).Updates(&temp).Error; err != nil {
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
