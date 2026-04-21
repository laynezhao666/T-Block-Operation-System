// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/redis"
	"dac/entity/utils"
	"dac/repo/dac"
)

// ============ 告警类型定义 ============

// 报警源 AS2, AS1 定义（REMARK=5 时的事件来源）
const (
	// AS2=0 的情况
	AlarmIRStart         = 0x0000 // 红外报警开始
	AlarmIRStop          = 0x0001 // 红外停止报警
	AlarmDoorForceOpen   = 0x0002 // 非正常开门
	AlarmDoorClose       = 0x0003 // 门被关闭（非正常开门对应）
	AlarmIRMonitorOff    = 0x0007 // 入侵监测被关闭
	AlarmIRMonitorOn     = 0x0008 // 入侵监测开启
	AlarmDoorMonitorOff  = 0x0009 // 门碰开关监测被关闭
	AlarmDoorMonitorOn   = 0x000A // 门碰开关监测开启
	AlarmExternalRemoved = 0x000B // 外部设备被拆除
	AlarmDeviceRemoved   = 0x000C // 本机被拆除
	AlarmNoEntryInTime   = 0x0030 // 合法刷卡未在延时内开门进入
	AlarmDoorOpenTooLong = 0x0032 // 进入后未关好门
	AlarmEmergencyStart  = 0x0022 // 紧急输入开始
	AlarmEmergencyEnd    = 0x0122 // 紧急输入结束（AS2=1, AS1=0x22）
)

// 告警类型描述映射
var alarmDescMap = map[uint16]string{
	AlarmIRStart:         "红外报警开始",
	AlarmIRStop:          "红外停止报警",
	AlarmDoorForceOpen:   "非正常开门（强制开门）",
	AlarmDoorClose:       "门被关闭",
	AlarmIRMonitorOff:    "入侵监测被关闭",
	AlarmIRMonitorOn:     "入侵监测开启",
	AlarmDoorMonitorOff:  "门碰开关监测被关闭",
	AlarmDoorMonitorOn:   "门碰开关监测开启",
	AlarmExternalRemoved: "外部设备被拆除",
	AlarmDeviceRemoved:   "本机被拆除",
	AlarmNoEntryInTime:   "刷卡后未及时开门",
	AlarmDoorOpenTooLong: "门开超时未关闭",
	AlarmEmergencyStart:  "紧急输入开始",
	AlarmEmergencyEnd:    "紧急输入结束",
}

// GetAlarms 获取告警记录（从数据库读取，与 xbrother 一致）
func (c *Controller) GetAlarms(offset int) (driver.AlarmData, error) {
	return c.getAlarmsFromDB(offset)
}

// GetAlarmsByTime 按时间获取告警记录
func (c *Controller) GetAlarmsByTime(_ driver.TimeInterval) (driver.AlarmData, error) {
	// TODO: 实现按时间范围查询
	return driver.AlarmData{}, nil
}

// GetAlarmsWhenVerify 验证时获取告警（与 GetAlarms 相同）
func (c *Controller) GetAlarmsWhenVerify(offset interface{}) (driver.AlarmData, error) {
	iOffset, ok := offset.(int)
	if !ok {
		return driver.AlarmData{}, fmt.Errorf("unexpected type offset, expect int")
	}
	return c.getAlarmsFromDB(iOffset)
}

// GetCurrentAlarm 获取当前告警（从 Redis 读取）
func (c *Controller) GetCurrentAlarm() ([]driver.CurrentAlarmData, error) {
	// 与 xbrother 一致，从 Redis 读取当前告警
	redisClient := redis.GetClient()
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	currentAlarmMap, err := redisClient.HGetAll(
		context.Background(),
		utils.GenerateRedisKeyCHDCurrentAlarm(c.chanInfo.ChannelID),
	).Result()
	if err != nil {
		return nil, err
	}

	data := make([]driver.CurrentAlarmData, 0)
	for doorNoStr, v := range currentAlarmMap {
		var driverCurrentAlarmData driver.CurrentAlarmData
		var doorNo int
		if _, err := fmt.Sscanf(doorNoStr, "%d", &doorNo); err != nil {
			continue
		}
		driverCurrentAlarmData.Door = doorNo

		var currentAlarmEvent driver.CurrentAlarmEvent
		if err := json.Unmarshal([]byte(v), &currentAlarmEvent); err != nil {
			continue
		}
		driverCurrentAlarmData.Alarms = []driver.CurrentAlarmEvent{currentAlarmEvent}
		data = append(data, driverCurrentAlarmData)
	}
	return data, nil
}

// getAlarmsFromDB 从数据库读取告警（与 xbrother 一致）
func (c *Controller) getAlarmsFromDB(offset int) (driver.AlarmData, error) {
	const defaultLimit = 100

	totalCount, driverAlarms, err := dac.GetRW().GetDriverAlarms(
		context.Background(),
		c.baseInfo.ID,
		c.chanInfo.ChannelID,
		offset,
		defaultLimit,
	)
	if err != nil {
		return driver.AlarmData{}, err
	}

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

	return driver.AlarmData{
		Last:   int(totalCount),
		Offset: offset + len(driverAlarms),
		Alarms: alarms,
	}, nil
}

// ParseAlarmFromRecord 从 CHD 记录中解析告警详情（REMARK=5 时）
// 事件来源格式: 3字节全0 + AS2(1字节) + AS1(1字节)
func ParseAlarmFromRecord(record CHDRecord) (alarmCode uint16, desc string) {
	if record.Remark != RemarkAlarm {
		return 0, ""
	}

	as2 := record.Source[3]
	as1 := record.Source[4]
	alarmCode = uint16(as2)<<8 | uint16(as1)

	desc = alarmDescMap[alarmCode]
	if desc == "" {
		desc = fmt.Sprintf("未知告警(0x%04X)", alarmCode)
	}
	return alarmCode, desc
}

// ============ 告警存储方法 ============

// saveAlarmToRedis 保存当前告警到 Redis（用于 GetCurrentAlarm）
func (c *Controller) saveAlarmToRedis(doorNo int, alarmType int, desc string) error {
	redisClient := redis.GetClient()
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	event := driver.CurrentAlarmEvent{
		Type: alarmType,
		Desc: desc,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return redisClient.HSet(
		context.Background(),
		utils.GenerateRedisKeyCHDCurrentAlarm(c.chanInfo.ChannelID),
		fmt.Sprintf("%d", doorNo),
		eventJson,
	).Err()
}

// clearAlarmFromRedis 从 Redis 清除告警（告警恢复时调用）
func (c *Controller) clearAlarmFromRedis(doorNo int) error {
	redisClient := redis.GetClient()
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return redisClient.HDel(
		context.Background(),
		utils.GenerateRedisKeyCHDCurrentAlarm(c.chanInfo.ChannelID),
		fmt.Sprintf("%d", doorNo),
	).Err()
}

// ProcessAlarmRecord 处理告警记录
func (c *Controller) ProcessAlarmRecord(record CHDRecord) (*db.DriverAlarm, error) {
	timestamp := recordToTimestamp(record)
	alarmCode, desc := ParseAlarmFromRecord(record)

	// 判断告警状态（开始/恢复）
	state := db.AlarmStateType(1) // 默认告警中
	switch alarmCode {
	case AlarmIRStop, AlarmDoorClose, AlarmIRMonitorOn,
		AlarmDoorMonitorOn, AlarmEmergencyEnd:
		state = db.AlarmStateType(0) // 告警恢复
		// 从 Redis 清除当前告警
		_ = c.clearAlarmFromRedis(int(record.DoorNo))
	default:
		// 保存到 Redis（当前告警）
		_ = c.saveAlarmToRedis(int(record.DoorNo), int(alarmCode), desc)
	}

	alarm := &db.DriverAlarm{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		Timestamp:    timestamp,
		DoorNumber:   db.DoorNumberType(record.DoorNo),
		Type:         db.AlarmType(alarmCode),
		State:        state,
		Description:  desc,
	}

	return alarm, nil
}

// ============ 测试辅助方法 ============

// GetAlarmsFromDevice 直接从设备读取告警（用于测试）
func (c *Controller) GetAlarmsFromDevice() ([]driver.Alarm, error) {
	records, err := c.ReadRecordsFromDevice()
	if err != nil {
		return nil, err
	}

	alarms := make([]driver.Alarm, 0)
	index := 0
	for _, record := range records {
		// 只返回告警类型的记录
		if !alarmRemarks[record.Remark] {
			continue
		}

		index++
		timestamp := recordToTimestamp(record)

		var desc string
		var alarmType int

		if record.Remark == RemarkAlarm {
			// REMARK=5，从事件来源解析告警码
			alarmCode, alarmDesc := ParseAlarmFromRecord(record)
			desc = alarmDesc
			alarmType = int(alarmCode)
		} else {
			// 其他告警类型
			desc = remarkDescMap[record.Remark]
			alarmType = int(record.Remark)
		}

		// 判断状态
		state := driver.AlarmStateAlarming
		alarmCode := uint16(alarmType)
		if alarmCode == AlarmIRStop || alarmCode == AlarmDoorClose ||
			alarmCode == AlarmEmergencyEnd {
			state = driver.AlarmStateRecovery
		}

		alarms = append(alarms, driver.Alarm{
			Index:       index,
			Timestamp:   timestamp,
			DoorNumber:  driver.DoorNumberType(record.DoorNo),
			Type:        driver.AlarmType(alarmType),
			State:       state,
			Description: desc,
		})

		// 打印详细信息
		c.logger.Debugf("  [%d] 门%d %s 状态=%d 时间=%s",
			index, record.DoorNo, desc, state,
			time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	}

	return alarms, nil
}
