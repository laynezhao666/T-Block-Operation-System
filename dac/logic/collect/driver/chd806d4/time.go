// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"fmt"
	"time"
)

// ============ 时间管理 ============

// CHD 协议日期时间结构说明：
// 1. 日期时间（8字节）：世纪，年，月，日，星期，时，分，秒

// GetTime 读取时间
// 对应协议 4.2.1：COMMAND TYPE=0x81
// DATA格式：门号(1字节) + 表序号(1字节,0-31) + 4组时段(16字节,HH:MM-HH:MM×4)
func (c *Controller) GetTime() (string, error) {
	// 发送读取实时钟命令
	// CID2 = 0x4A (读取DCU存储信息数据包)
	// COMMAND GROUP = 0x03 (读取信息组)
	// COMMAND TYPE = 0x80 (读取实时钟)
	// DATAF = 0 或无
	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeReadTime,
		[]byte{0x00}, // DATAF = 0
		consts2.GetRRPCReadTime,
	)
	if err != nil {
		return "", fmt.Errorf("读取时间失败: %w", err)
	}

	// 解析返回的时间数据
	// 返回格式：世纪,年，月，日，星期，时，分，秒，共8字节BCD
	if len(respInfo) != 8 {
		return "", fmt.Errorf("时间数据长度错误: 期望8字节，实际%d字节", len(respInfo))
	}

	// BCD转十进制
	bcdToDec := func(bcd byte) int {
		return int(bcd>>4)*10 + int(bcd&0x0F)
	}

	century := bcdToDec(respInfo[0])
	year := bcdToDec(respInfo[1])
	month := bcdToDec(respInfo[2])
	day := bcdToDec(respInfo[3])
	weekday := bcdToDec(respInfo[4])
	hour := bcdToDec(respInfo[5])
	minute := bcdToDec(respInfo[6])
	second := bcdToDec(respInfo[7])

	// 构造完整年份（世纪+年）
	fullYear := century*100 + year

	// 格式化时间字符串
	timeStr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d (星期%d)",
		fullYear, month, day, hour, minute, second, weekday)

	return timeStr, nil
}

// SetTime 设置时间
// 对应协议 4.1：COMMAND TYPE = 0x80
// DATA格式：世纪年(2字节) + 月(1字节) + 日(1字节) + 星期(1字节) + 时(1字节) + 分(1字节) + 秒(1字节)
func (c *Controller) SetTime() error {
	// 获取当前系统时间
	now := time.Now()

	// 十进制转BCD
	decToBCD := func(dec int) byte {
		return byte((dec/10)<<4 | (dec % 10))
	}

	// 构造8字节BCD时间数据
	// 格式：世纪,年，月，日，星期，时，分，秒
	century := now.Year() / 100
	year := now.Year() % 100
	month := int(now.Month())
	day := now.Day()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 星期日用7表示
	}
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	// 构造DATAF（8字节BCD）
	dataf := []byte{
		decToBCD(century), // 世纪（20）
		decToBCD(year),    // 年（00-99）
		decToBCD(month),   // 月（1-12）
		decToBCD(day),     // 日（1-31）
		decToBCD(weekday), // 星期（1-7，7=星期日）
		decToBCD(hour),    // 时（0-23）
		decToBCD(minute),  // 分（0-59）
		decToBCD(second),  // 秒（0-59）
	}

	// 发送设置时间命令
	// CID2 = 0x49 (设置参数命令)
	// COMMAND GROUP = 0x02 (设置命令组)
	// COMMAND TYPE = 0x80 (日期时间同步)
	// DATAF = 8字节BCD时间数据
	respInfo, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeTimeSync,
		dataf,
		consts2.GetRRPCTimeSet,
	)
	if err != nil {
		return fmt.Errorf("设置时间失败: %w", err)
	}

	// 检查返回结果
	// RTN=0 表示设置成功，非零值表示失败
	if len(respInfo) > 0 {
		return fmt.Errorf("设置时间失败: 返回了额外数据 %v", respInfo)
	}

	return nil
}
