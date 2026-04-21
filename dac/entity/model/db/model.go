// Package db 定义门禁系统的数据库模型和表结构。
package db

import (
	"dac/entity/utils/tgorm"
)

// TimestampIndexRecord 基于起止时间同步记录时使用的结构
// 程序并行同步当前数据和历史数据，
// HistorySyncedTimestamp 和 CurrentSyncedTimestamp 初始化为开始同步数据的时间
// HistoryBeginTimestamp 初始化为 2024-01-01 00:00:00
// 1, 当前数据持续同步，CurrentSyncedTimestamp 更新为每次同步完成的时间
// 2, 历史数据同步直到 HistorySyncedTimestamp <= HistoryBeginTimestamp
//
//	历史数据每次同步成功后，HistorySyncedTimestamp 减小
type TimestampIndexRecord struct {
	ControllerID IDType `gorm:"primaryKey;column:controller_id"`
	// [HistoryBeginTimestamp, HistorySyncedTimestamp) 未同步
	// [HistorySyncedTimestamp, CurrentSyncedTimestamp) 已同步
	// [CurrentSyncedTimestamp, +∞) 未同步
	HistoryBeginTimestamp  int64  `gorm:"column:history_begin_timestamp"`  // 历史数据开始时间
	HistorySyncedTimestamp int64  `gorm:"column:history_synced_timestamp"` // 历史数据已完成同步时间
	CurrentSyncedTimestamp int64  `gorm:"column:current_synced_timestamp"` // 当前数据已完成同步结束
	MozuID                 string `gorm:"column:mozu_id"`
	tgorm.Model
}

// SetControllerID 设置控制器ID
func (r *TimestampIndexRecord) SetControllerID(controllerID IDType) {
	r.ControllerID = controllerID
}

// SetHistoryBeginTimestamp 设置历史数据开始时间
func (r *TimestampIndexRecord) SetHistoryBeginTimestamp(timestamp int64) {
	r.HistoryBeginTimestamp = timestamp
}

// SetHistorySyncedTimestamp 设置历史数据已同步时间
func (r *TimestampIndexRecord) SetHistorySyncedTimestamp(timestamp int64) {
	r.HistorySyncedTimestamp = timestamp
}

// SetCurrentSyncedTimestamp 设置当前数据已同步时间
func (r *TimestampIndexRecord) SetCurrentSyncedTimestamp(timestamp int64) {
	r.CurrentSyncedTimestamp = timestamp
}

// SetMozuID 设置模组ID
func (r *TimestampIndexRecord) SetMozuID(mozuID string) {
	r.MozuID = mozuID
}
