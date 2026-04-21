// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"fmt"

	"dac/entity/model/driver"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"dac/repo/dac"
)

// GetDoorParameter 获取门参数
// 从数据库读取
func (c *Controller) GetDoorParameter() ([]driver.DoorParameter, error) {
	res, err := dac.GetRW().GetDriverDoorParameters(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("doors is empty")
	}
	return utils.ConvertDBDriverDoorParams(res), nil
}

// SetDoorParameter 设置门参数
// 协议：COMMAND=0x02, TYPE=0x86, DATAF=门号+8字节参数
func (c *Controller) SetDoorParameter(params []driver.DoorParameter) error {
	for _, param := range params {
		if err := c.setOneDoorParam(param); err != nil {
			return fmt.Errorf("设置门%d参数失败: %w", param.Number, err)
		}
	}
	return nil
}

// setOneDoorParam 设置单门参数
func (c *Controller) setOneDoorParam(param driver.DoorParameter) error {
	// 构建 DATAF: 门号(1字节) + 8字节参数
	dataf := make([]byte, 9)
	dataf[0] = byte(param.Number) // 门号

	// 构建控制字 CTRL1（CHD806D4 协议定义）
	// D7: 监控门状态, D6: 监控红外状态, D5: 第2感应头密码确认, D4: 第1感应头密码确认
	// D3: 门磁开路, D2: 红外开路, D1: 密码时段, D0: 紧急输入
	ctrl1 := byte(0x88) // 默认值: D7=1(监控门状态), D3=1(门磁开路)

	// 构建控制字 CTRL2
	// D7: 报警继电器, D6: 手动出门, D5/D4: 刷卡继电器, D3: 无效卡, D2: 手动按钮, D1: 第2感应头, D0: 同步
	ctrl2 := byte(0x86) // 默认值: D7=1(报警), D2=1(手动按钮), D1=1(第2感应头)

	// 构建控制字 CTRL3 和 CTRL4（使用默认值）
	ctrl3 := byte(0x00)
	ctrl4 := byte(0x00)

	// 将秒转换为协议的0.1秒单位（driver.DoorParameter.KeepOpenTime 单位是秒）
	relayDelay := param.KeepOpenTime * 10 // 秒转0.1秒，例如：5秒 -> 50
	if relayDelay > 255 {
		relayDelay = 255 // 最大25.5秒
	}

	// OpenTimeout 限制范围 2-255 秒
	openTimeout := param.OpenTimeout
	if openTimeout < 2 {
		openTimeout = 2
	}
	if openTimeout > 255 {
		openTimeout = 255
	}

	dataf[1] = ctrl1
	dataf[2] = byte(relayDelay)  // 门锁动作延时（0.1秒单位）
	dataf[3] = byte(openTimeout) // 开门等待进入延时（秒）
	dataf[4] = 2                 // IR SURE - 红外报警确认延时（默认2秒）
	dataf[5] = 5                 // IR ONDLY - 布防开启延时（默认5秒）
	dataf[6] = ctrl2
	dataf[7] = ctrl3
	dataf[8] = ctrl4

	// Server.Request(cid, groupCode, cmdType, data, rrpcKeyFunc)
	// CID=0x49(设置参数), GROUP=0x02, TYPE=0x86
	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeDoorParamSet,
		dataf,
		consts2.GetRRPCDoorParamSet,
	)
	if err != nil {
		return err
	}

	// 将完整参数存入数据库
	resDoorParameter := utils.ConvertDriverDoorParamToDB(c.baseInfo.ID, c.chanInfo.ChannelID, param)
	return dac.GetRW().AddDriverDoorParameter(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, resDoorParameter)
}

// getDoorNum 获取门数量
func (c *Controller) getDoorNum() int {
	doorNum := 2 // 默认2门
	if num, ok := c.chanInfo.Extend["door_num"].(int); ok && num > 0 {
		doorNum = num
	}
	return doorNum
}

// GetDoorParameterFromDevice 从设备读取门参数（用于测试验证设备返回数据）
// 协议：COMMAND=0x03, TYPE=0x90, DATAF=门号(0xFF表示所有门)
func (c *Controller) GetDoorParameterFromDevice() ([]driver.DoorParameter, error) {
	// 发送请求，门号 0xFF 表示读取所有门参数
	dataf := []byte{0xFF}

	resp, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeDoorParamGet,
		dataf,
		consts2.GetRRPCDoorParamGet,
	)
	if err != nil {
		return nil, fmt.Errorf("读取门参数失败: %w", err)
	}

	// 每门参数 8 字节
	doorNum := c.getDoorNum()
	expectedLen := doorNum * 8
	if len(resp) < expectedLen {
		return nil, fmt.Errorf("响应数据长度不足: 期望%d字节, 实际%d字节", expectedLen, len(resp))
	}

	doorParams := make([]driver.DoorParameter, doorNum)
	for i := 0; i < doorNum; i++ {
		offset := i * 8
		doorParams[i] = c.parseDoorParamFromDevice(byte(i+1), resp[offset:offset+8])
	}

	return doorParams, nil
}

// parseDoorParamFromDevice 解析设备返回的门参数（用于测试）
// 8字节参数格式: CTRL1, RELAY_DELAY, OPEN_DELAY, IR_SURE, IR_ONDLY, CTRL2, CTRL3, CTRL4
func (c *Controller) parseDoorParamFromDevice(doorNo byte, data []byte) driver.DoorParameter {
	if len(data) < 8 {
		return driver.DoorParameter{Number: driver.DoorNumberType(doorNo)}
	}

	ctrl1 := data[0]      // 控制字1
	relayDelay := data[1] // 门锁动作延时（0.1秒单位）
	openDelay := data[2]  // 开门等待进入延时（秒）
	_ = data[3]           // IR_SURE 红外报警确认延时
	_ = data[4]           // IR_ONDLY 布防开启延时
	ctrl2 := data[5]      // 控制字2
	ctrl3 := data[6]      // 控制字3
	ctrl4 := data[7]      // 控制字4

	// 将 0.1秒 转换为 秒
	keepOpenTimeSec := int(relayDelay) / 10

	// 解析开门模式（从 CTRL3 和 CTRL4 解析多卡确认模式）
	// CTRL4-D0 + CTRL3-D3: 00=单卡, 01=双卡, 10=三卡, 11=四卡
	openMode := 0
	if ctrl3&0x08 != 0 { // D3
		openMode |= 1
	}
	if ctrl4&0x01 != 0 { // D0
		openMode |= 2
	}

	// 解析紧急输入模式（CTRL1-D0: 0=常开, 1=常闭）
	fireSignalMode := 0
	if ctrl1&0x01 != 0 {
		fireSignalMode = 1
	}

	fmt.Printf("[解析门%d参数] 原始数据: %02X %02X %02X %02X %02X %02X %02X %02X\n",
		doorNo, data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7])
	fmt.Printf("  CTRL1=0x%02X, RELAY_DELAY=%d(0.1秒)=%d秒, OPEN_DELAY=%d秒\n",
		ctrl1, relayDelay, keepOpenTimeSec, openDelay)
	fmt.Printf("  CTRL2=0x%02X, CTRL3=0x%02X, CTRL4=0x%02X\n", ctrl2, ctrl3, ctrl4)
	fmt.Printf("  解析结果: KeepOpenTime=%d秒, OpenTimeout=%d秒, OpenMode=%d, FireSignalMode=%d\n",
		keepOpenTimeSec, openDelay, openMode, fireSignalMode)

	return driver.DoorParameter{
		Number:         driver.DoorNumberType(doorNo),
		Name:           fmt.Sprintf("门%d", doorNo),
		Password:       "",              // 密码需要通过 0x92 命令单独读取
		KeepOpenTime:   keepOpenTimeSec, // 门锁动作延时（秒）
		OpenTimeout:    int(openDelay),  // 开门等待进入延时（秒）
		LockCount:      0,               // 协议中暂无此字段
		LockTime:       0,               // 协议中暂无此字段
		VerifyInterval: 0,               // 协议中暂无此字段
		OpenMode:       driver.OpenModeType(openMode),
		FireSignalMode: fireSignalMode,
	}
}
