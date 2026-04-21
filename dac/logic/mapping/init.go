// Package mapping 提供采集编码到全局唯一标识符(GID)的映射管理。
package mapping

import (
	"context"

	"dac/entity/model/rt"
	"dac/repo/dac"
)

// Init 初始化映射模块，从数据库加载所有编码到GID的映射
func Init(ctx context.Context) error {
	var err error
	if err = GetWorker().init(ctx); err != nil {
		return err
	}

	return nil
}

// init 从数据库加载控制器和门的GID映射，并启动通知拉取协程
func (w *worker) init(ctx context.Context) error {
	go w.notifyFetchLoop(ctx)

	controllers, controllerDoors, err :=
		dac.GetRW().GetAllDoorControllersAndDoors(ctx, "")
	if err != nil {
		return err
	}

	codeGIDs := make(rt.CodeGIDMapType)

	// 遍历所有控制器，建立编码到GID的映射
	for i := range controllers {
		c := &controllers[i]

		controllerCode := c.GetCollectCode()

		if len(c.GID) == 0 {
			continue
		}
		codeGIDs[controllerCode] = c.GID

		doors, ok := controllerDoors[c.ID]
		if !ok {
			continue
		}
		// 遍历控制器下的所有门
		for j := range doors {
			d := &doors[j]

			if len(d.GID) == 0 {
				continue
			}

			codeGIDs[d.GetCollectCode(controllerCode)] = d.GID
		}
	}

	w.setGIDs(codeGIDs)
	return nil
}
