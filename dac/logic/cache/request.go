package cache

import (
	"context"
	"fmt"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/collect/dispatcher"
	"dac/logic/dlm"
	"dac/repo/dac"
)

const (
	fetchRequestNumber = 300

	maxWaitTime = time.Hour
)

func deleteRequests(ctx context.Context, invalidReqs map[db.IDType]struct{}) {
	toDeleteReqs := make([]db.IDType, 0, len(invalidReqs))

	for id := range invalidReqs {
		toDeleteReqs = append(toDeleteReqs, id)
	}

	if len(toDeleteReqs) == 0 {
		return
	}

	retry := 100
	// 重试最多一百次删除请求，记录日志
	utils.Retry(retry, maxWaitTime, func() error {
		return dac.GetRW().DeleteRequests(ctx, toDeleteReqs)
	}, func() {
		config.Log.Infof("delete invalid requests: %v", invalidReqs)
	}, func(e error) {
		config.Log.Warnf("delete requests %v error: %v", toDeleteReqs, e)
	}, func(e error) {
		config.Log.Errorf("delete requests %v failed...", toDeleteReqs)
	})
}

func updateSuccessRequests(ctx context.Context, successReqs []db.IDType) {
	if err := dac.GetRW().UpdateSuccessRequests(ctx, successReqs); err != nil {
		config.Log.Warnf("update success requests %v, error: %v", successReqs, err)
	}
}

// refreshRequests 定期检查并处理数据库中未处理的请求
func (c *Cache) refreshRequests(ctx context.Context) {
	// 1. 检查是否有锁
	if !dlm.GetWorker().HasLock() {
		config.Log.Infof("has no lock, do not refresh requests, sleeping...")
		return
	}

	// 2. 获取指定数量的未处理的请求
	reqs, err := dac.GetRW().FetchRequests(ctx, fetchRequestNumber)
	if err != nil {
		config.Log.Warnf("fetch requests error: %v", err)
		return
	}
	if len(reqs) == 0 {
		return
	}

	successReqs := make([]db.IDType, 0, len(reqs))
	failedReqs := make([]db.IDType, 0, len(reqs))
	invalidReqs := make(map[db.IDType]struct{}, len(reqs))

	finalReqs := make([]db.Request, 0, len(reqs))
	for i := range reqs {
		r := &reqs[i]

		// 如果门禁控制器已不存在，需要将其删除
		if !Get().HasController(r.ControllerID) {
			config.Log.Infof("controller %v not exist", r.ControllerID)
			invalidReqs[r.ID] = struct{}{}
			continue
		}

		finalReqs = append(finalReqs, *r)
	}

	// 3. 将请求分发给控制器，异步处理请求
	results := dispatcher.Get().DoAsyncRequests(finalReqs)
	messages := make(map[db.IDType]string, len(finalReqs))
	for i := range finalReqs {
		id := finalReqs[i].ID
		err, ok := results[id]
		if !ok || err != nil {
			failedReqs = append(failedReqs, id)
			if err != nil {
				messages[id] = err.Error()
			} else {
				messages[id] = fmt.Sprintf("result not found")
			}
		} else {
			successReqs = append(successReqs, id)
		}
	}
	// 4. 成功处理的请求需要更新state，不合理的请求直接删除，失败的请求需要更新记录状态和错误信息
	go deleteRequests(ctx, invalidReqs)
	go updateSuccessRequests(ctx, successReqs)

	if err = dac.GetRW().UpdateFailedRequests(ctx, failedReqs, messages); err != nil {
		config.Log.Warnf("update failed requests %v, messages %v error: %v", failedReqs, messages, err)
	}
}

// refreshRequestTime 循环，定期执行处理请求
func (c *Cache) refreshRequestLoop(ctx context.Context) {
	for {
		c.refreshRequests(ctx)
		select {
		case <-time.After(refreshRequestTime):
			break
		case <-ctx.Done():
			config.Log.Infof("stop refresh request loop.")
			return
		}
	}
}
