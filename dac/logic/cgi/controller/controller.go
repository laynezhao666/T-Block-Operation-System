// Package controller 提供门禁控制器的增删改查、导入导出和远程控制功能。
package controller

import (
	"context"
	"dac/logic/push"
	"fmt"
	"net/http"
	"trpc.group/trpc-go/trpc-go/log"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/redis"
	"dac/entity/utils"
	"dac/logic/cache"
	"dac/logic/card"
	"dac/logic/collect/dispatcher"
	"dac/logic/dlm"
	"dac/logic/mapping"
	"dac/repo/dac"

	"dac/entity/utils/encoding"
	"dac/entity/utils/ghttp"
	"dac/entity/utils/thttp"
	"dac/entity/utils/ttime"
	"gorm.io/gorm"
)

// GetAll 获取模组对应的所有门禁控制器与门信息
func GetAll(ctx context.Context, mozuID string) ([]cgi.DoorController, error) {
	controllerRecords, doorMap, err := dac.GetRW().GetAllDoorControllersAndDoors(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	controllerIndex := make(map[db.IDType]int, len(controllerRecords))
	for i := range controllerRecords {
		controllerIndex[controllerRecords[i].ID] = i
	}

	controllers := make([]cgi.DoorController, len(controllerRecords))
	for i := range controllers {
		controllers[i].DoorController = controllerRecords[i]
		controllers[i].CommID = utils.GenerateCommID(controllerRecords[i].ID)
		controllers[i].FaultID = utils.GenerateFaultID(controllerRecords[i].ID)

		// 不返回账号密码
		controllers[i].Account = ""
		controllers[i].Password = ""
		// 清除 key
		delete(controllers[i].Extend, consts.KeyProtocolHTTPKey)
	}

	// 关联门禁控制器与门信息
	for controllerID, doors := range doorMap {
		cIndex, ok := controllerIndex[controllerID]
		if !ok {
			continue
		}
		c := &controllers[cIndex]

		c.Doors = make([]cgi.Door, 0, len(doors))
		for i := range doors {
			c.Doors = append(c.Doors, utils.GetCGIDoor(controllerID, doors[i]))
		}
	}

	return controllers, nil
}

// fillKey 根据账号密码生成MD5密钥并存入扩展属性
func fillKey(c *db.DoorController) {
	if c.Extend == nil {
		c.Extend = make(map[string]interface{})
	}
	c.Extend[consts.KeyProtocolHTTPKey] = encoding.MD5String(c.Account + c.Password)
}

// Update 更新门禁控制器信息，合并扩展属性并同步GID
func Update(ctx context.Context, id db.IDType,
	controller db.DoorController, mozuID string,
) error {
	c := &controller
	c.Version = ttime.GetNowUTC().UnixMilli()
	if len(c.Account) > 0 || len(c.Password) > 0 {
		// 若修改了账号或密码，需更新 key
		fillKey(c)
	}

	collectCodeChanged := false

	return dac.GetRW().UpdateDoorController(ctx, id, c, func(tx *gorm.DB) error {
		oldController, err := dac.GetControllerRecord(tx, id)
		if err != nil {
			return err
		}
		extend := oldController.Extend

		// 合并新数据与旧数据的 extend，如同时存在则使用新数据
		for k, v := range extend {
			if _, ok := c.Extend[k]; ok {
				continue
			}

			c.Extend[k] = v
		}

		collectCodeChanged = oldController.GetCollectCode() != controller.GetCollectCode()

		return nil
	}, func(tx *gorm.DB) error {
		if !collectCodeChanged || config.C.IgnoreGID(mozuID) {
			return nil
		}

		gid, err := mapping.GetWorker().FetchGID(controller.GetCollectCode())
		if err != nil {
			return fmt.Errorf("get controller %v gid error: %w", id, err)
		}

		return dac.UpdateControllerGIDByCollectCode(tx, controller.GetCollectCode(), gid)
	})
}

// verifyControllers 校验两组控制器IP是否有重复
func verifyControllers(c1, c2 []db.DoorController) error {
	chIDMap := make(map[string]db.DoorController, len(c1)+len(c2))
	for i := range c1 {
		if _, ok := chIDMap[c1[i].Channel.ID]; ok {
			return fmt.Errorf("ip 重复: %v", c1[i].Channel.ID)
		}
		chIDMap[c1[i].Channel.ID] = c1[i]
	}
	for i := range c2 {
		if _, ok := chIDMap[c2[i].Channel.ID]; ok {
			return fmt.Errorf("ip 重复: %v", c2[i].Channel.ID)
		}
		chIDMap[c2[i].Channel.ID] = c2[i]
	}
	return nil
}

// BatchCreate 批量创建门禁控制器，支持删除旧数据后重建
func BatchCreate(ctx context.Context, mozuID string, controllers []rt.DoorController, deleteOld bool) error {
	controllerNum := len(controllers)

	if controllerNum == 0 {
		return nil
	}

	// 提取DoorController用于校验
	c1 := make([]db.DoorController, 0, controllerNum)
	for i := range controllers {
		c1 = append(c1, controllers[i].DoorController)
	}

	t := ttime.GetNowUTC().UnixMilli()

	collectCodes := make([]string, 0, controllerNum)

	for i := range controllers {
		c := &controllers[i]

		c.Version = t

		c.MozuID = mozuID

		// 默认协议为 HTTP
		if len(c.Protocol.Name) == 0 {
			c.Protocol.Name = consts.ProtocolHTTP
		}
		if len(c.Account) == 0 && len(c.Password) == 0 {
			c.Account = consts.DefaultControllerAccount
			c.Password = consts.DefaultControllerPassword
		}

		switch c.Protocol.Name {
		case consts.ProtocolHTTP:
			fillKey(&c.DoorController)
		}

		for j := range c.Doors {
			c.Doors[j].GroupID = db.DefaultGroupID
		}

		if code := c.GetCollectCode(); len(code) > 0 {
			collectCodes = append(collectCodes, code)
		}
	}

	config.Log.Infof("try add controllers: %+v", controllers)

	return dac.GetRW().AddDoorControllers(ctx, controllers, func(tx *gorm.DB) error {
		// 获取全量
		c2, err := dac.GetAllControllerRecords(tx, "")
		if err != nil {
			return err
		}

		if deleteOld {
			oldIDs := make([]db.IDType, 0, len(c2))
			for i := range c2 {
				if c2[i].MozuID != mozuID {
					// 跳过其它模组门禁控制器
					continue
				}
				oldIDs = append(oldIDs, c2[i].ID)
			}
			if err = dac.DeleteDoorControllers(tx, oldIDs); err != nil {
				return err
			}
			// 清理门禁周期推送测点缓存，防止上报脏数据
			if err = CleanControllerPointsCache(oldIDs); err != nil {
				return err
			}

			c2 = nil
		}

		if err = verifyControllers(c1, c2); err != nil {
			return fmt.Errorf("校验门禁控制器失败: %w", err)
		}
		return nil
	}, func(tx *gorm.DB) error {
		if config.C.IgnoreGID(mozuID) {
			return nil
		}
		codeGIDs, err := mapping.GetWorker().FetchGIDs(collectCodes)
		if err != nil {
			return err
		}
		config.Log.Infof("set code -> gid: %+v", codeGIDs)
		return dac.UpdateControllerAndDoorGIDsByCode(tx, codeGIDs)
	})
}

// CleanControllerPointsCache 清理缓存
func CleanControllerPointsCache(controllerIDs []db.IDType) error {
	worker := push.GetWorker()
	if worker == nil {
		return fmt.Errorf("push worker is nil")
	}

	worker.CleanControllerPointsCache(controllerIDs)
	log.Infof("clean controller points cache: %+v", controllerIDs)
	return nil
}

// BatchDelete 批量删除门禁控制器
func BatchDelete(ctx context.Context, ids []int) error {
	if err := CleanControllerPointsCache(ids); err != nil {
		return err
	}
	return dac.GetRW().DeleteDoorControllers(ctx, ids)
}

// Reset 重置门禁控制器（消防重置）
func Reset(ctx context.Context, id db.IDType, headers http.Header) error {
	// 消防重置请求不写入数据库，直接下发门禁控制器，幂等性操作，无需分布式锁检查
	//var httpReq struct {
	//	ID int `json:"id"`
	//}
	//httpReq.ID = id
	//if !dlm.GetWorker().HasLock() {
	//	channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
	//	if err != nil {
	//		return err
	//	}
	//	var response ghttp.Response
	//	if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/controller/reset",
	//		channelID, consts.ServicePort), http.MethodDelete,
	//		headers, httpReq, consts.HTTPTimeout, &response.Data); err != nil {
	//		return fmt.Errorf("http request failed, err: %s", err.Error())
	//	}
	//	return nil
	//}
	req := &db.Request{ControllerID: id, Method: driver.MethodReset}
	_, err := dispatcher.Get().DoSyncRequest(req)
	return err
}

// allControllersAction 对所有控制器执行指定操作的通用函数
func allControllersAction(ctx context.Context, mozuID string, headers http.Header,
	urlPath string, httpMethod string, driverMethod string, actionName string) error {
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi%s",
			channelID, consts.ServicePort, urlPath), httpMethod, headers, nil, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil
	}

	controllers, err := dac.GetRW().GetAllDoorControllers(ctx, mozuID)
	if err != nil {
		return err
	}

	go func() {
		for i := range controllers {
			c := &controllers[i]
			_, e := dispatcher.Get().DoSyncRequest(&db.Request{
				ControllerID: c.ID,
				Method:       driverMethod,
			})
			if e != nil {
				config.Log.Warnf("%s controller %v error: %v", actionName, c.ID, e)
				return
			}
			config.Log.Infof("%s controller %v success", actionName, c.ID)
		}
	}()

	return nil
}

// AllReset 重置所有门禁控制器
func AllReset(ctx context.Context, mozuID string, headers http.Header) error {
	return allControllersAction(ctx, mozuID, headers,
		"/controllers/reset", http.MethodDelete, driver.MethodReset, "reset")
}

// Clean 门禁控制器格式化，清空后重新下发关联的门禁卡
func Clean(ctx context.Context, id db.IDType) error {
	c, _ := cache.Get().GetController(id)
	reqs := []db.Request{{
		ControllerID: id,
		Method:       driver.MethodClean,
		MozuID:       c.MozuID,
		State:        consts.StateToBeExecuted,
	}}

	var err error
	_, e := dac.GetRW().GetControllerDoors(ctx, id, func(tx *gorm.DB, doors []db.Door) error {
		if err = dac.AddRequests(tx, reqs); err != nil {
			return err
		}

		doorIDs := make([]db.IDType, 0, len(doors))
		for i := range doors {
			doorIDs = append(doorIDs, doors[i].ID)
		}

		accessGroupIDs, err := dac.GetAccessGroupIDsByDoors(tx, doorIDs)
		if err != nil {
			return err
		}

		cardControllerTimeGroups, cardControllerDoors, _, err :=
			dac.GetCardCtrlTimeGroupDoorsByGroups(tx, accessGroupIDs)
		if err != nil {
			return err
		}

		cardNumbers := make([]string, 0, len(cardControllerDoors))

		finalCardControllerTimeGroups, finalCardControllerDoors := make(
			map[string]map[db.IDType]int), make(map[string]map[db.IDType]map[int]struct{})

		for cardNumber, controllerTimeGroups := range cardControllerTimeGroups {
			cardNumbers = append(cardNumbers, cardNumber)

			finalCardControllerTimeGroups[cardNumber] = make(map[db.IDType]int, 1)
			for controllerID, timeGroup := range controllerTimeGroups {
				if controllerID != id {
					continue
				}
				// 只需要下发至当前门禁控制器
				finalCardControllerTimeGroups[cardNumber][controllerID] = timeGroup
			}
		}

		for cardNumber, controllerDoors := range cardControllerDoors {
			finalCardControllerDoors[cardNumber] = make(map[db.IDType]map[int]struct{}, 1)
			for controllerID, ds := range controllerDoors {
				if controllerID != id {
					continue
				}
				// 只需要下发至当前门禁控制器
				finalCardControllerDoors[cardNumber][controllerID] = ds
			}
		}

		return card.AddByControllerTimeGroupAndDoors(tx, cardNumbers, c.MozuID,
			finalCardControllerTimeGroups, finalCardControllerDoors)
	})
	return e
}

// ClearTimeGroups 清除门禁控制器上的所有时间组
func ClearTimeGroups(ctx context.Context, id db.IDType, headers http.Header) error {
	var req struct {
		ID int `json:"id"`
	}
	req.ID = id
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/controller/time-groups",
			channelID, consts.ServicePort), http.MethodDelete, headers, req, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil
	}
	groups, err := dac.GetRW().GetAllTimeGroups(ctx)
	if err != nil {
		return err
	}

	reqs := make([]db.Request, 0, len(groups))
	for i := range groups {
		b, err := driver.Marshal(groups[i].GroupNo)
		if err != nil {
			return err
		}

		reqs = append(reqs, db.Request{
			ControllerID: id,
			Method:       driver.MethodClearTimeGroup,
			Payload:      b,
		})
	}

	_, err = dispatcher.Get().DoSyncRequests(reqs)
	return err
}

// SyncTime 同步门禁控制器时间
func SyncTime(ctx context.Context, id db.IDType, headers http.Header) error {
	var req struct {
		ID int `json:"id"`
	}
	req.ID = id
	if !dlm.GetWorker().HasLock() {
		channelID, err := redis.GetClient().Get(ctx, consts.RedisKeyGetLockChannelIP).Result()
		if err != nil {
			return err
		}
		var response ghttp.Response
		if err := thttp.RequestJSONWithHeader(fmt.Sprintf("http://%s:%d/api/dcos/tdac-cgi/controller/sync-time",
			channelID, consts.ServicePort), http.MethodPost, headers, req, consts.HTTPTimeout, &response.Data); err != nil {
			return fmt.Errorf("http request failed, err: %s", err.Error())
		}
		return nil
	}
	_, err := dispatcher.Get().DoSyncRequest(&db.Request{
		ControllerID: id,
		Method:       driver.MethodSetTime,
	})
	return err
}

// AllSyncTime 同步所有门禁控制器时间
func AllSyncTime(ctx context.Context, mozuID string, headers http.Header) error {
	return allControllersAction(ctx, mozuID, headers,
		"/controllers/sync-time", http.MethodPost, driver.MethodSetTime, "set time")
}

// GetCardFromController 从门控器查询卡是否存在
func GetCardFromController(ctx context.Context, controllerID db.IDType, cardNo string) (driver.Card, bool, error) {
	payload, err := driver.Marshal(cardNo)
	if err != nil {
		return driver.Card{}, false, fmt.Errorf("marshal card no error: %w", err)
	}

	req := &db.Request{
		ControllerID: controllerID,
		Method:       driver.MethodGetCard,
		Payload:      payload,
	}

	resp, err := dispatcher.Get().DoSyncRequest(req)
	if err != nil {
		return driver.Card{}, false, err
	}

	card, ok := resp.(driver.Card)
	if !ok {
		return driver.Card{}, false, fmt.Errorf("response type error, expected driver.Card")
	}

	return card, true, nil
}
