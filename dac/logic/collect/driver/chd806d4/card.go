// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"context"
	"fmt"
	"strconv"

	"dac/entity/model/driver"
	consts2 "dac/logic/collect/driver/chd806d4/consts"
	"dac/logic/collect/driver/xbrother/consts"
	"dac/repo/dac"
)

// ============ 卡管理 ============

// CHD 协议 4.3.1 授权用户卡
// COMMAND GROUP = 0x02, COMMAND TYPE = 0x83
// DATAF（19字节）用户描述：
//   - 卡片编号: 5字节(HEX)，低字节在前
//   - 用户序号: 4字节(BCD)
//   - 用户密码: 2字节(BCD)
//   - 有效期:   4字节(YYYY,MM,DD)，每字节1个BCD数
//   - 门1-4权限: 4字节，每门1字节

// AddCard 授权一个用户卡（至门控制器）
// 对应协议 4.3.1：COMMAND TYPE=0x83
// DATAF格式（19字节）：卡片编号(5B) + 用户序号(4B) + 用户密码(2B) + 有效期(4B) + 门权限(4B)
//
// CardFlag 含义（业务层面）：
//   - 0 = 启用（授权卡到门控器）
//   - 1 = 禁用（取消卡权限）
func (c *Controller) AddCard(card driver.Card) error {
	// CardFlag=1 表示禁用，需要取消卡权限而非授权
	if card.CardFlag == 1 {
		return c.cancelCardPermission(card.CardNo)
	}

	// CardFlag=0 表示启用，继续授权流程

	// 十进制转BCD编码（两位数字压缩为1字节）
	decToBCD := func(dec int) byte {
		return byte((dec/10)<<4 | (dec % 10))
	}

	// 构建 DATAF（19字节用户描述）
	dataf := make([]byte, 19)
	offset := 0

	// ===== 字段1：卡片编号（5字节HEX）=====
	// 低字节在前（小端序），不足5字节高位补0
	cardNoInt, err := strconv.ParseUint(card.CardNo, 10, 64)
	if err != nil {
		return fmt.Errorf("卡号解析失败: %w", err)
	}
	for i := 0; i < 5; i++ {
		dataf[offset+i] = byte(cardNoInt & 0xFF)
		cardNoInt >>= 8
	}
	offset += 5

	// ===== 字段2：用户序号（4字节BCD）=====
	// 例如 "202512" -> 0x00 0x20 0x25 0x12
	userSeq, _ := strconv.Atoi(card.UserName)
	dataf[offset] = decToBCD(userSeq / 1000000 % 100) // 高2位
	dataf[offset+1] = decToBCD(userSeq / 10000 % 100) // 次高2位
	dataf[offset+2] = decToBCD(userSeq / 100 % 100)   // 次低2位
	dataf[offset+3] = decToBCD(userSeq % 100)         // 低2位
	offset += 4

	// ===== 字段3：用户密码（2字节BCD）=====
	// 例如 "1234" -> 0x12 0x34，最多4位数字
	// 如果密码超过4位，取前4位
	pwdStr := card.Password
	if len(pwdStr) > 4 {
		pwdStr = pwdStr[:4] // 取前4位
	}
	pwd, _ := strconv.Atoi(pwdStr)
	dataf[offset] = decToBCD(pwd / 100 % 100) // 高2位
	dataf[offset+1] = decToBCD(pwd % 100)     // 低2位
	offset += 2

	// ===== 字段4：有效期（4字节）=====
	// 格式：YYYY, MM, DD，每字节1个BCD数
	// 默认设置为2099年12月31日
	expireYear := 2099
	expireMonth := 12
	expireDay := 31
	dataf[offset] = decToBCD(expireYear / 100)   // 世纪 20
	dataf[offset+1] = decToBCD(expireYear % 100) // 年 99
	dataf[offset+2] = decToBCD(expireMonth)      // 月 12
	dataf[offset+3] = decToBCD(expireDay)        // 日 31
	offset += 4

	// ===== 字段5：门1-4权限（4字节）=====
	// 每个门1字节，CHD协议门权限值含义：
	//   0x00 = 特权卡（不受准进时段限制）
	//   0x01 = 第一类普通卡（受准进时段限制）
	//   0x02 = 第二类普通卡
	//   0x03 = 第三类普通卡
	//   0x04 = 第四类普通卡
	//   0xFF = 无权限
	// 注意：card.CardFlag 代表的是卡的启用/禁用状态（0=启用，1=禁用），不是门权限类型
	doorPermissions := [4]byte{0xFF, 0xFF, 0xFF, 0xFF} // 默认无权限

	// 根据 card.DoorNos 设置有权限的门
	// 统一使用第一类普通卡(0x01)作为门权限值
	const doorPermValue = byte(0x01) // 第一类普通卡
	for _, doorNo := range card.DoorNos {
		if doorNo >= 1 && doorNo <= 4 {
			doorPermissions[doorNo-1] = doorPermValue
		}
	}
	copy(dataf[offset:offset+4], doorPermissions[:])

	// 发送授权用户卡命令
	// CID2 = 0x49 (设置参数命令)
	// COMMAND GROUP = 0x02 (设置操作组)
	// COMMAND TYPE = 0x83 (授权用户卡)
	_, err = c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeCardAuthUser,
		dataf,
		consts2.GetRRPCCardAuthUser,
	)
	if err != nil {
		return fmt.Errorf("授权用户卡失败: %w", err)
	}

	return nil
}

// decToBCD 十进制转BCD编码（两位数字压缩为1字节）
func decToBCD(dec int) byte {
	return byte((dec/10)<<4 | (dec % 10))
}

// bcdToDec BCD编码转十进制
func bcdToDec(bcd byte) int {
	return int(bcd>>4)*10 + int(bcd&0x0F)
}

// ============ 读取卡用户 ============

// GetCardCount 读取已授权的用户数量
// 对应协议 5.4.1：COMMAND TYPE=0x8B, DATAF=0或无
// 返回：用户数量（2字节HEX，低8位在前）
func (c *Controller) GetCardCount() (int, error) {
	// CID2 = 0x4A (读取信息命令)
	// COMMAND GROUP = 0x03 (读取信息组)
	// COMMAND TYPE = 0x8B (读取用户数量)
	// DATAF = 0x00 或空
	resp, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeCardGetCount,
		[]byte{0x00},
		consts2.GetRRPCCardGetCount,
	)
	if err != nil {
		return 0, fmt.Errorf("读取用户数量失败: %w", err)
	}

	// 响应：2字节HEX，低8位在前，高8位在后
	if len(resp) < 2 {
		return 0, fmt.Errorf("响应数据长度不足: %d", len(resp))
	}
	count := int(resp[0]) | (int(resp[1]) << 8)
	return count, nil
}

// GetCardByPosition 按存储位置读取用户信息
// 对应协议 5.4.2：COMMAND TYPE=0x8C, DATAF=2字节(位置序号，低位在前)
// 返回：用户卡信息
func (c *Controller) GetCardByPosition(position int) (*driver.Card, error) {
	// DATAF = 2字节位置序号（低位在前）
	dataf := []byte{
		byte(position & 0xFF),
		byte((position >> 8) & 0xFF),
	}

	resp, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeCardGetByPos,
		dataf,
		consts2.GetRRPCCardGetByPos,
	)
	if err != nil {
		return nil, fmt.Errorf("按位置读取用户失败: %w", err)
	}

	return c.parseCardResponse(resp)
}

// GetCardByUserID 按用户ID查询用户是否存在
// 对应协议 5.4.3：COMMAND TYPE=0x8D, DATAF=4字节用户编号(BCD)
// 返回：用户卡信息
func (c *Controller) GetCardByUserID(userID string) (*driver.Card, error) {
	// DATAF = 4字节用户编号(BCD)
	userSeq, _ := strconv.Atoi(userID)
	dataf := []byte{
		decToBCD(userSeq / 1000000 % 100), // 高2位
		decToBCD(userSeq / 10000 % 100),   // 次高2位
		decToBCD(userSeq / 100 % 100),     // 次低2位
		decToBCD(userSeq % 100),           // 低2位
	}

	resp, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeCardGetByUserID,
		dataf,
		consts2.GetRRPCCardGetByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("按用户ID查询失败: %w", err)
	}

	return c.parseCardResponse(resp)
}

// GetCardByCardNo 按卡号查询用户是否存在
// 对应协议 5.4.4：COMMAND TYPE=0x8E, DATAF=5字节卡号(HEX)
// 返回：用户卡信息
func (c *Controller) GetCardByCardNo(cardNo string) (*driver.Card, error) {
	// DATAF = 5字节卡号(HEX)，低字节在前
	cardNoInt, err := strconv.ParseUint(cardNo, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("卡号解析失败: %w", err)
	}

	dataf := make([]byte, 5)
	for i := 0; i < 5; i++ {
		dataf[i] = byte(cardNoInt & 0xFF)
		cardNoInt >>= 8
	}

	resp, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeCardGetByCardNo,
		dataf,
		consts2.GetRRPCCardGetByCardNo,
	)
	if err != nil {
		return nil, fmt.Errorf("按卡号查询失败: %w", err)
	}

	return c.parseCardResponse(resp)
}

// GetCards 获取卡用户（遍历方式）
// CHD协议没有批量获取接口，需要先获取总数，再逐个读取
// offset: 起始位置（从1开始）
func (c *Controller) GetCards(offset int) (driver.CardData, error) {
	var cardData driver.CardData

	// 1. 获取用户总数
	count, err := c.GetCardCount()
	if err != nil {
		return cardData, fmt.Errorf("获取用户总数失败: %w", err)
	}
	cardData.Total = count

	if count == 0 {
		cardData.Cards = []driver.Card{}
		return cardData, nil
	}

	// 2. 从 offset 位置开始逐个读取用户信息
	if offset <= 0 {
		offset = 1
	}
	cards := make([]driver.Card, 0)
	for i := offset; i <= count; i++ {
		card, err := c.GetCardByPosition(i)
		if err != nil {
			// 跳过读取失败的位置，可能是空位置
			continue
		}
		if card != nil {
			cards = append(cards, *card)
		}
	}

	cardData.Cards = cards
	cardData.Offset = offset + len(cards)
	return cardData, nil
}

// parseCardResponse 解析卡用户响应数据
// 协议 5.4.2 返回格式（同 5.4.3, 5.4.4）：
//
//	卡编号(5字节HEX) + 用户编号(4字节BCD) + 用户密码(2字节BCD)
//	+ 有效期(4字节YYYY,MM,DD) + 门1权限(1字节) + ... + 门4权限(1字节)
//	= 共19字节
func (c *Controller) parseCardResponse(resp []byte) (*driver.Card, error) {
	// 最少需要19字节
	if len(resp) < 19 {
		return nil, fmt.Errorf("响应数据长度不足: %d, 需要至少19字节", len(resp))
	}

	offset := 0

	// 1. 解析卡片编号（5字节HEX，低字节在前）
	var cardNoInt uint64
	for i := 4; i >= 0; i-- {
		cardNoInt = (cardNoInt << 8) | uint64(resp[offset+i])
	}
	cardNo := strconv.FormatUint(cardNoInt, 10)
	offset += 5

	// 2. 解析用户序号（4字节BCD）
	userSeq := bcdToDec(resp[offset])*1000000 +
		bcdToDec(resp[offset+1])*10000 +
		bcdToDec(resp[offset+2])*100 +
		bcdToDec(resp[offset+3])
	userName := strconv.Itoa(userSeq)
	offset += 4

	// 3. 解析用户密码（2字节BCD）
	pwd := bcdToDec(resp[offset])*100 + bcdToDec(resp[offset+1])
	password := fmt.Sprintf("%04d", pwd)
	offset += 2

	// 4. 解析有效期（4字节BCD: 世纪、年、月、日）
	// year := bcdToDec(resp[offset])*100 + bcdToDec(resp[offset+1])
	// month := bcdToDec(resp[offset+2])
	// day := bcdToDec(resp[offset+3])
	offset += 4

	// 5. 解析门权限（4字节，每门1字节）
	// CHD协议门权限值：0x00=特权卡, 0x01-0x04=普通卡, 0xFF=无权限
	// 注意：CardFlag代表启用/禁用状态（0=启用，1=禁用），不是门权限类型
	doorNos := make([]int, 0)
	for i := 0; i < 4; i++ {
		perm := resp[offset+i]
		// 0x00-0x04 表示有权限，0xFF 表示无权限
		if perm <= 0x04 {
			doorNos = append(doorNos, i+1)
		}
	}

	// CardFlag: 0=启用，1=禁用
	// 从门控器读取的卡默认为启用状态
	const cardFlagEnable = 0

	return &driver.Card{
		CardNo:   cardNo,
		UserName: userName,
		Password: password,
		DoorNos:  doorNos,
		CardFlag: cardFlagEnable, // 从门控器读取的卡都是启用状态
	}, nil
}

// cancelCardPermission 取消用户卡权限
// 对应协议 4.3.2：COMMAND TYPE=0x84
// DATAF格式（6字节）：1字节指引 + 5字节卡号
//
// 指引值说明：
//   - 0 = 删除用户（取消所有门权限）
//   - 1 = 取消门1权限
//   - 2 = 取消门2权限
//   - 3 = 取消门3权限
//   - 4 = 取消门4权限
//
// 返回码：
//   - 0x00 = 删除成功
//   - 0xE5 = SM内没有该用户
//   - 0xE4 = 全空
//   - 0xE3 = 不成功
func (c *Controller) cancelCardPermission(cardNo string) error {
	// 构建 DATAF（6字节）
	dataf := make([]byte, 6)

	// 字段1：指引（1字节）
	// 使用 0 表示删除用户（取消所有门权限）
	dataf[0] = 0x00

	// 字段2：卡号（5字节HEX，低字节在前）
	cardNoInt, err := strconv.ParseUint(cardNo, 10, 64)
	if err != nil {
		return fmt.Errorf("卡号解析失败: %w", err)
	}
	for i := 0; i < 5; i++ {
		dataf[1+i] = byte(cardNoInt & 0xFF)
		cardNoInt >>= 8
	}

	// 发送取消用户卡权限命令
	// CID2 = 0x49 (设置参数命令)
	// COMMAND GROUP = 0x02 (设置操作组)
	// COMMAND TYPE = 0x84 (取消用户卡权限)
	_, err = c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeCardCancel,
		dataf,
		consts2.GetRRPCCardDelete,
	)
	if err != nil {
		return fmt.Errorf("取消用户卡权限失败: %w", err)
	}

	return nil
}

// DeleteCard 删除用户卡
// 对应协议 4.3.2：COMMAND TYPE=0x84
// 直接调用 cancelCardPermission 实现
func (c *Controller) DeleteCard(cardNo string) error {
	return c.cancelCardPermission(cardNo)
}

// UpdateCard 更新用户卡信息
// CHD协议没有单独的更新命令，通过重新授权实现
// 对应协议 4.3.1：COMMAND TYPE=0x83
func (c *Controller) UpdateCard(card driver.Card) error {
	// CHD协议中，重新授权相同卡号会覆盖原有信息
	return c.AddCard(card)
}

// GetCard 获取单张卡信息
// 对应协议 5.4.4：COMMAND TYPE=0x8E
// 直接调用 GetCardByCardNo 实现
func (c *Controller) GetCard(cardNo string) (driver.Card, error) {
	card, err := c.GetCardByCardNo(cardNo)
	if err != nil {
		return driver.Card{}, err
	}
	if card == nil {
		return driver.Card{}, fmt.Errorf("卡号 %s 不存在", cardNo)
	}
	return *card, nil
}

// GetAllCards 获取所有卡片（从数据库读取启用和禁用状态的卡片）
func (c *Controller) GetAllCards() ([]driver.Card, error) {
	driverCards, err := dac.GetRW().GetDriverAllCards(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID,
		[]int{consts.DriverCardStatusEnable, consts.DriverCardStatusDisable})
	if err != nil {
		return nil, err
	}
	res := make([]driver.Card, len(driverCards))
	for i, card := range driverCards {
		res[i] = driver.Card{
			CardNo:      card.CardNo,
			CardFlag:    card.CardFlag,
			DoorNos:     card.DoorNos,
			TimeGroupNo: card.TimeGroupNo,
			UserName:    card.UserName,
			Password:    card.Password,
		}
	}
	return res, nil
}
