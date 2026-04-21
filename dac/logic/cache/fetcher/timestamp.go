// Package fetcher 提供门禁事件和告警数据的增量拉取框架。
package fetcher

import (
	"context"
	"fmt"
	"time"

	"dac/entity/model/db"
	"dac/logic/dlm"

	"dac/entity/utils/ttime"
)

// daySeconds 一天的秒数
const (
	daySeconds = 24 * 3600
)

// FetchHistoryIndexByTimestampFunction 根据时间戳获取历史数据索引的函数类型
type FetchHistoryIndexByTimestampFunction func(context.Context, db.IDType, string) (int64, int64, int64, error)

// FetchDataByTimestampFunction 根据时间戳范围拉取数据的函数类型
type FetchDataByTimestampFunction func(context.Context, db.IDType, int64, int64, GetControllerFun, GetArgFun) (int64, error)

// TimestampFetcher 基于时间戳的增量数据拉取器
type TimestampFetcher struct {
	*fetcher
	timestamp                       int64                                // 当前同步到的时间戳
	historyBeginTimestamp           int64                                // 历史数据起始时间戳
	historySyncedTimestamp          int64                                // 历史数据已同步到的时间戳
	fetchHistoryIndexByTimestampFun FetchHistoryIndexByTimestampFunction // 获取历史索引函数
	fetchDataByTimestampFun         FetchDataByTimestampFunction         // 按时间戳拉取数据函数
}

// NewByTimestamp 创建基于时间戳的增量拉取器
func NewByTimestamp(controllerID db.IDType, name string, loopWaitTime, fetchWaitTime time.Duration,
	fetchHistoryIndexByTimestampFun FetchHistoryIndexByTimestampFunction,
	fetchDataByTimestampFun FetchDataByTimestampFunction,
	getController GetControllerFun, getArg GetArgFun, mozuID string) *TimestampFetcher {
	f := new(TimestampFetcher)
	f.fetcher = newFetcher(controllerID, name, f.fetchNext, loopWaitTime, fetchWaitTime, getController, getArg, mozuID)

	f.fetchHistoryIndexByTimestampFun = fetchHistoryIndexByTimestampFun
	f.fetchDataByTimestampFun = fetchDataByTimestampFun

	return f
}

// Start 启动时间戳拉取器，加载初始索引并启动拉取协程
func (f *TimestampFetcher) Start(ctx context.Context) error {
	if f.fetchHistoryIndexByTimestampFun == nil {
		return fmt.Errorf("nil function")
	}

	currentSynced, historyBegin, historySynced, err := f.fetchHistoryIndexByTimestampFun(ctx, f.controllerID, f.mozuID)
	if err != nil {
		return err
	}

	f.setTimestamp(currentSynced)
	f.setHistoryBeginTimestamp(historyBegin)
	f.setHistorySyncedTimestamp(historySynced)

	go f.fetchLoop(ctx)
	go f.fetchHistoryLoop(ctx)

	return nil
}

// setTimestamp 设置当前同步时间戳
func (f *TimestampFetcher) setTimestamp(t int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.timestamp = t
}

// getTimestamp 获取当前同步时间戳
func (f *TimestampFetcher) getTimestamp() int64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.timestamp
}

// setHistoryBeginTimestamp 设置历史数据起始时间戳
func (f *TimestampFetcher) setHistoryBeginTimestamp(t int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.historyBeginTimestamp = t
}

// setHistorySyncedTimestamp 设置历史数据已同步时间戳
func (f *TimestampFetcher) setHistorySyncedTimestamp(t int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.historySyncedTimestamp = t
}

// getHistoryTimestamp 获取历史数据的起始和已同步时间戳
func (f *TimestampFetcher) getHistoryTimestamp() (int64, int64) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.historyBeginTimestamp, f.historySyncedTimestamp
}

// fetchNext 拉取下一批增量数据（向前同步）
func (f *TimestampFetcher) fetchNext(ctx context.Context) (bool, error) {
	if f.fetchDataByTimestampFun == nil {
		return false, nil
	}

	t := f.getTimestamp()
	currentTime := ttime.GetNowLocal().Unix()
	waitTimeSecond := int64(f.fetchWaitTime.Seconds())
	needFetchNext := t+waitTimeSecond < currentTime
	if !needFetchNext {
		// 时间已同步到最新
		return false, nil
	}

	newT, err := f.fetchDataByTimestampFun(ctx, f.controllerID, t, 0, f.getControllerFun, f.getArgFun)
	if err != nil {
		return needFetchNext, fmt.Errorf("get data by timestamp error: %w", err)
	}

	f.setTimestamp(newT)

	return needFetchNext, nil
}

// fetchHistoryLoop 后台循环拉取历史数据（向后回溯）
func (f *TimestampFetcher) fetchHistoryLoop(ctx context.Context) {
	f.logger.Infof("start fetch history loop.")
	for {
		beg, synced := f.getHistoryTimestamp()
		if beg >= synced {
			f.logger.Infof("all history data synced, history begin: %v, history synced: %v", beg, synced)
			return
		}

		f.fetchHistory(ctx)
		select {
		case <-time.After(f.loopWaitTime):
			break
		case <-f.historyStopChan:
			f.logger.Infof("stop history loop.")
			return
		case <-ctx.Done():
			f.logger.Infof("cancel history loop")
		}
	}
}

// fetchPrevious 拉取一天的历史数据（向后回溯）
func (f *TimestampFetcher) fetchPrevious(ctx context.Context) (bool, error) {
	if f.fetchDataByTimestampFun == nil {
		return false, nil
	}

	beg, synced := f.getHistoryTimestamp()
	if synced <= beg {
		return false, nil
	}

	newSynced := synced - daySeconds
	if newSynced < beg {
		newSynced = beg
	}

	_, err := f.fetchDataByTimestampFun(ctx, f.controllerID, newSynced, synced, f.getControllerFun, f.getArgFun)
	if err != nil {
		return true, fmt.Errorf("get data by history timestamp error: %w", err)
	}

	f.logger.Infof("success fetch history from %v to %v", newSynced, synced)

	f.setHistorySyncedTimestamp(newSynced)

	return newSynced > beg, nil
}

// fetchHistory 执行一次完整的历史数据回溯拉取
func (f *TimestampFetcher) fetchHistory(ctx context.Context) {
	if !dlm.GetWorker().HasLock() {
		return
	}

	var (
		err               error
		needFetchPrevious bool
	)

	for {
		if needFetchPrevious, err = f.fetchPrevious(ctx); err != nil {
			f.logger.Warnf("fetch previous error: %v", err)
			return
		}

		if !needFetchPrevious {
			return
		}

		time.Sleep(f.fetchWaitTime)
	}
}
