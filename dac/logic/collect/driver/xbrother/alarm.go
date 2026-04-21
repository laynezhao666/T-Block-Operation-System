// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"context"
	"encoding/json"
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

// getAlarmsWithOffset 根据偏移量从数据库获取告警数据
func (c *Controller) getAlarmsWithOffset(offset any) (driver.AlarmData, error) {
	var (
		alarmDatas = driver.AlarmData{}
		err        error
	)
	iOffset, ok := offset.(int)
	if !ok {
		return alarmDatas, fmt.Errorf("unexpected type offset, expect int, offset: %v", offset)
	}
	totalCount, driverAlarms, err := dac.GetRW().GetDriverAlarms(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID,
		iOffset, consts.DefaultLimit)
	if err != nil {
		return alarmDatas, err
	}
	alarmDatas.Last = int(totalCount)
	alarms := make([]driver.Alarm, len(driverAlarms))
	for i := range driverAlarms {
		alarm := &driverAlarms[i]
		alarms[i] = driver.Alarm{
			Index:       alarm.Index,
			Timestamp:   alarm.Timestamp,
			DoorNumber:  driver.DoorNumberType(alarm.DoorNumber),
			Type:        driver.AlarmType(alarm.Type),
			State:       driver.AlarmStateType(alarm.State),
			Description: alarm.Description,
		}
	}
	alarmDatas.Alarms = alarms
	alarmDatas.Offset = iOffset + len(driverAlarms)
	return alarmDatas, nil
}

// GetAlarms 根据偏移量获取告警数据
func (c *Controller) GetAlarms(offset int) (driver.AlarmData, error) {
	return c.getAlarmsWithOffset(offset)
}

// GetAlarmsByTime 按时间区间获取告警（XBrother暂未实现）
func (c *Controller) GetAlarmsByTime(_ driver.TimeInterval) (driver.AlarmData, error) {
	// TODO: implement GetAlarmsByTime for xbrother protocol
	return driver.AlarmData{}, nil
}

// GetAlarmsWhenVerify 校验模式下获取告警数据
func (c *Controller) GetAlarmsWhenVerify(offset any) (driver.AlarmData, error) {
	return c.getAlarmsWithOffset(offset)
}

// GetCurrentAlarm 从Redis获取当前活跃的告警列表
func (c *Controller) GetCurrentAlarm() ([]driver.CurrentAlarmData, error) {
	currentAlarmMap, err := c.redisClient.HGetAll(context.Background(),
		utils.GenerateRedisKeyDoorOpenTimeout(c.chanInfo.ChannelID)).Result()
	if err != nil {
		return nil, err
	}
	data := make([]driver.CurrentAlarmData, 0)
	for doorNoStr, v := range currentAlarmMap {
		var driverCurrentAlarmData driver.CurrentAlarmData
		doorNo, err := strconv.Atoi(doorNoStr)
		if err != nil {
			return nil, err
		}
		driverCurrentAlarmData.Door = doorNo
		var currentAlarmEvent driver.CurrentAlarmEvent
		if err := json.Unmarshal([]byte(v), &currentAlarmEvent); err != nil {
			return nil, err
		}
		driverCurrentAlarmData.Alarms = []driver.CurrentAlarmEvent{currentAlarmEvent}
		data = append(data, driverCurrentAlarmData)
	}
	return data, nil
}

// setAlarm 发送设置告警命令到控制器
func (c *Controller) setAlarm(req xbrother.AlarmSettingReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCSetAlarm(c.chanInfo.ChannelID), consts.CommandSetAlarm)
}

// setFireAlarm 发送设置火警命令到控制器
func (c *Controller) setFireAlarm(req xbrother.AlarmSettingReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts.GetRRPCSetFireAlarm(c.chanInfo.ChannelID), consts.CommandSetFireAlarm)
}

// saveAlarms 持续监听告警通道并保存到数据库
func (c *Controller) saveAlarms(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("stop save alarms")
			return
		case alarm, ok := <-c.Server.alarmChan:
			if !ok {
				c.logger.Errorf("alarm channel close")
				break
			}
			err := c.saveDriverAlarmInDB(ctx, alarm)
			if err != nil {
				c.logger.Errorf("save driver alarm in db error, err: %s", err.Error())
			}
			rrpc.Manager().Set(consts.GetRRPCSetDriverAlarm(c.chanInfo.ChannelID), err)
			c.RedisDoorStatusKeyExpire()
		}
	}
}

// saveDriverAlarmInDB 将驱动层告警数据保存到数据库
func (c *Controller) saveDriverAlarmInDB(ctx context.Context, req xbrother.AlarmUploadReq) error {
	timeStamp, err := getTimeStamp(req.Year, req.Month, req.Day, req.Hour, req.Minute, req.Second)
	if err != nil {
		return err
	}
	desc, ok := consts.EventAlarmDescMap[req.Type]
	if !ok {
		desc = consts.UnknownEventAlarmDesc
	}
	item := db.DriverAlarm{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		Timestamp:    timeStamp,
		DoorNumber:   db.DoorNumberType(req.Door),
		Type:         db.AlarmType(req.Type),
		State:        0, // 没有
		Description:  desc,
	}

	// 目前只记录门开超时告警
	if req.Type == consts.TypeDoorOpenTooLong {
		event := driver.CurrentAlarmEvent{
			Type: int(req.Type),
			Desc: desc,
		}
		eventJson, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if err = c.redisClient.HSet(context.Background(), utils.GenerateRedisKeyDoorOpenTimeout(c.chanInfo.ChannelID),
			fmt.Sprintf("%d", req.Door), eventJson).Err(); err != nil {
			return err
		}
	}

	_, err = dac.GetRW().GetDriverAlarm(ctx, c.baseInfo.ID, item)
	if err != nil {
		return dac.GetRW().SetDriverAlarms(ctx, c.baseInfo.ID, []db.DriverAlarm{item})
	}
	return nil
}
