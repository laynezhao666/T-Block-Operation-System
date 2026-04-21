// Package fetcher 提供门禁事件和告警数据的增量拉取框架。
package fetcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/logic/dlm"

	"dac/entity/utils/tlog"
)

// GetArgFun 获取拉取参数的函数类型
type GetArgFun func(ctx context.Context) interface{}

// GetControllerFun 获取控制器信息的函数类型
type GetControllerFun func(context.Context) rt.DoorController

// FetchNextFun 拉取下一批数据的函数类型，返回是否需要继续拉取
type FetchNextFun func(ctx context.Context) (bool, error)

// Fetcher 数据拉取器接口
type Fetcher interface {
	Start(ctx context.Context) error
	Stop()
}

// fetcher 数据拉取器基础实现
type fetcher struct {
	controllerID db.IDType // 控制器ID
	mozuID       string    // 模组ID

	mutex           sync.RWMutex
	stopChan        chan struct{} // 停止信号通道
	historyStopChan chan struct{} // 历史数据停止信号

	getControllerFun GetControllerFun // 获取控制器函数
	getArgFun        GetArgFun        // 获取参数函数
	fetchNext        FetchNextFun     // 拉取下一批数据函数

	loopWaitTime  time.Duration // 拉取循环等待时间
	fetchWaitTime time.Duration // 单次拉取等待时间
	logger        tlog.Logger   // 日志器
}

// newFetcher 创建新的数据拉取器实例
func newFetcher(controllerID db.IDType, name string,
	fetchNext FetchNextFun, loopWaitTime, fetchWaitTime time.Duration,
	getController GetControllerFun, getArg GetArgFun, mozuID string) *fetcher {
	f := new(fetcher)

	f.logger = tlog.NewPrefixLogger(fmt.Sprintf("[%v fetcher@%v]", name, controllerID), config.Log)

	f.logger.Infof("loop wait time: %v, fetch wait time: %v", loopWaitTime, fetchWaitTime)

	f.controllerID = controllerID
	f.mozuID = mozuID

	f.stopChan = make(chan struct{}, 1)
	f.historyStopChan = make(chan struct{}, 1)

	f.getControllerFun = getController
	f.getArgFun = getArg

	f.fetchNext = fetchNext
	f.loopWaitTime = loopWaitTime
	f.fetchWaitTime = fetchWaitTime

	return f
}

// fetch 执行一次数据拉取，循环调用fetchNext直到无更多数据
func (f *fetcher) fetch(ctx context.Context) {
	if !dlm.GetWorker().HasLock() {
		return
	}

	var (
		err           error
		needFetchNext bool
	)
	for {
		if f.fetchNext == nil {
			return
		}

		if needFetchNext, err = f.fetchNext(ctx); err != nil {
			if needFetchNext {
				f.logger.Warnf("fetch next error: %v", err)
			}
			return
		}
		if !needFetchNext {
			return
		}

		time.Sleep(f.fetchWaitTime)
	}
}

// fetchLoop 后台循环执行数据拉取
func (f *fetcher) fetchLoop(ctx context.Context) {
	f.logger.Infof("start fetch loop.")
	for {
		f.fetch(ctx)
		select {
		case <-time.After(f.loopWaitTime):
			break
		case <-f.stopChan:
			f.logger.Infof("stop loop.")
			return
		case <-ctx.Done():
			f.logger.Infof("cancel loop")
			return
		}
	}
}

// Stop 停止数据拉取器
func (f *fetcher) Stop() {
	f.stopChan <- struct{}{}
	f.historyStopChan <- struct{}{}
}
