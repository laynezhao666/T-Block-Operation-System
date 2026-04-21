package cacs

import (
	"fmt"
	"strings"

	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/entity/utils"
	"dac/logic/collect/driver/cacs/consts"

	"dac/entity/utils/rrpc"
)

// 时间组删除类型和特殊标识常量
var (
	TypeDeleteOneDay          uint8  = 0    // 删除某天
	TypeDeleteTimeGroup       uint8  = 1    // 删除整个时间组
	TypeDeleteAllTimeGroup    uint8  = 2    // 删除所有时间组
	IdDeleteAllTimeGroup      uint8  = 0xff // 删除所有时间组的ID
	WhatDayDeleteAllTimeGroup uint8  = 0xff // 删除所有天的标识
	TimeZoneSeparator         string = ":"  // 时间段分隔符
)

// GetTimeGroup 从门控器获取指定ID的时间组信息。
// CACS协议只支持8张时间表(0-7)，需逐天查询。
func (c *Controller) GetTimeGroup(timeGroupId int) (driver.TimeGroup, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.TimeGroup{}, err
	}
	// 因为CACS协议只支持8张时间表，序号为(0-7)
	if timeGroupId < 0 || timeGroupId > 7 {
		return driver.TimeGroup{}, fmt.Errorf(
			"timeGroup out of range, timeGroupId = %d, should be (0-7)",
			timeGroupId)
	}
	var res driver.TimeGroup
	res.GroupNo = timeGroupId
	existCount := 0 // 记录实际存在的天数

	// CACS协议只能查询某个时间表星期几的准进时间段，所以遍历每周七天
	for i := 1; i <= 7; i++ {
		resp, ok, packetRtn, _, err := c.getTimeGroups(
			cacs.GetTimeGroupsReq{
				Id:      uint8(timeGroupId),
				WhatDay: uint8(i),
			})
		if !ok {
			return driver.TimeGroup{}, fmt.Errorf(
				"get timegroups failed, err: %s", err.Error())
		}

		// RTN=0x14 表示该星期几的时间表不存在，跳过
		if packetRtn == consts.KRtnTimeGroupNotFound {
			continue
		}
		if packetRtn != consts.KRtnNormal {
			return driver.TimeGroup{}, fmt.Errorf(
				consts.RtnInfoMap[packetRtn])
		}

		existCount++
		res.Week = append(res.Week, int(resp.WhatDay))
		for j := range resp.TimeGroups {
			timeZone := driver.TimeZone{
				Begin: fmt.Sprintf("%d:%d",
					resp.TimeGroups[j].StartHour,
					resp.TimeGroups[j].StartMinute),
				End: fmt.Sprintf("%d:%d",
					resp.TimeGroups[j].EndHour,
					resp.TimeGroups[j].EndMinute),
			}
			res.TimeZone = append(res.TimeZone, timeZone)
		}
	}

	// 如果所有天都不存在，返回错误
	if existCount == 0 {
		return driver.TimeGroup{}, fmt.Errorf(
			"time group %d not exist", timeGroupId)
	}

	return res, nil
}

// TimeInfo 时间信息，包含小时和分钟
type TimeInfo struct {
	Hour   uint8
	Minute uint8
}

// stringToTimeInfo 将 "HH:MM" 格式字符串解析为 TimeInfo。
func stringToTimeInfo(timeZone string) (TimeInfo, error) {
	parts := strings.Split(timeZone, TimeZoneSeparator)
	if len(parts) != 2 {
		return TimeInfo{}, fmt.Errorf(
			"timeZone format error, %s", timeZone)
	}
	hour, err := utils.StringToUint8(parts[0])
	if err != nil {
		return TimeInfo{}, fmt.Errorf(
			"timeGroupNo format error, %s", parts[0])
	}
	minute, err := utils.StringToUint8(parts[1])
	if err != nil {
		return TimeInfo{}, fmt.Errorf(
			"weekDays format error, %s", parts[1])
	}
	return TimeInfo{
		Hour:   hour,
		Minute: minute,
	}, nil
}

// timeZoneToTimeGroup 将驱动层时间段转换为CACS协议时间组。
func timeZoneToTimeGroup(
	timeZone driver.TimeZone,
) (cacs.TimeGroup, error) {
	var resp cacs.TimeGroup
	timeInfo, err := stringToTimeInfo(timeZone.Begin)
	if err != nil {
		return resp, err
	}
	resp.StartHour = timeInfo.Hour
	resp.StartMinute = timeInfo.Minute
	timeInfo, err = stringToTimeInfo(timeZone.End)
	if err != nil {
		return resp, err
	}
	resp.EndHour = timeInfo.Hour
	resp.EndMinute = timeInfo.Minute
	return resp, nil
}

// SetTimeGroup 设置门控器的时间组。
// 设置前先清除整个时间组，确保不会残留旧数据。
func (c *Controller) SetTimeGroup(timeGroup driver.TimeGroup) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	// CACS协议只支持8张时间表(0-7)
	if timeGroup.GroupNo < 0 || timeGroup.GroupNo > 7 {
		return nil
	}

	// 设置前先清除整个时间组
	if err := c.ClearTimeGroup(timeGroup.GroupNo); err != nil {
		c.Warnf("clear time group %d before setting failed: %v",
			timeGroup.GroupNo, err)
	}

	weekDays := timeGroup.Week
	for _, day := range weekDays {
		// CACS协议支持的星期几范围(1-7)
		if day < 1 || day > 7 {
			continue
		}
		timeGroups := make(
			[]cacs.TimeGroup, consts.KTimeGroupPeriodNum)
		timeGroupNum := 0
		for i := range timeGroup.TimeZone {
			// CACS协议支持的准进时间段共有6段
			if timeGroupNum >= consts.KTimeGroupPeriodNum {
				break
			}
			t, err := timeZoneToTimeGroup(timeGroup.TimeZone[i])
			if err != nil {
				continue
			}
			timeGroups[timeGroupNum] = t
			timeGroupNum++
		}

		var reqTimeGroups [6]cacs.TimeGroup
		copy(reqTimeGroups[:], timeGroups)
		_, ok, packetRtn, _, err := c.addTimeGroups(
			cacs.AddTimeGroupsReq{
				Id:         uint8(timeGroup.GroupNo),
				WhatDay:    uint8(day),
				TimeGroups: reqTimeGroups,
			})
		if !ok {
			return fmt.Errorf(
				"SetTimeGroup failed, err: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
	}
	return nil
}

// ClearTimeGroup 清除指定时间组的所有天数据。
func (c *Controller) ClearTimeGroup(timeGroup int) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	_, ok, packetRtn, _, err := c.deleteTimeGroups(
		cacs.DeleteTimeGroupsReq{
			Type:    TypeDeleteTimeGroup,
			Id:      uint8(timeGroup),
			WhatDay: WhatDayDeleteAllTimeGroup,
		})
	if !ok {
		return fmt.Errorf(
			"ClearTimeGroup failed, err: %s", err.Error())
	}
	if packetRtn != consts.KRtnNormal {
		return fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	return nil
}

// addTimeGroups 向门控器添加时间组的底层通信方法。
func (c *Controller) addTimeGroups(
	req cacs.AddTimeGroupsReq,
) (cacs.AddTimeGroupsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.AddTimeGroupsResp{}, false, 0,
			consts.KRequestError, err
	}

	cmd := consts.KCommandRequestAddTimeGroups
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCAddTimeGroups(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseAddTimeGroups, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	addTimeGroupsResp, ok := resp.(cacs.AddTimeGroupsResp)
	if !ok {
		c.Errorf("resp type error, expect AddTimeGroupsResp")
		return cacs.AddTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, expect AddTimeGroupsResp")
	}
	return addTimeGroupsResp, true, server.p.rtn,
		consts.KNormal, nil
}

// getTimeGroups 从门控器查询时间组的底层通信方法。
func (c *Controller) getTimeGroups(
	req cacs.GetTimeGroupsReq,
) (cacs.GetTimeGroupsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.GetTimeGroupsResp{}, false, 0,
			consts.KRequestError, err
	}

	cmd := consts.KCommandRequestGetTimeGroups
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCGetTimeGroups(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	// 检查RTN，非0则直接返回
	if server.p.rtn != consts.KRtnNormal {
		return cacs.GetTimeGroupsResp{}, true, server.p.rtn,
			consts.KNormal, nil
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseGetTimeGroups, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	getTimeGroupsResp, ok := resp.(cacs.GetTimeGroupsResp)
	if !ok {
		c.Errorf("resp type error, expect GetTimeGroupsResp")
		return cacs.GetTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, expect GetTimeGroupsResp")
	}
	return getTimeGroupsResp, true, server.p.rtn,
		consts.KNormal, nil
}

// DeleteAllTimeGroups 删除门控器中的所有时间组。
func (c *Controller) DeleteAllTimeGroups() error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	_, ok, packetRtn, _, err := c.deleteTimeGroups(
		cacs.DeleteTimeGroupsReq{
			Type:    TypeDeleteAllTimeGroup,
			Id:      IdDeleteAllTimeGroup,
			WhatDay: WhatDayDeleteAllTimeGroup,
		})
	if !ok {
		return fmt.Errorf("deleteTimeGroups failed, err: %v", err)
	}
	if packetRtn != consts.KNormal {
		return fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	return nil
}

// deleteTimeGroups 从门控器删除时间组的底层通信方法。
func (c *Controller) deleteTimeGroups(
	req cacs.DeleteTimeGroupsReq,
) (cacs.DeleteTimeGroupsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DeleteTimeGroupsResp{}, false, 0,
			consts.KRequestError, err
	}

	cmd := consts.KCommandRequestDeleteTimeGroups
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCDeleteTimeGroups(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf(consts.RequestInfoMap[consts.KRecvRespError])
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(
		consts.KCommandResponseDeleteTimeGroups, bytes)
	if err != nil {
		c.Errorf("resp unmarshal failed, err: %v", err)
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp unmarshal failed, err: %v", err)
	}
	deleteResp, ok := resp.(cacs.DeleteTimeGroupsResp)
	if !ok {
		c.Errorf("resp type error, expect DeleteTimeGroupsResp")
		return cacs.DeleteTimeGroupsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, expect DeleteTimeGroupsResp")
	}
	return deleteResp, true, server.p.rtn, consts.KNormal, nil
}
