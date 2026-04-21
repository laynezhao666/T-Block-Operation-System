// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"dac/repo/dac"
)

// ============ CHD806D4 事件记录格式（16字节/条） ============
//
// | 字节 | 长度 | 含义                    |
// |-----|------|------------------------|
// | 0-4 | 5    | 事件来源（卡号/用户ID等）  |
// | 5-11| 7    | 日期时间（世纪,年,月,日,时,分,秒）|
// | 12  | 1    | WORK-STATUS（工作状态）   |
// | 13  | 1    | REMARK（记录类型）        |
// | 14  | 1    | LINE-STATUS（线路状态）   |
// | 15  | 1    | 门号                     |
//
// REMARK 定义（事件类型）：
// 0  = 刷卡开门记录
// 1  = 键入ID+密码开门
// 2  = 远程开门记录
// 3  = 手动出门记录
// 5  = 报警/状态变化（告警）
// 6  = 掉电再上电（告警）
// 7  = 内部参数被修改
// 8  = 无效卡刷卡（告警）
// 9  = 卡有效期已过（告警）
// 10 = 当前时间无权限（告警）
// ...

const (
	RecordSize = 16 // 每条记录16字节

	RemarkCardAccess      = 0    // 刷卡开门
	RemarkPinAccess       = 1    // ID+密码开门
	RemarkRemoteOpen      = 2    // 远程开门
	RemarkManualExit      = 3    // 手动出门
	RemarkAlarm           = 5    // 报警/状态变化
	RemarkPowerCycle      = 6    // 掉电再上电
	RemarkParamModified   = 7    // 参数被修改
	RemarkInvalidCard     = 8    // 无效卡
	RemarkCardExpired     = 9    // 卡过期
	RemarkNoPermission    = 10   // 无权限
	RemarkLocalAddCard    = 15   // 本地加卡
	RemarkLocalDelCard    = 16   // 本地删卡
	RemarkEmergencyStart  = 0x22 // 紧急输入开始
	RemarkWaitConfirm     = 0x40 // 等待中心确认
	RemarkLocalConfirm    = 0x60 // 本地确认开门
	RemarkSuperCardAccess = 0x70 // 超权限卡开门
	RemarkRemoteNormClose = 0x41 // 远程常闭门
	RemarkRemoteNormOpen  = 0x43 // 远程常开门
)

var remarkDescMap = map[uint8]string{
	RemarkCardAccess:      "刷卡开门",
	RemarkPinAccess:       "ID+密码开门",
	RemarkRemoteOpen:      "远程开门",
	RemarkManualExit:      "手动出门",
	RemarkAlarm:           "报警/状态变化",
	RemarkPowerCycle:      "设备重启",
	RemarkParamModified:   "参数被修改",
	RemarkInvalidCard:     "无效卡",
	RemarkCardExpired:     "卡已过期",
	RemarkNoPermission:    "无进入权限",
	RemarkLocalAddCard:    "本地加卡",
	RemarkLocalDelCard:    "本地删卡",
	RemarkEmergencyStart:  "紧急输入",
	RemarkWaitConfirm:     "等待中心确认",
	RemarkLocalConfirm:    "本地确认开门",
	RemarkSuperCardAccess: "超权限卡开门",
	RemarkRemoteNormClose: "远程常闭门",
	RemarkRemoteNormOpen:  "远程常开门",
}

// 告警类型的 REMARK 集合
var alarmRemarks = map[uint8]bool{
	RemarkAlarm:          true, // 报警/状态变化
	RemarkPowerCycle:     true, // 掉电再上电
	RemarkInvalidCard:    true, // 无效卡
	RemarkCardExpired:    true, // 卡过期
	RemarkNoPermission:   true, // 无权限
	RemarkEmergencyStart: true, // 紧急输入
}

// RecordParams 记录参数（协议5.2.1返回）
type RecordParams struct {
	Bottom      uint16 // 存储区起始指针
	SaveP       uint16 // 最新记录存储位置
	LoadP       uint16 // 当前读取指针
	MaxLen      uint16 // 最大存储记录数
	UnreadCount int    // 未读记录数
}

// CHDRecord CHD 协议记录结构
type CHDRecord struct {
	Source     [5]byte // 事件来源（卡号等）
	Century    uint8   // 世纪
	Year       uint8   // 年
	Month      uint8   // 月
	Day        uint8   // 日
	Hour       uint8   // 时
	Minute     uint8   // 分
	Second     uint8   // 秒
	WorkStatus uint8   // 工作状态
	Remark     uint8   // 记录类型
	LineStatus uint8   // 线路状态
	DoorNo     uint8   // 门号
}

// GetEvents 获取事件记录（从数据库读取，与 xbrother 一致）
func (c *Controller) GetEvents(offset int) (driver.EventData, error) {
	return c.getEventsFromDB(offset)
}

// GetEventsByTime 按时间获取事件记录
func (c *Controller) GetEventsByTime(_ driver.TimeInterval) (driver.EventData, error) {
	// TODO: 实现按时间范围查询
	return driver.EventData{}, nil
}

// GetEventsWhenVerify 验证时获取事件（与 GetEvents 相同）
func (c *Controller) GetEventsWhenVerify(offset interface{}) (driver.EventData, error) {
	iOffset, ok := offset.(int)
	if !ok {
		return driver.EventData{}, fmt.Errorf("unexpected type offset, expect int")
	}
	return c.getEventsFromDB(iOffset)
}

// getEventsFromDB 从数据库读取事件
func (c *Controller) getEventsFromDB(offset int) (driver.EventData, error) {
	const defaultLimit = 100

	totalCount, driverEvents, err := dac.GetRW().GetDriverEvents(
		context.Background(),
		c.baseInfo.ID,
		c.chanInfo.ChannelID,
		offset,
		defaultLimit,
	)
	if err != nil {
		return driver.EventData{}, err
	}

	return driver.EventData{
		Last:   int(totalCount),
		Offset: offset + len(driverEvents),
		Events: utils.ConvertDBEventsToDriver(driverEvents),
	}, nil
}

// ReadRecordsFromDevice 从设备读取历史记录
// 使用协议 5.2.2 顺序读取方式，每次读取一条记录
func (c *Controller) ReadRecordsFromDevice() ([]CHDRecord, error) {
	if err := c.checkAuth(); err != nil {
		return nil, err
	}

	// 1. 先读取记录参数，了解有多少条未读记录
	params, err := c.GetRecordParams()
	if err != nil {
		return nil, fmt.Errorf("读取记录参数失败: %w", err)
	}

	c.logger.Debugf("记录参数: SaveP=%d, LoadP=%d, MaxLen=%d, 未读=%d",
		params.SaveP, params.LoadP, params.MaxLen, params.UnreadCount)

	if params.UnreadCount == 0 {
		c.logger.Debugf("没有新记录")
		return nil, nil
	}

	// 2. 循环读取记录（使用 TYPE=0x82 顺序读取）
	records := make([]CHDRecord, 0, params.UnreadCount)
	maxRead := 100 // 限制单次最多读取100条，防止死循环

	for i := 0; i < maxRead && i < params.UnreadCount; i++ {
		record, err := c.readOneRecord()
		if err != nil {
			// 如果返回"无记录"错误，说明已读完
			if err.Error() == "no_record" {
				c.logger.Debugf("已读取完毕，共 %d 条记录", len(records))
				break
			}
			return records, fmt.Errorf("读取第%d条记录失败: %w", i+1, err)
		}
		records = append(records, record)

		// 打印进度
		if (i+1)%10 == 0 {
			c.logger.Debugf("已读取 %d 条记录...", i+1)
		}
	}

	c.logger.Infof("读取完成，共 %d 条记录", len(records))
	return records, nil
}

// GetRecordParams 读取记录参数（协议5.2.1）
// 返回 SAVEP、LOADP、MAXLEN 等指针信息
func (c *Controller) GetRecordParams() (*RecordParams, error) {
	if err := c.checkAuth(); err != nil {
		return nil, err
	}

	// 协议 5.2.1: 读取记录参数
	// CID2 = 0x4A (读取信息)
	// GROUP = 0x03 (读取信息组)
	// TYPE = 0x81 (读取记录参数)
	// DATAF = 0 或无

	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeRecordGetParam, // 0x81
		[]byte{0x00},               // DATAF = 0
		consts2.GetRRPCRecordGetParam,
	)
	if err != nil {
		return nil, fmt.Errorf("读取记录参数失败: %w", err)
	}

	// 响应格式（9字节）:
	// BOTTOM(2字节) + SAVEP(2字节) + LOADP(2字节) + 备份(1字节) + MAXLEN(2字节)
	if len(respInfo) < 9 {
		return nil, fmt.Errorf("响应数据太短: %d 字节，期望 9 字节", len(respInfo))
	}

	params := &RecordParams{
		Bottom: uint16(respInfo[0]) | uint16(respInfo[1])<<8,
		SaveP:  uint16(respInfo[2]) | uint16(respInfo[3])<<8,
		LoadP:  uint16(respInfo[4]) | uint16(respInfo[5])<<8,
		// respInfo[6] 是备份字节，忽略
		MaxLen: uint16(respInfo[7]) | uint16(respInfo[8])<<8,
	}

	// 计算未读记录数
	if params.SaveP >= params.LoadP {
		params.UnreadCount = int(params.SaveP - params.LoadP)
	} else {
		// 环形缓冲区回绕
		params.UnreadCount = int(params.MaxLen) - int(params.LoadP) + int(params.SaveP)
	}

	return params, nil
}

// readOneRecord 顺序读取一条记录（协议5.2.2）
// SM 从 LOADP 位置读取一条记录返给 SU，SM 自动将 LOADP 指向下一条记录
func (c *Controller) readOneRecord() (CHDRecord, error) {
	// 协议 5.2.2: 顺序读取一条历史记录
	// CID2 = 0x4A (读取信息)
	// GROUP = 0x03 (读取信息组)
	// TYPE = 0x82 (顺序读取)
	// DATAF = 无或1字节（无定义）

	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeRecordGetSeq, // 0x82
		nil,                      // DATAF 可以为空
		consts2.GetRRPCRecordGetSeq,
	)
	if err != nil {
		// 检查是否是"无记录"错误（RTN=0xE4 或 RTN=0xE5）
		if err.Error() == "存储空间已空" || err.Error() == "无相应信息项" {
			return CHDRecord{}, fmt.Errorf("no_record")
		}
		return CHDRecord{}, err
	}

	// 响应格式: 16字节记录
	if len(respInfo) < RecordSize {
		return CHDRecord{}, fmt.Errorf("no_record")
	}

	return parseRecord(respInfo[:RecordSize]), nil
}

// ReadLatestRecord 读取最新事件记录（协议5.2.5）
func (c *Controller) ReadLatestRecord() (*CHDRecord, error) {
	if err := c.checkAuth(); err != nil {
		return nil, err
	}

	// 协议 5.2.5: 查询最新事件记录
	// TYPE = 0x85
	// 返回最新发生的事件记录原形，共16字节

	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeRecordGetLatest, // 0x85
		[]byte{0x00},
		consts2.GetRRPCRecordGetLatest,
	)
	if err != nil {
		return nil, fmt.Errorf("读取最新记录失败: %w", err)
	}

	if len(respInfo) < RecordSize {
		return nil, fmt.Errorf("无最新事件")
	}

	record := parseRecord(respInfo[:RecordSize])
	return &record, nil
}

// parseRecord 解析单条记录（16字节）
func parseRecord(data []byte) CHDRecord {
	if len(data) < RecordSize {
		return CHDRecord{}
	}

	var record CHDRecord
	copy(record.Source[:], data[0:5])
	record.Century = data[5]
	record.Year = data[6]
	record.Month = data[7]
	record.Day = data[8]
	record.Hour = data[9]
	record.Minute = data[10]
	record.Second = data[11]
	record.WorkStatus = data[12]
	record.Remark = data[13]
	record.LineStatus = data[14]
	record.DoorNo = data[15]

	return record
}

// ProcessAndSaveRecords 处理并保存记录到数据库
// 根据 REMARK 类型分类为事件或告警
func (c *Controller) ProcessAndSaveRecords(records []CHDRecord) error {
	ctx := context.Background()

	events := make([]db.DriverEvent, 0)
	alarms := make([]db.DriverAlarm, 0)

	for _, record := range records {
		// 解析时间戳
		timestamp := recordToTimestamp(record)

		// 解析卡号
		cardNo := hex.EncodeToString(record.Source[:])

		// 获取描述
		desc := remarkDescMap[record.Remark]
		if desc == "" {
			desc = fmt.Sprintf("未知类型(0x%02X)", record.Remark)
		}

		// 根据 REMARK 分类
		if alarmRemarks[record.Remark] {
			// 告警类型
			alarm := db.DriverAlarm{
				ControllerID: c.baseInfo.ID,
				ChannelID:    c.chanInfo.ChannelID,
				Timestamp:    timestamp,
				DoorNumber:   db.DoorNumberType(record.DoorNo),
				Type:         db.AlarmType(record.Remark),
				State:        db.AlarmStateType(1), // 告警中
				Description:  desc,
			}
			alarms = append(alarms, alarm)
		} else {
			// 事件类型
			direction := parseDirection(record)
			event := db.DriverEvent{
				ControllerID: c.baseInfo.ID,
				ChannelID:    c.chanInfo.ChannelID,
				Timestamp:    timestamp,
				CardNumber:   cardNo,
				Username:     "", // 需要查询用户名
				DoorNumber:   db.DoorNumberType(record.DoorNo),
				Direction:    db.DirectionType(direction),
				Type:         db.EventType(record.Remark),
				Description:  desc,
			}
			events = append(events, event)
		}
	}

	// 保存事件
	if len(events) > 0 {
		if err := dac.GetRW().SetDriverEvents(ctx, c.baseInfo.ID, events); err != nil {
			return fmt.Errorf("保存事件失败: %w", err)
		}
		c.logger.Infof("保存 %d 条事件到数据库", len(events))
	}

	// 保存告警
	if len(alarms) > 0 {
		if err := dac.GetRW().SetDriverAlarms(ctx, c.baseInfo.ID, alarms); err != nil {
			return fmt.Errorf("保存告警失败: %w", err)
		}
		c.logger.Infof("保存 %d 条告警到数据库", len(alarms))
	}

	return nil
}

// recordToTimestamp 将记录中的日期时间转换为时间戳
func recordToTimestamp(record CHDRecord) int64 {
	year := int(record.Century)*100 + int(record.Year)
	t := time.Date(
		year,
		time.Month(record.Month),
		int(record.Day),
		int(record.Hour),
		int(record.Minute),
		int(record.Second),
		0,
		time.Local,
	)
	return t.Unix()
}

// parseDirection 解析进出方向
func parseDirection(record CHDRecord) int {
	// 根据 WORK-STATUS 的 D1 位判断
	// D1=0 进入刷卡（第1头），D1=1 出门刷卡（第2头）
	if record.WorkStatus&0x02 != 0 {
		return 1 // 出门
	}
	return 0 // 进入
}

// FetchAndSaveRecords 拉取并保存记录（触发包后调用）
func (c *Controller) FetchAndSaveRecords() error {
	records, err := c.ReadRecordsFromDevice()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		c.logger.Debugf("没有新记录")
		return nil
	}

	return c.ProcessAndSaveRecords(records)
}

// StartEventListener 启动事件监听（后台协程）
func (c *Controller) StartEventListener(ctx context.Context) {
	go func() {
		for {
			// 检查 Server 是否已初始化，避免空指针访问
			if c.Server == nil {
				select {
				case <-ctx.Done():
					c.logger.Infof("事件监听停止")
					return
				case <-time.After(time.Second):
					continue
				}
			}

			select {
			case <-ctx.Done():
				c.logger.Infof("事件监听停止")
				return
			case triggerData, ok := <-c.Server.eventChan:
				if !ok {
					c.logger.Warnf("事件通道已关闭")
					return
				}

				// 收到触发包，拉取记录
				c.logger.Infof("收到触发包: % X", triggerData)
				if err := c.FetchAndSaveRecords(); err != nil {
					c.logger.Errorf("拉取记录失败: %v", err)
				}
			}
		}
	}()
}

// ============ 测试辅助方法 ============

// GetEventsFromDevice 直接从设备读取事件（用于测试）
func (c *Controller) GetEventsFromDevice() ([]driver.Event, error) {
	records, err := c.ReadRecordsFromDevice()
	if err != nil {
		return nil, err
	}

	events := make([]driver.Event, 0)
	for i, record := range records {
		// 只返回事件类型的记录
		if alarmRemarks[record.Remark] {
			continue
		}

		timestamp := recordToTimestamp(record)
		cardNo := hex.EncodeToString(record.Source[:])
		desc := remarkDescMap[record.Remark]
		if desc == "" {
			desc = fmt.Sprintf("未知类型(0x%02X)", record.Remark)
		}

		events = append(events, driver.Event{
			Index:       i + 1,
			Timestamp:   timestamp,
			CardNumber:  cardNo,
			Username:    "",
			DoorNumber:  driver.DoorNumberType(record.DoorNo),
			Direction:   driver.DirectionType(parseDirection(record)),
			Type:        driver.EventType(record.Remark),
			Description: desc,
		})

		// 打印详细信息
		fmt.Printf("  [%d] 门%d %s 卡号=%s 时间=%s\n",
			i+1, record.DoorNo, desc, cardNo,
			time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	}

	return events, nil
}

// PrintRecordDetails 打印记录详情（用于调试）
func PrintRecordDetails(record CHDRecord) {
	timestamp := recordToTimestamp(record)
	cardNo := hex.EncodeToString(record.Source[:])
	desc := remarkDescMap[record.Remark]

	fmt.Println("----------------------------------------")
	fmt.Printf("记录类型(REMARK): 0x%02X (%s)\n", record.Remark, desc)
	fmt.Printf("门号: %d\n", record.DoorNo)
	fmt.Printf("事件来源: %s\n", cardNo)
	fmt.Printf("时间: %s\n", time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("工作状态(WORK-STATUS): 0x%02X\n", record.WorkStatus)
	fmt.Printf("线路状态(LINE-STATUS): 0x%02X\n", record.LineStatus)

	// 解析 LINE-STATUS
	fmt.Println("  线路状态详情:")
	if record.LineStatus&0x80 != 0 {
		fmt.Println("    D7=1: 设备被拆除")
	}
	if record.LineStatus&0x40 != 0 {
		fmt.Println("    D6=1: 当前正在报警")
	}
	if record.LineStatus&0x08 != 0 {
		fmt.Println("    D3=1: 门是开的")
	} else {
		fmt.Println("    D3=0: 门是关的")
	}
	if record.LineStatus&0x04 != 0 {
		fmt.Println("    D2=1: 红外报警")
	}
	if record.LineStatus&0x02 != 0 {
		fmt.Println("    D1=1: 出门按钮按下")
	}
	if record.LineStatus&0x01 != 0 {
		fmt.Println("    D0=1: 紧急输入有效")
	}

	// 根据 REMARK 解析 WORK-STATUS
	switch record.Remark {
	case RemarkCardAccess:
		fmt.Println("  工作状态详情(刷卡开门):")
		if record.WorkStatus&0x80 != 0 {
			fmt.Println("    D7=1: 确认了用户密码")
		}
		if record.WorkStatus&0x40 != 0 {
			fmt.Println("    D6=1: 原来门是开的")
		}
		if record.WorkStatus&0x20 != 0 {
			fmt.Println("    D5=1: 未在规定延时内开门")
		}
		if record.WorkStatus&0x10 != 0 {
			fmt.Println("    D4=1: 门一直开着")
		}
		if record.WorkStatus&0x04 != 0 {
			fmt.Println("    D2=1: 胁迫状态")
		}
		if record.WorkStatus&0x02 != 0 {
			fmt.Println("    D1=1: 出门刷卡(第2头)")
		} else {
			fmt.Println("    D1=0: 进入刷卡(第1头)")
		}
	}
}

// parseCardNo 解析卡号（从5字节转换为字符串）
func parseCardNo(source [5]byte) string {
	// 方式1：转换为十进制数字字符串
	cardNo := uint64(source[0]) |
		uint64(source[1])<<8 |
		uint64(source[2])<<16 |
		uint64(source[3])<<24 |
		uint64(source[4])<<32
	if cardNo == 0 {
		return ""
	}
	return strconv.FormatUint(cardNo, 10)
}
