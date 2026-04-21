// Package dispatcher 提供门禁控制器的请求分发和管理功能。
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/logic/collect/worker"

	"dac/entity/utils/batch"
)

// channelMap 控制器ID到工作通道的映射
type channelMap map[db.IDType]*worker.Channel

// d 全局分发器单例
var (
	d = Dispatcher{
		channels: make(channelMap),
	}
)

// Dispatcher 请求分发器，管理所有控制器的工作通道
type Dispatcher struct {
	sync.RWMutex
	channels channelMap
}

// Get 获取全局分发器实例
func Get() *Dispatcher {
	return &d
}

// AddControllers 批量添加控制器到分发器
func (d *Dispatcher) AddControllers(
	ctx context.Context, controllers []rt.DoorController,
) {
	if len(controllers) == 0 {
		return
	}

	for i := range controllers {
		d.addController(ctx, controllers[i])
	}
}

// addController 添加单个控制器到分发器
func (d *Dispatcher) addController(
	ctx context.Context, controller rt.DoorController,
) {
	d.Lock()
	defer d.Unlock()

	channelID := controller.Channel.ID
	c, ok := d.channels[controller.ID]
	if !ok {
		c = worker.NewChannel(channelID)
		d.channels[controller.ID] = c
	}

	config.Log.Infof("try start controller: %+v", controller)
	c.Start(ctx, controller)
}

// DeleteControllers 批量删除控制器
func (d *Dispatcher) DeleteControllers(ids []db.IDType) {
	if len(ids) == 0 {
		return
	}

	for _, id := range ids {
		d.deleteController(id)
	}
}

// deleteController 删除单个控制器
func (d *Dispatcher) deleteController(id db.IDType) {
	d.Lock()
	defer d.Unlock()

	c, ok := d.channels[id]
	if ok && c != nil {
		c.Stop()
	}

	delete(d.channels, id)
}

// mergeRequests 将请求按照controllerId分组，并按createTime排序
func mergeRequests(
	reqs []db.Request,
) map[db.IDType][]db.Request {
	controllerReqs := make(map[db.IDType][]db.Request)
	for i := range reqs {
		r := &reqs[i]
		controllerReqs[r.ControllerID] = append(
			controllerReqs[r.ControllerID], reqs[i])
	}
	newControllerReqs := make(
		map[db.IDType][]db.Request, len(controllerReqs))
	for id, reqs := range controllerReqs {
		sort.Slice(reqs, func(i, j int) bool {
			return reqs[i].CreateTime < reqs[j].CreateTime
		})
		newControllerReqs[id] = reqs
	}
	return newControllerReqs
}

// DoAsyncRequests 批量异步处理请求，返回每个请求的执行结果
func (d *Dispatcher) DoAsyncRequests(
	reqs []db.Request,
) map[db.IDType]error {
	var (
		resultMutex sync.Mutex
		results     = make(map[db.IDType]error, len(reqs))
	)

	controllerReqs := mergeRequests(reqs)

	args := make([]interface{}, 0, len(controllerReqs))
	for c := range controllerReqs {
		args = append(args, c)
	}

	_ = batch.ExecuteAll(context.Background(), args,
		func(_ context.Context, arg interface{}) error {
			controllerID, ok := (arg).(db.IDType)
			if !ok {
				return nil
			}

			d.RLock()
			c, ok := d.channels[controllerID]
			d.RUnlock()

			if !ok || c == nil {
				err := fmt.Errorf(
					"channel of controller %v not found",
					controllerID)
				config.Log.Warnf("%v", err)
				resultMutex.Lock()
				for i := range controllerReqs[controllerID] {
					results[controllerReqs[controllerID][i].ID] = err
				}
				resultMutex.Unlock()
				return nil
			}

			toDoReqs := controllerReqs[controllerID]
			reqResults, err := c.DoGeneralRequests(toDoReqs)

			resultMutex.Lock()
			defer resultMutex.Unlock()

			if err != nil {
				for i := range toDoReqs {
					results[toDoReqs[i].ID] = err
				}
			} else {
				for reqID, e := range reqResults {
					results[reqID] = e
				}
			}

			return nil
		})

	return results
}

// getArgs 将不同类型的请求集合转换为统一的接口切片
func getArgs(reqs interface{}) ([]interface{}, error) {
	args := make([]interface{}, 0)
	switch t := reqs.(type) {
	case []db.Request:
		for i := range t {
			args = append(args, &t[i])
		}
	case []*db.Request:
		for i := range t {
			args = append(args, t[i])
		}
	case map[db.IDType]*db.Request:
		for _, r := range t {
			args = append(args, r)
		}
	default:
		return nil, errors.New("error requests type")
	}
	return args, nil
}

// DoSyncRequest 同步执行单个请求并返回响应
func (d *Dispatcher) DoSyncRequest(
	r *db.Request,
) (interface{}, error) {
	if r == nil {
		return nil, errors.New("nil pointer")
	}

	d.RLock()
	c, ok := d.channels[r.ControllerID]
	d.RUnlock()

	if !ok || c == nil {
		return nil, fmt.Errorf(
			"control %v not found", r.ControllerID)
	}

	resp, err := c.DoGeneralRequest(r)
	if err != nil {
		return nil, fmt.Errorf(
			"do request %+v, payload: %v, error: %w",
			*r, string(r.Payload), err)
	}

	return resp, nil
}

// DoSyncRequests 批量同步执行请求并返回所有响应
func (d *Dispatcher) DoSyncRequests(
	reqs interface{},
) (map[db.IDType]interface{}, error) {
	args, err := getArgs(reqs)
	if err != nil {
		return nil, err
	}
	l := len(args)

	var (
		resultMutex sync.Mutex
		results     = make(map[db.IDType]interface{}, l)
	)

	err = batch.Execute(context.Background(), args,
		func(_ context.Context, arg interface{}) error {
			r, ok := (arg).(*db.Request)
			if !ok {
				return errors.New("type assertion failed")
			}

			d.RLock()
			c, ok := d.channels[r.ControllerID]
			d.RUnlock()

			if !ok || c == nil {
				return fmt.Errorf(
					"control %v not found", r.ControllerID)
			}

			resp, err := c.DoGeneralRequest(r)
			if err != nil {
				return fmt.Errorf(
					"do request %+v, payload: %v, error: %w",
					*r, string(r.Payload), err)
			}

			resultMutex.Lock()
			defer resultMutex.Unlock()

			results[r.ID] = resp

			return nil
		})

	return results, err
}
