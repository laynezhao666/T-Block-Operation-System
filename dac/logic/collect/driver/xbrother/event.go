package xbrother

import (
	"context"
	"fmt"
	"strconv"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/utils"
	"dac/logic/collect/driver/xbrother/consts"
	"dac/repo/dac"

	"dac/entity/utils/rrpc"
)

// getEvents 根据偏移量从数据库获取事件列表
func (c *Controller) getEvents(offset any) (driver.EventData, error) {
	var (
		eventData = driver.EventData{}
		err       error
	)
	iOffset, ok := offset.(int)
	if !ok {
		return eventData, fmt.Errorf("unexpected type offset, expect int ")
	}
	totalCount, driverEvents, err := dac.GetRW().GetDriverEvents(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID,
		iOffset, consts.DefaultLimit)
	if err != nil {
		return eventData, err
	}

	eventData.Last = int(totalCount)
	eventData.Events = utils.ConvertDBEventsToDriver(driverEvents)
	eventData.Offset = iOffset + len(driverEvents)
	return eventData, nil
}

// GetEvents 获取事件记录（从数据库读取）
func (c *Controller) GetEvents(offset int) (driver.EventData, error) {
	return c.getEvents(offset)
}

// GetEventsByTime 按时间范围获取事件记录（暂未实现）
func (c *Controller) GetEventsByTime(_ driver.TimeInterval) (driver.EventData, error) {
	// TODO: implement GetEventsByTime for xbrother protocol
	return driver.EventData{}, nil
}

// GetEventsWhenVerify 验证时获取事件（与GetEvents相同逻辑）
func (c *Controller) GetEventsWhenVerify(offset any) (driver.EventData, error) {
	return c.getEvents(offset)
}

// saveEvents 后台协程持续监听事件通道，将接收到的事件保存到数据库
func (c *Controller) saveEvents(ctx context.Context) {
	var err error
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("stop save events.")
			return
		case event, ok := <-c.Server.eventChan:
			if !ok {
				c.logger.Warnf("event channel closed")
				break
			}
			if err = c.saveDriverEventInDB(ctx, event); err != nil {
				c.logger.Errorf("save driver event in db error, err: %v", err)
			}
			rrpc.Manager().Set(consts.GetRRPCSetDriverEvent(c.chanInfo.ChannelID), err)
			c.RedisDoorStatusKeyExpire()
		}
	}
}

// saveEvent 通过事件通道发送事件并等待保存结果
func (s *DoorServer) saveEvent(req xbrother.EventUploadReq) error {
	s.eventChan <- req
	iErr, ok := rrpc.Manager().Get(consts.GetRRPCSetDriverEvent(s.channelID), s.timeout)
	if !ok {
		return fmt.Errorf("rrpc get save Event error")
	}
	if iErr != nil {
		err, ok := iErr.(error)
		if !ok {
			return fmt.Errorf("unexpected rrpc get result type, expect error")
		}
		return err
	}
	return nil
}

// saveDriverEventInDB 将门禁事件请求解析并保存到数据库（去重处理）
func (c *Controller) saveDriverEventInDB(ctx context.Context, req xbrother.EventUploadReq) error {
	timeStamp, err := getTimeStamp(req.Year, req.Month, req.Day, req.Hour, req.Minute, req.Second)
	if err != nil {
		return err
	}
	var userName string
	card, err := c.getCardByCardNo(strconv.Itoa(int(req.CardNo)))
	if err != nil {
		userName = fmt.Sprintf("unknown@%d", req.CardNo)
	} else {
		userName = card.UserName
	}

	desc, ok := consts.EventAlarmDescMap[req.Type]
	if !ok {
		desc = consts.UnknownEventAlarmDesc
	}

	direction := consts.DirectionUnKnown
	switch req.Type {
	case consts.TypeEntryAccess, consts.TypeEntryByCardAndPIN,
		consts.TypeEntryByPIN, consts.TypeEntryByFree:
		direction = consts.DirectionEnter
	case consts.TypeExitAccess, consts.TypeExitByCardAndPIN,
		consts.TypeExitByPIN, consts.TypeExitByFree:
		direction = consts.DirectionExit
	}

	item := db.DriverEvent{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		Timestamp:    timeStamp,
		CardNumber:   strconv.Itoa(int(req.CardNo)),
		Username:     userName,
		DoorNumber:   db.DoorNumberType(req.Door),
		Direction:    db.DirectionType(direction),
		Type:         db.EventType(req.Type),
		Description:  desc,
	}

	// 如果事件已存在，忽略
	if _, err = dac.GetRW().GetDriverEvent(ctx, c.baseInfo.ID, item); err == nil {
		return nil
	}

	// 否则，添加记录到控制器数据库
	return dac.GetRW().SetDriverEvents(ctx, c.baseInfo.ID, []db.DriverEvent{item})
}
