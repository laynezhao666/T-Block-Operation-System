package xbrother

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/xbrother/consts"
	"dac/repo/dac"
)

// GetTimeGroup 获取指定编号的时间组（从数据库读取并转换为驱动模型）
func (c *Controller) GetTimeGroup(timeGroup int) (driver.TimeGroup, error) {
	driverTimeGroup, err := dac.GetRW().GetDriverTimeGroup(
		context.Background(), c.baseInfo.ID,
		c.chanInfo.ChannelID, timeGroup)
	if err != nil {
		return driver.TimeGroup{}, err
	}
	res := driver.TimeGroup{
		GroupNo:  driverTimeGroup.GroupNo,
		Week:     driverTimeGroup.Week,
		TimeZone: utils.DBTimeZone2DriverTimeZone(driverTimeGroup.TimeZone),
	}
	return res, nil
}

// GetTimeZoneMinuteAndSecond 解析时区字符串"HH:MM"，返回小时和分钟
func GetTimeZoneMinuteAndSecond(input string) (uint8, uint8, error) {
	strs := strings.Split(input, ":")
	if len(strs) != 2 {
		return 0, 0, fmt.Errorf("error timeZone format")
	}
	hour, err := strconv.Atoi(strs[0])
	if err != nil {
		return 0, 0, fmt.Errorf("error timeZone format")
	}
	minute, err := strconv.Atoi(strs[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error timeZone format")
	}
	return uint8(hour), uint8(minute), nil
}

// SetTimeGroup 设置时间组（先下发到门控器，再保存到数据库）
func (c *Controller) SetTimeGroup(timeGroup driver.TimeGroup) error {
	if err := c.SetTimeGroupInController(timeGroup); err != nil {
		return err
	}
	return c.SetTimeGroupInDB(timeGroup)
}

// SetTimeGroupInDB 将时间组保存到数据库
func (c *Controller) SetTimeGroupInDB(timeGroup driver.TimeGroup) error {
	driverTimeGroup := db.DriverTimeGroup{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		GroupNo:      timeGroup.GroupNo,
		Week:         timeGroup.Week,
		TimeZone:     utils.DriverTimeZone2DBTimeZone(timeGroup.TimeZone),
	}
	return dac.GetRW().AddDriverTimeGroup(context.Background(), driverTimeGroup)
}

// SetTimeGroupInController 将时间组下发到门控器（根据门数校验时区号范围，构建协议请求）
func (c *Controller) SetTimeGroupInController(timeGroup driver.TimeGroup) error {
	switch c.doorNum {
	case consts.FourDoorPerController:
		// 协议规定， 四门控制器的时区号为0-7
		if timeGroup.GroupNo < 0 || timeGroup.GroupNo > 7 {
			c.logger.Warnf("unexpected GroupNo: %d, expect:[0-7]", timeGroup.GroupNo)
			return nil
		}
	case consts.OneDoorPerController, consts.TwoDoorPerController:
		// 协议规定，一门和二门控制器的时区号为0-15
		if timeGroup.GroupNo < 0 || timeGroup.GroupNo > 15 {
			return fmt.Errorf("unexpected GroupNo: %d, expect:[0-15]", timeGroup.GroupNo)
		}
	default:
		return fmt.Errorf("unexpected DoorNum: %d", c.doorNum)
	}

	var weekReq uint8 = 0
	isWeekInValid := false
	for _, v := range timeGroup.Week {
		if v < 1 || v > 7 { // Week [1-7] 超过范围为非法
			isWeekInValid = true
			break
		}
		weekReq |= uint8(1) << (v % consts2.DaysOneWeek)
	}
	if isWeekInValid {
		return fmt.Errorf("invalid week, week: %v", timeGroup.Week)
	}

	var openDoorType uint8 = 0
	openDoorType |= consts2.OpenDoorTypeOnlyCard | consts2.OpenDoorTypeCardPass | consts2.OpenDoorTypePass
	for i := range timeGroup.TimeZone {
		startHour, startMinute, err := GetTimeZoneMinuteAndSecond(timeGroup.TimeZone[i].Begin)
		if err != nil {
			return err
		}
		endHour, endMinute, err := GetTimeZoneMinuteAndSecond(timeGroup.TimeZone[i].End)
		if err != nil {
			return err
		}
		req := xbrother.AddTimeGroupReq{
			TimeZone:      uint8(timeGroup.GroupNo),
			StartHour:     startHour,
			StartMinute:   startMinute,
			EndHour:       endHour,
			EndMinute:     endMinute,
			WeekDay:       weekReq,
			OpenDoorType:  openDoorType,
			DeadlineYear:  consts2.TimeGroupDeadlineYear,
			DeadlineMonth: consts2.TimeGroupDeadlineMonth,
			DeadlineDay:   consts2.TimeGroupDeadlineDay,
		}
		// 给控制器控制的门都添加时间组
		if err = c.addTimeGroupsForAllDoors(req); err != nil {
			return err
		}
	}
	return nil
}

// ClearTimeGroup 清除指定时间组（从数据库删除后，清空门控器并重新设置剩余时间组）
func (c *Controller) ClearTimeGroup(timeGroup int) error {
	return dac.GetRW().ClearDriverTimeGroup(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, timeGroup,
		func(dbTimeGroups []db.DriverTimeGroup) error {
			if err := c.clearTimeGroupsForAllDoors(xbrother.ClearDoorTimeGroupsReq{}); err != nil {
				return err
			}
			for i := range dbTimeGroups {
				if err := c.SetTimeGroupInController(driver.TimeGroup{
					GroupNo:  dbTimeGroups[i].GroupNo,
					Week:     dbTimeGroups[i].Week,
					TimeZone: utils.DBTimeZone2DriverTimeZone(dbTimeGroups[i].TimeZone),
				}); err != nil {
					return err
				}
				time.Sleep(consts2.DurationSleepTime)
			}
			return nil
		})
}

// addTimeGroupsForAllDoors 为门控器的所有门添加时间组
func (c *Controller) addTimeGroupsForAllDoors(req xbrother.AddTimeGroupReq) error {
	for i := 0; i < c.doorNum; i++ {
		if _, err := c.addTimeGroups(req, uint8(i+1)); err != nil {
			return err
		}
		time.Sleep(consts2.DurationSleepTime)
	}
	return nil
}

// clearTimeGroupsForAllDoors 清空门控器所有门的时间组
func (c *Controller) clearTimeGroupsForAllDoors(req xbrother.ClearDoorTimeGroupsReq) error {
	for i := 0; i < c.doorNum; i++ {
		if _, err := c.clearDoorTimeGroups(req, uint8(i+1)); err != nil {
			return err
		}
		time.Sleep(consts2.DurationSleepTime)
	}
	return nil
}

// clearDoorTimeGroups 清空指定门的时间组
func (c *Controller) clearDoorTimeGroups(
	req xbrother.ClearDoorTimeGroupsReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo,
		consts2.GetRRPCClearDoorTimeGroups(c.chanInfo.ChannelID),
		consts2.CommandClearTimeGroups)
}

// addTimeGroups 向指定门添加时间组
func (c *Controller) addTimeGroups(req xbrother.AddTimeGroupReq, doorNo uint8) (xbrother.CommonResp, error) {
	time.Sleep(100 * time.Millisecond)
	return c.sendRequest(req, doorNo, consts2.GetRRPCAddTimeGroup(c.chanInfo.ChannelID), consts2.CommandAddTimeGroup)
}
