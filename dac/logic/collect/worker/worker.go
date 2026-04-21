// Package worker 提供门禁控制器的工作通道管理。
package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	ctrl "dac/logic/collect/controller"

	"dac/entity/utils/tlog"
)

// Channel 控制器工作通道，管理单个控制器的生命周期和请求处理
type Channel struct {
	channelID  string
	controller *ctrl.Controller
	logger     tlog.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewChannel 创建新的工作通道
func NewChannel(channelID string) *Channel {
	c := new(Channel)
	c.channelID = channelID
	c.logger = tlog.NewPrefixLogger(
		fmt.Sprintf("[channel %v]: ", c.channelID),
		config.Log)
	return c
}

// Start 启动控制器工作通道，开始处理请求
func (c *Channel) Start(
	ctx context.Context, controller rt.DoorController,
) {
	if c.controller != nil {
		c.Stop()
	}

	c.controller = ctrl.NewController(ctx, controller)
	if c.controller == nil {
		c.Warnf("new controller %+v failed", controller)
		return
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	c.Infof("start.")

	go c.requestLoop(c.ctx)
}

// StopController 停止指定ID的控制器，返回是否匹配
func (c *Channel) StopController(id db.IDType) bool {
	if c.controller == nil || c.controller.ID() != id {
		return false
	}

	c.Stop()

	return true
}

// Stop 停止工作通道，关闭控制器连接
func (c *Channel) Stop() {
	if c.controller == nil {
		return
	}

	c.controller.Close()

	c.Infof("stop.")

	if c.cancel != nil {
		c.cancel()
	}
}

// requestLoop 后台请求处理循环
func (c *Channel) requestLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("ctx is done, stop request loop.")
			return
		case <-time.After(time.Second):
			c.request(ctx)
		}
	}
}

// request 执行一次请求处理
func (c *Channel) request(ctx context.Context) {
	if c.controller == nil {
		return
	}

	c.controller.DoRequest(ctx)
	c.controller.WaitBetweenRequest()
}

// DoGeneralRequest 同步执行单个门禁请求并返回响应
func (c *Channel) DoGeneralRequest(
	req *db.Request,
) (interface{}, error) {
	if c == nil || req == nil || c.controller == nil {
		logStr := ""
		if c == nil {
			logStr = fmt.Sprintf("nil Channel")
		} else if req == nil {
			logStr = fmt.Sprintf("nil req")
		} else {
			logStr = fmt.Sprintf("nil controller")
		}

		return nil, errors.New("nil pointer: " + logStr)
	}

	return c.controller.DoGeneralRequest(req)
}

// DoGeneralRequests 批量执行门禁请求 （批处理错误操作减少了并发请求，但是会导致错误传播范围过大，对用户不友好，可优化）
func (c *Channel) DoGeneralRequests(
	reqs []db.Request,
) (map[db.IDType]error, error) {
	if c == nil || len(reqs) == 0 || c.controller == nil {
		return nil, errors.New("nil pointer")
	}

	var (
		err          error
		results      = make(map[db.IDType]error, len(reqs))
		skippedError error
	)

	for i := range reqs {
		if skippedError != nil {
			results[reqs[i].ID] = skippedError
			continue
		}

		if _, err = c.controller.DoGeneralRequest(&reqs[i]); err != nil {
			results[reqs[i].ID] = err

			skippedError = fmt.Errorf("other req id: %v, controller: %v, method: %v, payload: %v, "+
				"has error: %w, current request skipped",
				reqs[i].ID, reqs[i].ControllerID, reqs[i].Method, string(reqs[i].Payload), err)
			continue
		}

		results[reqs[i].ID] = nil
		time.Sleep(10 * time.Millisecond)
	}
	return results, nil
}
