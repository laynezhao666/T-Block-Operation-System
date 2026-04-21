// Package fetcher 提供门禁事件和告警的增量拉取框架。
package fetcher

import (
	"context"
	"fmt"
	"time"

	"dac/entity/model/db"
	"dac/entity/utils"
)

// FetchIndexFunction 获取索引的函数类型
type FetchIndexFunction func(context.Context, db.IDType) (int, int, error)

// FetchDataFunction 按索引拉取数据的函数类型
type FetchDataFunction func(context.Context, db.IDType, int, int, GetControllerFun, GetArgFun) (int, int, error)

// IndexFetcher 基于索引的增量数据拉取器
type IndexFetcher struct {
	*fetcher
	index         int                // 当前已同步的索引
	last          int                // 控制器中最新的索引
	fetchIndexFun FetchIndexFunction // 获取索引函数
	fetchDataFun  FetchDataFunction  // 拉取数据函数
}

// NewByIndex 创建基于索引的增量拉取器实例
func NewByIndex(controllerID db.IDType, name string,
	loopWaitTime, fetchWaitTime time.Duration,
	fetchIndex FetchIndexFunction,
	fetchData FetchDataFunction,
	getController GetControllerFun,
	getArg GetArgFun, mozuID string,
) *IndexFetcher {
	f := new(IndexFetcher)

	f.fetcher = newFetcher(controllerID, name, f.fetchNext, loopWaitTime, fetchWaitTime, getController, getArg, mozuID)

	f.fetchIndexFun = fetchIndex
	f.fetchDataFun = fetchData

	return f
}

// setIndex 设置当前索引和最新索引
func (f *IndexFetcher) setIndex(index, last int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.index = index
	f.last = last
}

// getIndex 获取当前索引和最新索引
func (f *IndexFetcher) getIndex() (int, int) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.index, f.last
}

// Start 启动索引拉取器，初始化索引并开始循环拉取
func (f *IndexFetcher) Start(ctx context.Context) error {
	if f.fetchIndexFun == nil {
		return fmt.Errorf("nil function")
	}

	index, last, err := f.fetchIndexFun(ctx, f.controllerID)
	if err != nil {
		return err
	}

	f.setIndex(index, last)

	go f.fetchLoop(ctx)

	return nil
}

// fetchNext 判断记录是否最新，需要继续获取
func (f *IndexFetcher) fetchNext(ctx context.Context) (bool, error) {
	if f.fetchDataFun == nil {
		f.logger.Warnf("fetchDataFun not found.")
		return false, nil
	}

	index, last := f.getIndex()

	if index > last {
		// 若当前索引超过最后索引，则门禁控制器中的刷卡记录已重置
		// 需要重新从 0 开始拉取
		f.logger.Warnf("index reset, index: %v, last: %v", index, last)
		f.setIndex(0, 0)
		return true, nil
	}

	needFetchNext := true
	// 若已是最新记录，需要定时更新
	if index == last && index > 0 {
		needFetchNext = false
	}
	// 若不是最新记录，需要立刻更新
	newIndex, newLast, err := f.fetchDataFun(ctx, f.controllerID, index, last,
		f.getControllerFun, f.getArgFun)
	if err != nil {
		if utils.IsRecordIndexOutOfRange(err) {
			// 若当前记录总数比先前记录总数少，导致参数超出范围，说明门禁被重置，需要重新从 0 开始拉取
			f.setIndex(0, 0)
			return true, nil
		}
		f.logger.Warnf("fetch data error: %v", err)
		return needFetchNext, err
	}

	f.setIndex(newIndex, newLast)

	return needFetchNext, nil
}
