package dac

import (
	"errors"
	"time"

	"dac/entity/model/db"
	"dac/entity/utils"

	tgorm "dac/entity/utils/tgorm"
	"gorm.io/gorm"
)

// ControllerIDSetter 设置控制器ID的接口。
type ControllerIDSetter interface {
	SetControllerID(controllerID db.IDType)
}

// MozuIDSetter 设置模组ID的接口。
type MozuIDSetter interface {
	SetMozuID(mozuID string)
}

// IndexSetter 设置索引和最后位置的接口。
type IndexSetter interface {
	SetIndex(index int)
	SetLast(last int)
}

// TimestampSetter 设置时间戳索引的接口。
type TimestampSetter interface {
	SetHistoryBeginTimestamp(timestamp int64)
	SetHistorySyncedTimestamp(timestamp int64)
	SetCurrentSyncedTimestamp(timestamp int64)
}

// Indexer 索引器接口（组合 IndexSetter 和 ControllerIDSetter）。
type Indexer interface {
	IndexSetter
	ControllerIDSetter
}

// TimestampIndexer 时间戳索引器接口（组合 TimestampSetter、ControllerIDSetter 和 MozuIDSetter）。
type TimestampIndexer interface {
	TimestampSetter
	ControllerIDSetter
	MozuIDSetter
}

// getOrCreateIndexer 获取或创建索引记录（不存在时初始化为0并创建）。
func getOrCreateIndexer(tx *gorm.DB, controllerID db.IDType, indexer Indexer) error {
	e := tgorm.WithOptions(tx, withControllerIDOption(controllerID)).First(indexer).Error
	if e == nil {
		return nil
	}
	if !errors.Is(e, gorm.ErrRecordNotFound) {
		return e
	}

	indexer.SetIndex(0)
	indexer.SetLast(0)
	indexer.SetControllerID(controllerID)
	return tx.Create(indexer).Error
}

// getOrCreateTimestampIndexer 获取或创建时间戳索引记录（不存在时初始化为当前时间并创建）。
func getOrCreateTimestampIndexer(tx *gorm.DB, controllerID db.IDType, mozuID string, indexer TimestampIndexer) error {
	e := tgorm.WithOptions(tx, withControllerIDOption(controllerID)).First(indexer).Error
	if e == nil {
		return nil
	}
	if !errors.Is(e, gorm.ErrRecordNotFound) {
		return e
	}

	t := time.Now().UTC().Unix()
	indexer.SetHistoryBeginTimestamp(utils.GetHistoryBeginTimestamp())
	indexer.SetHistorySyncedTimestamp(t)
	indexer.SetCurrentSyncedTimestamp(t)
	indexer.SetControllerID(controllerID)
	indexer.SetMozuID(mozuID)
	return tx.Create(indexer).Error
}
