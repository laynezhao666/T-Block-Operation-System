// Package controller 提供门禁控制器的采集管理和请求调度功能。
package controller

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/door"
)

// deleteCard 处理删除卡请求
func (c *Controller) deleteCard(request *db.Request) (interface{}, error) {
	payload := ""
	if err := driver.Unmarshal(&payload, request.Payload); err != nil {
		return nil, err
	}
	return nil, c.controller.DeleteCard(payload)
}

// updateCard 处理更新卡请求（更新卡状态或人员信息）
func (c *Controller) updateCard(req *db.Request) (interface{}, error) {
	var (
		payload driver.Card
		err     error
	)
	if err = driver.Unmarshal(&payload, req.Payload); err != nil {
		return nil, err
	}

	card, err := c.controller.GetCard(payload.CardNo)
	if err != nil {
		return nil, err
	}

	// 等待，防止请求过快导致无响应
	time.Sleep(time.Second)

	switch req.Method {
	case driver.MethodUpdateCardFlag:
		card.CardFlag = payload.CardFlag
		// 在更新卡状态时，只更新CardFlag，不更新DoorNos
		//card.DoorNos = payload.DoorNos
	case driver.MethodUpdateCardStaff:
		card.UserName = payload.UserName
		card.Password = payload.Password
	}
	return nil, c.controller.UpdateCard(card)
}

// DoGeneralRequest 根据请求方法分发并执行通用控制器请求
func (c *Controller) DoGeneralRequest(request *db.Request) (interface{}, error) {
	if c.controller == nil {
		return nil, fmt.Errorf("nil controller")
	}

	var err error
	// 根据请求方法分发到对应的控制器操作
	switch request.Method {
	case driver.MethodGetDoorPositionState:
		return c.controller.GetDoorPositionState()
	case driver.MethodReset:
		return nil, c.controller.Reset()
	case driver.MethodClean:
		return nil, c.controller.Clean()
	case driver.MethodGetAllCards:
		return c.controller.GetAllCards()
	case driver.MethodGetCards:
		var payload int
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetCards(payload)
	case driver.MethodGetCard:
		payload := ""
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetCard(payload)
	case driver.MethodDeleteCard:
		return c.deleteCard(request)
	case driver.MethodSetTime:
		return nil, c.controller.SetTime()
	case driver.MethodGetTimeGroup:
		var payload int
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetTimeGroup(payload)
	case driver.MethodSetTimeGroup:
		// 逐个下发时间组，每次间隔1-2秒防止设备过载
		var payload []driver.TimeGroup
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		for i := range payload {
			if err = c.controller.SetTimeGroup(payload[i]); err != nil {
				return nil, err
			}
			time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(1000)))
		}
		return nil, nil
	case driver.MethodClearTimeGroup:
		var payload int
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.ClearTimeGroup(payload)
	case driver.MethodAddUser:
		var payload driver.CardWithStaffInfo
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.AddUser(payload)
	case driver.MethodDeleteUser:
		var payload driver.UserID
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.DeleteUser(payload)
	case driver.MethodAddCard:
		var payload driver.Card
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.AddCard(payload)
	case driver.MethodUpdateCard:
		var payload driver.Card
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.UpdateCard(payload)
	case driver.MethodUpdateCardFlag, driver.MethodUpdateCardStaff:
		// MDC协议不支持更新卡标志，直接忽略
		if c.record.IsMDC() && request.Method == driver.MethodUpdateCardFlag {
			c.Warnf("not support method: %v, ignore", request.Method)
			return nil, nil
		}
		return c.updateCard(request)
	case driver.MethodGetAlarm:
		var payload int
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetAlarms(payload)
	case driver.MethodGetAlarmByTime:
		var payload driver.TimeInterval
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetAlarmsByTime(payload)
	case driver.MethodGetEvent:
		var payload int
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetEvents(payload)
	case driver.MethodGetEventByTime:
		var payload driver.TimeInterval
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return c.controller.GetEventsByTime(payload)
	case driver.MethodSetDoorState:
		var payload driver.SetDoorStateRequest
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.SetDoorState(payload)
	case driver.MethodSetDoorParameter:
		var payload []driver.DoorParameter
		if err = driver.Unmarshal(&payload, request.Payload); err != nil {
			return nil, err
		}
		return nil, c.controller.SetDoorParameter(payload)
	case driver.MethodGetDoorParameter:
		// 获取门参数后保存到数据库
		params, err := c.controller.GetDoorParameter()
		if err != nil {
			return nil, err
		}
		return nil, door.SaveDoorParameters(context.Background(), c.ID(), params, request.MozuID)
	case driver.MethodGetCurrentAlarm:
		return c.controller.GetCurrentAlarm()
	default:
		break
	}
	return nil, fmt.Errorf("unsupported request method: %v", request.Method)
}
