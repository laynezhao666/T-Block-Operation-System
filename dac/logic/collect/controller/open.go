// Package controller 提供门禁控制器的采集管理和请求调度功能。
package controller

import (
	"context"
	"dac/entity/config"
	"dac/entity/consts"
	"fmt"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/repo/dac"
)

// addGetDoorParameterRequest 异步添加获取门参数的请求到数据库，支持重试
func (c *Controller) addGetDoorParameterRequest() {
	// 若出现门禁通讯中断等情况需要重试请求
	// 因此将获取门参数的请求写入数据库中
	req := db.Request{
		ControllerID: c.ID(),
		Method:       driver.MethodGetDoorParameter,
		MozuID:       c.MozuID(),
		State:        consts.StateToBeExecuted,
	}
	config.Log.Infof("collect controller add request: %v, id: %v", req, c.ID())
	go func() {
		utils.Retry(100, time.Hour, func() error {
			return dac.GetRW().AddRequests(context.Background(), []db.Request{req})
		}, func() {
			c.Infof(" will get door parameters...")
		}, func(err error) {
			c.Warnf("add get door parameters requests error: %v", err)
		}, func(err error) {
			c.Errorf("add get door parameters requests failed: %v", err)
		})
	}()
}

// addSetTimeGroupRequest 异步添加设置时间组的请求到数据库，支持重试
func (c *Controller) addSetTimeGroupRequest() {
	req := db.Request{
		ControllerID: c.ID(),
		Method:       driver.MethodSetTimeGroup,
		MozuID:       c.MozuID(),
		State:        consts.StateToBeExecuted,
	}

	var (
		groups []driver.TimeGroup
	)

	go func() {
		utils.Retry(100, time.Hour, func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			groupsRecords, err := dac.GetRW().GetAllTimeGroups(ctx)
			if err != nil {
				return fmt.Errorf("get time groups error: %w", err)
			}

			groups = make([]driver.TimeGroup, 0, len(groupsRecords))
			for i := range groupsRecords {
				t, err := utils.ConvertTimeGroupDBToDriver(groupsRecords[i])
				if err != nil {
					continue
				}

				groups = append(groups, t)
			}

			b, err := driver.Marshal(groups)
			if err != nil {
				return fmt.Errorf("json marshal %+v error: %w", groups, err)
			}
			req.Payload = b

			return dac.GetRW().AddRequests(ctx, []db.Request{req})
		}, func() {
			c.Infof("will set time groups %+v", groups)
		}, func(err error) {
			c.Warnf("add set time groups requests error: %v", err)
		}, func(err error) {
			c.Errorf("add set time groups requests failed: %v", err)
		})
	}()
}
