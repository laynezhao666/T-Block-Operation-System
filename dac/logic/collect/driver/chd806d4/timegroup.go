// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"dac/repo/dac"
)

// ============ 时间组管理 ============
//
// CHD 协议时间组结构说明：
// 1. 时段表（32张，序号0-31）：每张表定义4个时间段，每段为 HH:MM-HH:MM
// 2. 星期时段索引：每个门7天×8个索引（第1-4类卡、门常开、刷卡加密码、自动布防、N+1屏蔽）

// ============ 上层业务接口实现 ============

// GetTimeGroup 获取时间组（从数据库读取）
func (c *Controller) GetTimeGroup(timeGroupNo int) (driver.TimeGroup, error) {
	driverTimeGroup, err := dac.GetRW().GetDriverTimeGroup(context.Background(),
		c.baseInfo.ID, c.chanInfo.ChannelID, timeGroupNo)
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

// SetTimeGroup 设置时间组（适配上层业务）
// 1. 先设置到门控器
// 2. 再保存到数据库
func (c *Controller) SetTimeGroup(timeGroup driver.TimeGroup) error {
	// 先设置到门控器
	if err := c.SetTimeGroupInController(timeGroup); err != nil {
		return err
	}
	//// 再保存到数据库
	//return c.SetTimeGroupInDB(timeGroup)
	return nil
}

// SetTimeGroupInDB 保存时间组到数据库
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

// SetTimeGroupInController 设置时间组到门控器
// 实现逻辑：
// 1. 设置时段表（将 TimeZone 写入指定的时段表序号，需要对所有门设置）
// 2. 设置星期时段索引（将 Week 对应的天指向该时段表，需要对所有门设置）
func (c *Controller) SetTimeGroupInController(timeGroup driver.TimeGroup) error {
	if err := c.checkAuth(); err != nil {
		return err
	}

	// 检查时间组序号范围（0-31）
	if timeGroup.GroupNo < 0 || timeGroup.GroupNo > 31 {
		return fmt.Errorf("时间组序号超出范围(0-31): %d", timeGroup.GroupNo)
	}

	// 检查 Week 有效性（1-7）
	for _, w := range timeGroup.Week {
		if w < 1 || w > 7 {
			return fmt.Errorf("星期值超出范围(1-7): %d", w)
		}
	}

	// 步骤1：设置时段表（时段表是全局的，只需设置一次）
	if err := c.setTimeSlotTable(timeGroup.GroupNo, timeGroup.TimeZone); err != nil {
		return fmt.Errorf("设置时段表失败: %w", err)
	}
	fmt.Printf("设置时段表成功✅, groupNo: %d, timezone: %v\n", timeGroup.GroupNo, timeGroup.TimeZone)
	time.Sleep(100 * time.Millisecond)

	// 步骤2：设置星期时段索引（使用门号 0xFF 一次性设置所有门）
	if err := c.setWeekTimeIndexForDays(0xFF, timeGroup.Week, byte(timeGroup.GroupNo)); err != nil {
		return fmt.Errorf("设置所有门星期时段索引失败: %w", err)
	}
	fmt.Printf("设置所有门星期时段索引成功✅, groupNo: %d, week: %v\n", timeGroup.GroupNo, timeGroup.Week)

	return nil
}

// ClearTimeGroup 清除时间组
// 1. 从数据库删除
// 2. 重新设置门控器（清除后重新设置剩余的时间组）
func (c *Controller) ClearTimeGroup(timeGroupNo int) error {
	return dac.GetRW().ClearDriverTimeGroup(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, timeGroupNo,
		func(dbTimeGroups []db.DriverTimeGroup) error {
			// 先清空门控器上该时间组对应的时段表
			if err := c.clearTimeSlotTable(timeGroupNo); err != nil {
				return err
			}
			// 重新设置剩余的时间组
			for i := range dbTimeGroups {
				if err := c.SetTimeGroupInController(driver.TimeGroup{
					GroupNo:  dbTimeGroups[i].GroupNo,
					Week:     dbTimeGroups[i].Week,
					TimeZone: utils.DBTimeZone2DriverTimeZone(dbTimeGroups[i].TimeZone),
				}); err != nil {
					return err
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})
}

// ============ 协议底层接口 ============

// setTimeSlotTable 设置时段表（协议 4.2.1）
// DATAF 共 17 字节：表序号(1字节,0-31) + 4组时段(16字节,HH:MM-HH:MM×4)
// 注意：时段表是全局的（所有门共享 32 张时段表），不需要指定门号
// tableNo: 时段表序号（0-31）
func (c *Controller) setTimeSlotTable(tableNo int, timeZones []driver.TimeZone) error {
	// 构建 DATAF（17字节）- 协议 4.2.1
	dataf := make([]byte, 17)
	dataf[0] = byte(tableNo) // 表序号（0-31）

	// 填充4组时段，每组4字节（HH:MM-HH:MM），使用二进制编码
	for i := 0; i < 4; i++ {
		offset := 1 + i*4 // 从第2字节开始
		if i < len(timeZones) {
			tz := timeZones[i]
			// 解析起始时间
			beginH, beginM, err := parseTimeString(tz.Begin)
			if err != nil {
				return fmt.Errorf("解析起始时间失败: %w", err)
			}
			dataf[offset] = beginH
			dataf[offset+1] = beginM

			// 解析结束时间
			endH, endM, err := parseTimeString(tz.End)
			if err != nil {
				return fmt.Errorf("解析结束时间失败: %w", err)
			}
			dataf[offset+2] = endH
			dataf[offset+3] = endM
		} else {
			// 未使用的时段设为 00:00-00:00
			dataf[offset] = 0x00
			dataf[offset+1] = 0x00
			dataf[offset+2] = 0x00
			dataf[offset+3] = 0x00
		}
	}

	// 发送设置命令
	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeGroupSet,
		dataf,
		consts2.GetRRPCTimeGroupSet,
	)
	return err
}

// clearTimeSlotTable 清空时段表（设置为全0）
// 时段表是全局共享的，只需清除一次
func (c *Controller) clearTimeSlotTable(tableNo int) error {
	return c.setTimeSlotTable(tableNo, []driver.TimeZone{})
}

// GetTimeSlotTable 读取时段表（协议 5.3.1）
// COMMAND TYPE=0X89
// DATAF：1字节，表序号（0-31）
// SM 返回：DATAINFO 共 16 字节（4组时段：HH:MM-HH:MM × 4）
// tableNo: 时段表序号（0-31）
func (c *Controller) GetTimeSlotTable(tableNo int) ([]driver.TimeZone, error) {
	// 构建 DATAF（1字节）- 协议5.3.1要求只发送表序号
	dataf := []byte{
		byte(tableNo), // 表序号（0-31）
	}

	// 发送读取命令
	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeTimeGroupGet,
		dataf,
		consts2.GetRRPCTimeGroupGet,
	)
	if err != nil {
		return nil, err
	}

	// 解析返回数据（16字节时段数据）
	if len(respInfo) < 16 {
		return nil, fmt.Errorf("返回数据长度不足: 期望16字节，实际%d字节", len(respInfo))
	}

	timeZones := make([]driver.TimeZone, 0)
	// 解析4组时段（每组4字节：HH:MM-HH:MM）
	for i := 0; i < 4; i++ {
		offset := i * 4
		beginH := int(respInfo[offset])
		beginM := int(respInfo[offset+1])
		endH := int(respInfo[offset+2])
		endM := int(respInfo[offset+3])

		// 跳过未使用的时段（00:00-00:00）
		if beginH == 0 && beginM == 0 && endH == 0 && endM == 0 {
			continue
		}

		tz := driver.TimeZone{
			Begin: fmt.Sprintf("%02d:%02d", beginH, beginM),
			End:   fmt.Sprintf("%02d:%02d", endH, endM),
		}
		timeZones = append(timeZones, tz)
	}

	return timeZones, nil
}

// setWeekTimeIndexForDays 为指定的星期设置时段索引
// doorNo: 门号（1-32，或 0xFF 表示所有门）
// weekdays: 星期列表（1-7，7=星期日）
// tableIndex: 时段表索引（0-31）
// 注意：只有指定的星期会设置时段索引，未指定的星期索引设为0（不限制）
func (c *Controller) setWeekTimeIndexForDays(doorNo int, weekdays []int, tableIndex byte) error {
	// 初始化7天的索引，全部设为0（不限制）
	var weekIndexes [7][8]byte

	// 只设置指定星期的时段索引
	for _, weekday := range weekdays {
		if weekday < 1 || weekday > 7 {
			continue
		}
		// 设置第1-4类卡的准进时段索引都指向该时段表
		weekIndexes[weekday-1][0] = tableIndex // 第1类卡
		weekIndexes[weekday-1][1] = tableIndex // 第2类卡
		weekIndexes[weekday-1][2] = tableIndex // 第3类卡
		weekIndexes[weekday-1][3] = tableIndex // 第4类卡
		// 其他索引（门常开、刷卡加密码、自动布防、N+1屏蔽）保持为0
	}

	// 设置到门控器
	return c.SetWeekTimeIndex(doorNo, weekIndexes)
}

// ============ 星期时段索引操作 ============

// SetWeekTimeIndex 设置星期时段索引（协议 4.2.2）
// DATAF 共 58 字节：门号(1字节) + 0(1字节,标识) + 7天×8字节索引表(56字节)
// doorNo: 门号（1-32，或 0xFF 表示所有门同时设置）
func (c *Controller) SetWeekTimeIndex(doorNo int, weekIndexes [7][8]byte) error {
	if err := c.checkAuth(); err != nil {
		return err
	}

	// 检查门号范围（1-32，或 0xFF 表示所有门）
	if doorNo != 0xFF && (doorNo < 1 || doorNo > 32) {
		return fmt.Errorf("门号超出范围(1-32或0xFF): %d", doorNo)
	}

	// 构建 DATAF（58字节）：门号 + 标识0 + 56字节索引表
	dataf := make([]byte, 58)
	dataf[0] = byte(doorNo)
	dataf[1] = 0x00 // 标识：设置星期时段索引

	// 填充7天×8字节的索引表（共56字节）
	for day := 0; day < 7; day++ {
		for idx := 0; idx < 8; idx++ {
			dataf[2+day*8+idx] = weekIndexes[day][idx]
		}
	}

	// 发送设置命令
	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeGroupWeekSet,
		dataf,
		consts2.GetRRPCTimeGroupWeekSet,
	)
	return err
}

// GetWeekTimeIndex 读取星期时段索引（协议 5.3.2）
// DATAF：门号(1字节) + 0(1字节) + 星期(1字节,1-7)
// SM 返回：DATAINFO 10 字节
//   - 第1字节：0（固定值）
//   - 第2字节：星期值(1-7)
//   - 第3-10字节：8字节时段索引表
func (c *Controller) GetWeekTimeIndex(doorNo int, weekday int) ([8]byte, error) {
	if err := c.checkAuth(); err != nil {
		return [8]byte{}, err
	}

	// 检查参数范围
	if doorNo < 1 || doorNo > 32 {
		return [8]byte{}, fmt.Errorf("门号超出范围(1-32): %d", doorNo)
	}
	if weekday < 1 || weekday > 7 {
		return [8]byte{}, fmt.Errorf("星期超出范围(1-7): %d", weekday)
	}

	// 构建 DATAF（3字节）
	dataf := []byte{
		byte(doorNo),
		0x00,          // 标识：读取星期时段索引
		byte(weekday), // 星期（1-7）
	}

	// 发送读取命令
	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeTimeGroupWeekGet,
		dataf,
		consts2.GetRRPCTimeGroupWeekGet,
	)
	if err != nil {
		return [8]byte{}, err
	}

	// 解析返回数据（协议5.3.2：DATAINFO 10字节）
	// 第1字节：0（固定值），第2字节：星期值，第3-10字节：8字节索引表
	if len(respInfo) < 10 {
		return [8]byte{}, fmt.Errorf("返回数据长度不足: 期望10字节，实际%d字节", len(respInfo))
	}

	var indexes [8]byte
	copy(indexes[:], respInfo[2:10]) // 从第3字节开始（索引2）读取8字节
	return indexes, nil
}

// ============ 特殊日期（节假日）时段索引操作 ============

// SetHolidayTimeIndex 设置特殊日期时段索引（协议 4.2.3）
func (c *Controller) SetHolidayTimeIndex(doorNo int, month, day int, indexes [8]byte) error {
	if err := c.checkAuth(); err != nil {
		return err
	}

	if doorNo < 1 || doorNo > 32 {
		return fmt.Errorf("门号超出范围(1-32): %d", doorNo)
	}

	// 构建 DATAF（12字节）
	dataf := make([]byte, 12)
	dataf[0] = byte(doorNo)
	dataf[1] = 0x01 // 标识：设置特殊日期
	dataf[2] = byte(month)
	dataf[3] = byte(day)
	copy(dataf[4:12], indexes[:])

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeGroupHolidaySet,
		dataf,
		consts2.GetRRPCTimeGroupHolidaySet,
	)
	return err
}

// DeleteHolidayTimeIndex 删除一组特殊日期时段索引（协议 4.2.4）
func (c *Controller) DeleteHolidayTimeIndex(doorNo int, month, day int) error {
	if err := c.checkAuth(); err != nil {
		return err
	}

	dataf := []byte{
		byte(doorNo),
		0x02, // 标识：删除特殊日期
		byte(month),
		byte(day),
	}

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeGroupHolidayDel,
		dataf,
		consts2.GetRRPCTimeGroupHolidayDel,
	)
	return err
}

// ClearAllHolidayTimeIndex 清空全部特殊日期时段索引（协议 4.2.5）
func (c *Controller) ClearAllHolidayTimeIndex(doorNo int) error {
	if err := c.checkAuth(); err != nil {
		return err
	}

	dataf := []byte{
		byte(doorNo),
		0x03, // 标识：清空全部特殊日期
		0x00,
		0x00,
	}

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeGroupHolidayClr,
		dataf,
		consts2.GetRRPCTimeGroupHolidayClr,
	)
	return err
}

// GetHolidayTimeIndexBySeq 按序列读取特殊日期时段索引（协议 5.3.3）
func (c *Controller) GetHolidayTimeIndexBySeq(doorNo int, seq int) (month, day int, indexes [8]byte, err error) {
	if authErr := c.checkAuth(); authErr != nil {
		return 0, 0, [8]byte{}, authErr
	}

	if doorNo < 1 || doorNo > 32 {
		return 0, 0, [8]byte{}, fmt.Errorf("门号超出范围(1-32): %d", doorNo)
	}
	if seq < 1 || seq > 44 {
		return 0, 0, [8]byte{}, fmt.Errorf("序号超出范围(1-44): %d", seq)
	}

	dataf := []byte{
		byte(doorNo),
		byte(seq),
		0x00,
	}

	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeTimeGroupHolidayGet,
		dataf,
		consts2.GetRRPCTimeGroupHolidayGet,
	)
	if err != nil {
		return 0, 0, [8]byte{}, err
	}

	if len(respInfo) < 11 {
		return 0, 0, [8]byte{}, fmt.Errorf("返回数据长度不足: 期望11字节，实际%d字节", len(respInfo))
	}

	month = int(respInfo[1])
	day = int(respInfo[2])
	copy(indexes[:], respInfo[3:11])
	return month, day, indexes, nil
}

// GetHolidayTimeIndexByDate 按日期读取特殊日期时段索引（协议 5.3.4）
func (c *Controller) GetHolidayTimeIndexByDate(doorNo int, month, day int) ([8]byte, error) {
	if err := c.checkAuth(); err != nil {
		return [8]byte{}, err
	}

	// 十进制转BCD
	decToBCD := func(dec int) byte {
		return byte((dec/10)<<4 | (dec % 10))
	}

	dataf := []byte{
		byte(doorNo),
		decToBCD(month),
		decToBCD(day),
	}

	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeTimeGroupHolidayByDay,
		dataf,
		consts2.GetRRPCTimeGroupHolidayGet,
	)
	if err != nil {
		return [8]byte{}, err
	}

	if len(respInfo) < 11 {
		return [8]byte{}, fmt.Errorf("返回数据长度不足: 期望11字节，实际%d字节", len(respInfo))
	}

	var indexes [8]byte
	copy(indexes[:], respInfo[3:11])
	return indexes, nil
}

// ============ 辅助函数 ============

// parseTimeString 解析时间字符串 "HH:MM"
func parseTimeString(timeStr string) (hour, minute byte, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("时间格式错误: %s", timeStr)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("解析小时失败: %w", err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("解析分钟失败: %w", err)
	}
	return byte(h), byte(m), nil
}
