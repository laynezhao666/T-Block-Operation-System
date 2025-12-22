// Package conf conf
package conf

import (
	"etrpc-go/config"
)

// ServerConfStruct 云端配置
type ServerConfStruct struct {
	SnowflakeConfig       SnowflakeConfigStruct  `yaml:"snowflake_config"`
	AlertManageConfig     BatchChannelConfStruct `yaml:"AlertManageConfig"`
	RestoreManageConfig   BatchChannelConfStruct `yaml:"RestoreManageConfig"`
	CollectorManageConfig BatchChannelConfStruct `yaml:"CollectorManageConfig"`
	SyncDeviceCacheConf   SyncDeviceCacheConf    `yaml:"SyncDeviceCacheConfig"`
	MysqlTableName        MysqlTableName         `yaml:"MysqlTableName"`
}

// SnowflakeConfigStruct 雪花算法配置
type SnowflakeConfigStruct struct {
	SetId          int32 `yaml:"set_id"`
	UpdateInterval int32 `yaml:"update_interval"`
}

// BatchChannelConfStruct batch通道配置
type BatchChannelConfStruct struct {
	BatchChannelSize     int32 `yaml:"BatchChannelSize"`
	BatchFetchIntervalMS int32 `yaml:"BatchFetchIntervalMS"`
	PoolSize             int32 `yaml:"PoolSize"`
}

// SyncDeviceCacheConf SyncDeviceCacheConf
type SyncDeviceCacheConf struct {
	BatchSize                 int32 `yaml:"BatchSize"`
	RegularSyncDeviceInterval int32 `yaml:"RegularSyncDeviceInterval"`
	TotalLoadIntervalCnt      int32 `yaml:"TotalLoadIntervalCnt"`
}

// MysqlTableName MysqlTableName
type MysqlTableName struct {
	ActiveAlarmTable  string `yaml:"ActiveAlarm"`
	HistoryAlarmTable string `yaml:"HistoryAlarm"`
}

// RobotUrlEle 单个企微群机器人url配置
type RobotUrlEle struct {
	Url string `yaml:"url"`
}

// RobotConfigStruct 企微群机器人配置
type RobotConfigStruct struct {
	Token          string                  `yaml:"token"`
	EncodingAESKey string                  `yaml:"encodingAESKey"`
	Webhook        map[int32][]RobotUrlEle `yaml:"webhook"`
}

var (
	ServerConf  ServerConfStruct
	RobotConfig RobotConfigStruct
)

func init() {
	config.RegisterConfig("serverconf", &ServerConf, true)
	config.RegisterConfig("robot_config", &RobotConfig, true)
}
