// Package door 提供门的查询、更新、状态控制和导入导出功能。
// 包含门信息的增删改查、批量操作、状态控制及导入导出等核心功能。
package door

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"unicode"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/redis"
	"dac/entity/utils"
	"dac/logic/collect/dispatcher"
	"dac/logic/dlm"
	"dac/repo/dac"

	"dac/entity/utils/batch"
	"dac/entity/utils/ghttp"
	"dac/entity/utils/thttp"
	"gorm.io/gorm"
)

// Get 根据ID获取门信息
// 查询成功后会对门数据进行格式化处理
func Get(ctx context.Context, id db.IDType) (db.Door, error) {
	d, err := dac.GetRW().GetDoor(ctx, id)
	if err != nil {
		return d, fmt.Errorf("get door %v error: %w", id, err)
	}
	utils.ProcessDBDoor(&d)
	return d, nil
}

// verifyPassword 校验门密码格式（至少4位纯数字）
// 空密码视为合法，非空密码必须为4位以上纯数字
func verifyPassword(p string) error {
	if len(p) == 0 {
		// 空密码直接通过校验
		return nil
	}
	if len(p) < 4 {
		return fmt.Errorf("password \"%v\" length should be greater than 4", p)
	}
	for _, c := range p {
		if !unicode.IsDigit(c) {
			return fmt.Errorf("invalid password: \"%v\"", p)
		}
	}

	return nil
}

// BatchUpdate 批量更新门参数和名称，支持分布式锁转发
// 当本节点未持有分布式锁时，会将请求转发到持锁节点
func BatchUpdate(ctx context.Context, ids []db.IDType,
	params driver.DoorParameter, names map[db.IDType]string,
	headers http.Header,
) error {
	var req struct {
		IDs    []int                `json:"ids"`
		Params driver.DoorParameter `json:"params"`
		Names  map[int]string       `json:"names"`
	}
	req.IDs = ids
	req.Params = params
	req.Names = names
	// 检查当前节点是否持有分布式锁
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/doors", channelID,
			consts.ServicePort), http.MethodPut, headers, req, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil
	}

	// 校验密码格式
	err := verifyPassword(params.Password)
	if err != nil {
		return err
	}

	// 校验门名称不能为空
	for _, id := range ids {
		if n, ok := names[id]; ok && len(n) == 0 {
			return fmt.Errorf("has empty name: %+v", names)
		}
	}

	// 转换参数格式并构建参数映射
	var doors []db.Door
	p := utils.ConvertDBDoorParameter(&params)
	// 为每个门ID创建独立的参数副本
	paramMap := make(map[db.IDType]*db.DoorParameter, len(ids))
	for _, id := range ids {
		temp := new(db.DoorParameter)
		*temp = p
		paramMap[id] = temp
	}

	e := dac.GetRW().UpdateDoorsNameAndParams(ctx, ids, names, paramMap, func(tx *gorm.DB) error {
		// 需在更新之前查询门数据
		if doors, err = dac.GetDoors(tx, ids); err != nil {
			return err
		}
		for i := range doors {
			d := &doors[i]
			temp, ok := paramMap[d.ID]
			if !ok || temp == nil {
				continue
			}
			if len(temp.Password) == 0 {
				temp.Password = d.Parameters.Password
			}
		}
		return nil
	}, func(tx *gorm.DB) error {
		if len(doors) == 0 {
			return nil
		}

		args := make([]interface{}, 0, len(doors))
		for i := range doors {
			args = append(args, &doors[i])
		}

		return batch.Execute(ctx, args, func(ctx context.Context, i interface{}) error {
			d, ok := i.(*db.Door)
			if !ok {
				return nil
			}

			pp := params

			pp.Name = d.GetName()
			n, ok := names[d.ID]
			if ok && len(n) > 0 {
				pp.Name = n
			}

			if len(pp.Password) == 0 {
				pp.Password = d.Parameters.Password
			}

			pp.Number = driver.DoorNumberType(d.Number)

			b, err := driver.Marshal([]driver.DoorParameter{pp})
			if err != nil {
				return err
			}

			req := db.Request{
				ControllerID: d.ControllerID,
				Method:       driver.MethodSetDoorParameter,
				Payload:      b,
			}
			_, err = dispatcher.Get().DoSyncRequests([]db.Request{req})
			return err
		})
	})
	if e != nil {
		return fmt.Errorf("update doors %v params %+v error: %w", ids, params, e)
	}
	return nil
}

// Update 更新单个门的参数、编号和扩展属性
// 支持分布式锁转发，事务内完成数据库更新和控制器下发
func Update(ctx context.Context, id db.IDType, code string,
	extend map[string]interface{}, params driver.DoorParameter,
	headers http.Header,
) error {
	var req struct {
		ID     int `json:"id"`
		Name   string
		Params driver.DoorParameter   `json:"params"`
		Code   string                 `json:"code"`
		Extend map[string]interface{} `json:"extend"`
	}
	req.ID = id
	req.Code = code
	req.Params = params
	req.Extend = extend

	// 检查当前节点是否持有分布式锁
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/door",
			channelID, consts.ServicePort), http.MethodPut, headers, req, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil

	}

	// 转换参数格式
	var (
		newDoor = new(db.Door)
	)
	p := utils.ConvertDBDoorParameter(&params)
	nameInPlatform := p.Name

	// 标记是否需要下发到控制器
	needSetToController := false

	// 更新门信息（事务内）
	e := dac.GetRW().UpdateDoor(ctx, id, newDoor, func(tx *gorm.DB) error {
		// 获取原始门信息用于对比参数变化
		tempDoor, err := dac.GetDoor(tx, id)
		if err != nil {
			return err
		}

		p.Name = tempDoor.Parameters.Name

		// 密码为空时需要填充原密码
		if len(p.Password) == 0 {
			p.Password = tempDoor.Parameters.Password
		}

		needSetToController = tempDoor.Parameters != p
		config.Log.Infof("door id: %v, code: %v, old door parameters: %+v, new door parameters: %+v, changed: %v",
			id, code, tempDoor.Parameters, p, needSetToController)

		*newDoor = tempDoor
		newDoor.SetIDCDBCode(code)

		if newDoor.Extend == nil {
			newDoor.Extend = make(map[string]interface{})
		}
		for k, v := range extend {
			newDoor.Extend[k] = v
		}

		if len(nameInPlatform) > 0 {
			newDoor.Name = nameInPlatform
		}

		newDoor.Parameters = p

		return nil
	}, func(tx *gorm.DB) error {
		if !needSetToController {
			config.Log.Infof("door %v parameters are not changed, skip set to controller", id)
			return nil
		}

		params.Name = p.Name
		if len(params.Name) == 0 {
			return fmt.Errorf("door name is empty")
		}

		if len(params.Password) == 0 {
			params.Password = newDoor.Parameters.Password
		}

		params.Number = driver.DoorNumberType(newDoor.Number)

		b, err := driver.Marshal([]driver.DoorParameter{params})
		if err != nil {
			return err
		}

		req := db.Request{
			ControllerID: newDoor.ControllerID,
			Method:       driver.MethodSetDoorParameter,
			Payload:      b,
		}
		_, err = dispatcher.Get().DoSyncRequests([]db.Request{req})
		return err
	})
	if e != nil {
		return fmt.Errorf("update door %v code %v, extend %+v, params %+v error: %w", id, code, extend, params, e)
	}

	return nil
}

// SetState 批量设置门状态（开门/关门/常开等）
// 支持分布式锁转发，按控制器分组合并请求
func SetState(ctx context.Context, ids []db.IDType,
	state int, headers http.Header,
) error {
	var req struct {
		IDs   []int `json:"ids"`
		State int   `json:"state"`
	}
	req.IDs = ids
	req.State = state

	// 检查当前节点是否持有分布式锁
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		config.Log.Infof("dlm has no lock, request url: %s", fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/door/state",
			channelID, consts.ServicePort))
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/door/state",
			channelID, consts.ServicePort), http.MethodPost, headers, req, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil
	}
	// 查询门信息
	doors, err := dac.GetRW().GetDoors(ctx, ids)
	if err != nil {
		return fmt.Errorf("get doors error: %w", err)
	}

	// 按控制器ID分组构建请求，同一控制器的门合并为一个请求
	s := driver.DoorStateType(state)

	stateMaps := make(map[db.IDType]driver.SetDoorStateRequest, len(doors))
	reqMaps := make(map[db.IDType]*db.Request, len(doors))
	for i := range doors {
		d := &doors[i]

		controllerID := d.ControllerID
		req, ok := reqMaps[controllerID]
		if !ok {
			req = new(db.Request)
			req.ControllerID = controllerID
			req.Method = driver.MethodSetDoorState

			reqMaps[controllerID] = req
			stateMaps[controllerID] = make(driver.SetDoorStateRequest, 2)
		}
		stateMaps[controllerID][driver.DoorNumberType(d.Number)] = s
	}

	// 序列化请求并发送到控制器
	reqs := make([]db.Request, 0, len(reqMaps))
	// 遍历所有控制器请求，序列化门状态数据
	for _, req := range reqMaps {
		data, ok := stateMaps[req.ControllerID]
		if !ok {
			return fmt.Errorf("not find payload of controller %v", req.ControllerID)
		}
		payload, err := driver.Marshal(data)
		if err != nil {
			return err
		}

		req.Payload = payload

		reqs = append(reqs, *req)
	}

	_, err = dispatcher.Get().DoSyncRequests(reqs)
	if err != nil {
		config.Log.Warnf("do reqs %+v error: %v", reqs, err)
		// 向用户返回友好的错误提示
		return errors.New("发送请求失败")
	}

	return nil
}

// UpdateCode 更新门的IDCDB编号
func UpdateCode(ctx context.Context, id db.IDType, code string) error {
	return dac.GetRW().UpdateDoorCode(ctx, id, code, nil, nil)
}
