package dac

import (
	"context"
	"fmt"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"

	tgorm "etrpc-go/client/gorm"
	"gorm.io/gorm"
)

// Init 初始化数据库连接并自动迁移所有表结构。
func (d *impl) Init() error {
	var err error

	if d.db, err = tgorm.NewClientProxy(consts.ClientMysql); err != nil {
		return fmt.Errorf("NewClientProxy %v error: %w", consts.ClientMysql, err)
	}

	config.Log.Info("init tables...")
	if err = initTables(d.db.Migrator()); err != nil {
		return fmt.Errorf("init tables error: %w", err)
	}

	return nil
}

// initTable 自动迁移单个表结构。
func initTable(m gorm.Migrator, t interface{}) error {
	return m.AutoMigrate(t)
}

// initTimeGroupTable 初始化时间组表并添加默认时间组数据。
func initTimeGroupTable(m gorm.Migrator) error {
	if err := initTable(m, &db.TimeGroup{}); err != nil {
		return err
	}
	if err := GetRW().AddDefaultTimeGroups(context.Background()); err != nil {
		return err
	}

	return nil
}

// tableEntry 定义表结构与名称的映射。
type tableEntry struct {
	model interface{}
	name  string
}

// initTables 初始化所有数据库表结构（包括门禁控制器、门、请求、人员、卡片、权限组、事件、告警等）。
func initTables(m gorm.Migrator) error {
	tables := []tableEntry{
		{&db.DoorController{}, "door controller"},
		{&db.DoorGroup{}, "door group"},
		{&db.Door{}, "door"},
		{&db.Request{}, "request"},
		{&db.Staff{}, "staff"},
		{&db.Card{}, "card"},
		{&db.AccessGroup{}, "access_group"},
		{&db.AccessGroupRelation{}, "access_group_relation"},
		{&db.CardAccessRelation{}, "card_access_relation"},
		{&db.Event{}, "event"},
		{&db.EventIndexRecord{}, "event index record"},
		{&db.EventTimestampIndexRecord{}, "event timestamp index record"},
		{&db.Alarm{}, "alarm"},
		{&db.AlarmIndexRecord{}, "alarm index record"},
		{&db.AlarmTimestampIndexRecord{}, "alarm timestamp index record"},
	}

	for _, t := range tables {
		if err := initTable(m, t.model); err != nil {
			return fmt.Errorf("init %s table error: %w", t.name, err)
		}
	}

	if err := initTimeGroupTable(m); err != nil {
		return fmt.Errorf("init time_group table error: %w", err)
	}

	// 初始化存储厂商私有协议需要保存的数据
	driverTables := []tableEntry{
		{&db.DriverCard{}, "driver_card"},
		{&db.DriverTimeGroup{}, "driver_time_group"},
		{&db.DriverDoorParameter{}, "driver_door_parameter"},
		{&db.DriverEvent{}, "driver_event"},
		{&db.DriverAlarm{}, "driver_alarm"},
	}

	for _, t := range driverTables {
		if err := initTable(m, t.model); err != nil {
			return fmt.Errorf("init %s table error: %w", t.name, err)
		}
	}

	return nil
}
