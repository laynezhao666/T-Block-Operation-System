package xbrother

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/driver/xbrother"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/xbrother/consts"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// GetCards 分页获取卡片列表（从数据库读取启用和禁用状态的卡片）
func (c *Controller) GetCards(offset int) (driver.CardData, error) {
	var (
		cardData = driver.CardData{}
		err      error
	)
	total, driverCards, err := dac.GetRW().GetDriverCards(
		context.Background(), c.baseInfo.ID,
		c.chanInfo.ChannelID, offset,
		consts2.DefaultLimit,
		[]int{
			consts2.DriverCardStatusEnable,
			consts2.DriverCardStatusDisable,
		})
	if err != nil {
		return cardData, fmt.Errorf("db get driver cards error, err: %s", err.Error())
	}
	cardData.Offset = int(total)
	cardData.Total = int(total)
	cardData.Cards = make([]driver.Card, len(driverCards))
	for i, card := range driverCards {
		cardData.Cards[i] = driver.Card{
			CardNo:      card.CardNo,
			CardFlag:    card.CardFlag,
			DoorNos:     card.DoorNos,
			TimeGroupNo: card.TimeGroupNo,
			UserName:    card.UserName,
			Password:    card.Password,
		}
	}
	return cardData, nil
}

// GetAllCards 获取所有卡片（启用和禁用状态）
func (c *Controller) GetAllCards() ([]driver.Card, error) {
	driverCards, err := dac.GetRW().GetDriverAllCards(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID,
		[]int{consts2.DriverCardStatusEnable, consts2.DriverCardStatusDisable})
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

// addCardInControllerByDoors 对权限下发涉及的门，向门禁发送加卡请求
func (c *Controller) addCardInControllerByDoors(req xbrother.AddCardReq, doors []int) error {
	for _, v := range doors {
		if v < 1 || v > c.doorNum {
			return fmt.Errorf("unexpected doorNo, it should be [1-%d]", c.doorNum)
		}
	}

	successDoors := make([]uint8, 0)
	for _, v := range doors {
		if _, err := c.addCard(req, uint8(v)); err != nil {
			c.logger.Errorf("add card in controller error: %v, cardNo: %d, doorNo: %d",
				err, req.CardId, v)
			break
		}
		successDoors = append(successDoors, uint8(v))
		time.Sleep(consts2.DurationSleepTime)
	}
	if len(successDoors) != len(doors) {
		// 尝试一次恢复
		for _, v := range successDoors {
			if _, err := c.deleteCard(xbrother.DeleteCardReq{CardIndex: req.CardIndex}, v); err != nil {
				c.logger.Errorf("restore add card error: %v, cardNo: %d, doorNo: %d",
					err, req.CardId, v)
				return err
			}
			time.Sleep(consts2.DurationSleepTime)
		}
		return fmt.Errorf("add card in controller by doors error, success doors: %v, target doors: %v",
			successDoors, doors)
	}
	return nil
}

// deleteCardInControllerByDoors 对权限下发涉及的门，向门禁发送删卡请求，失败时尝试恢复
func (c *Controller) deleteCardInControllerByDoors(req xbrother.AddCardReq, doors []int) error {
	for _, v := range doors {
		if v < 1 || v > c.doorNum {
			return fmt.Errorf("unexpected doorNo, it should be [1-%d]", c.doorNum)
		}
	}

	successDoors := make([]uint8, 0)
	for _, v := range doors {
		if _, err := c.deleteCard(xbrother.DeleteCardReq{CardIndex: req.CardIndex}, uint8(v)); err != nil {
			c.logger.Errorf("delete card in controller error: %v, channelID: %s, cardNo: %d, doorNo: %d",
				err, c.chanInfo.ChannelID, req.CardId, v)
			break
		}
		successDoors = append(successDoors, uint8(v))
		time.Sleep(consts2.DurationSleepTime)
	}
	if len(successDoors) != len(doors) {
		// 尝试一次恢复
		for _, v := range successDoors {
			if _, err := c.addCard(req, v); err != nil {
				c.logger.Errorf("restore delete card error: %v, channelID: %s, cardNo: %d, doorNo: %d",
					err, c.chanInfo.ChannelID, req.CardId, v)
				return err
			}
			time.Sleep(consts2.DurationSleepTime)
		}
		return fmt.Errorf("delete card in controller by doors error")
	}
	return nil
}

// addCardByUpdateOldCard 通过更新已有卡片记录来添加卡片（复用旧卡的CardIndex）
func (c *Controller) addCardByUpdateOldCard(tx *gorm.DB, card driver.Card, oldCard db.DriverCard) error {
	req, status, err := c.buildCardReq(
		card.CardNo, card.Password, card.CardFlag,
		card.TimeGroupNo, card.DoorNos,
		int(oldCard.CardIndex))
	if err != nil {
		return err
	}

	dbCard := db.DriverCard{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		CardNo:       card.CardNo,
		CardFlag:     card.CardFlag,
		DoorNos:      card.DoorNos,
		TimeGroupNo:  card.TimeGroupNo,
		UserName:     card.UserName,
		Password:     card.Password,
		CardIndex:    oldCard.CardIndex,
		Status:       int(status),
	}

	if err = c.addCardInControllerByDoors(req, card.DoorNos); err != nil {
		return err
	}
	if err = dac.UpdateDriverCard(tx, c.baseInfo.ID, oldCard, dbCard); err != nil {
		// 尝试一次恢复
		c.logger.Errorf("db add driver card error: %v, cardNo: %s", err, card.CardNo)
		if err = c.deleteCardInControllerByDoors(req, card.DoorNos); err != nil {
			c.logger.Errorf("controller add card success, db add card error: %v, "+
				"try to recover error, cardId: %s", err, card.CardNo)
		}
		return err
	}
	return nil
}

// addCardByNewCard 创建新卡片记录并下发到门控器
func (c *Controller) addCardByNewCard(tx *gorm.DB, card driver.Card) error {
	nextCardIndex := consts2.DefaultStartCardIndex
	if lastIndex, err := c.getLastCardIndex(); err == nil {
		nextCardIndex = lastIndex + 1
	}
	if uint32(nextCardIndex) >= consts2.CardIndexMax {
		return fmt.Errorf("add card error, card num out of range")
	}

	req, status, err := c.buildCardReq(
		card.CardNo, card.Password, card.CardFlag,
		card.TimeGroupNo, card.DoorNos, nextCardIndex)
	if err != nil {
		return err
	}

	dbCard := db.DriverCard{
		ControllerID: c.baseInfo.ID,
		ChannelID:    c.chanInfo.ChannelID,
		CardNo:       card.CardNo,
		CardFlag:     card.CardFlag,
		DoorNos:      card.DoorNos,
		TimeGroupNo:  card.TimeGroupNo,
		UserName:     card.UserName,
		Password:     card.Password,
		CardIndex:    nextCardIndex,
		Status:       int(status),
	}

	if err = c.addCardInControllerByDoors(req, card.DoorNos); err != nil {
		return err
	}
	if err = dac.AddDriverCard(tx, dbCard); err != nil {
		c.logger.Errorf("db update driver card error: %v", err)
		// 尝试一次恢复
		if err = c.deleteCardInControllerByDoors(req, card.DoorNos); err != nil {
			c.logger.Errorf("controller add card success, db add card error: %v, "+
				"try to recover error, cardId: %s", err, card.CardNo)
		}
		return err
	}
	return nil
}

// AddCard 添加卡片（先查询是否存在旧卡，存在则更新，否则新建）
func (c *Controller) AddCard(card driver.Card) error {
	return dac.GetRW().AddDriverCard(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, card,
		c.addCardByNewCard,
		c.addCardByUpdateOldCard)
}

// transferCardId 将卡号字符串转换为uint32类型
func transferCardId(cardIdStr string) (uint32, error) {
	cardId, err := strconv.Atoi(cardIdStr)
	if err != nil {
		return 0, fmt.Errorf("cardIdStr should be int, cardIdStr: %s, err: %w", cardIdStr, err)
	}
	return uint32(cardId), nil
}

// transferAccessTimeGroup 根据门控器门数类型，将时间组号和门号列表转换为协议要求的权限位图
func transferAccessTimeGroup(timeGroupNo int, doorNos []int, doorNum int) (uint32, error) {
	switch doorNum {
	case consts.OneDoorPerController:
		return transferAccessTimeGroupOneDoor(timeGroupNo, doorNos)
	case consts.TwoDoorPerController:
		return transferAccessTimeGroupTwoDoor(timeGroupNo, doorNos)
	case consts.FourDoorPerController:
		return transferAccessTimeGroupFourDoor(timeGroupNo, doorNos)
	default:
		return 0, fmt.Errorf("unknown doorNum")
	}
}

// transferAccessTimeGroupOneDoor 单门控制器的时间组权限位图转换（时间组号范围0-15）
func transferAccessTimeGroupOneDoor(timeGroupNo int, doorNos []int) (uint32, error) {
	if timeGroupNo >= consts2.OneDoorMaxTimeGroupNum {
		return 0, fmt.Errorf("one door timeGroupNo should be 0-15")
	}
	enableDoorNos := make([]bool, consts.OneDoorPerController)
	for _, doorNo := range doorNos {
		if doorNo > 0 && doorNo <= consts.OneDoorPerController {
			enableDoorNos[doorNo-1] = true
		}
	}
	// 这里处理字节序，共济协议中门权限为小端序
	var door1TimeGroup16Bit uint16 = 0
	if enableDoorNos[0] {
		door1TimeGroup16Bit |= 1 << (timeGroupNo + 8)
	}
	// 单门后两字节填0
	var res uint32 = 0
	res |= uint32(door1TimeGroup16Bit) << 16
	return res, nil
}

// transferAccessTimeGroupTwoDoor 双门控制器的时间组权限位图转换（时间组号范围0-15）
func transferAccessTimeGroupTwoDoor(timeGroupNo int, doorNos []int) (uint32, error) {
	if timeGroupNo >= consts2.TwoDoorMaxTimeGroupNum {
		return 0, fmt.Errorf("two door timeGroupNo should be 0-15")
	}
	enableDoorNos := make([]bool, consts.TwoDoorPerController)
	for _, doorNo := range doorNos {
		if doorNo > 0 && doorNo <= consts.TwoDoorPerController {
			enableDoorNos[doorNo-1] = true
		}
	}
	// 这里处理字节序，共济协议中门权限为小端序
	var door1TimeGroup16Bit uint16 = 0
	if enableDoorNos[0] {
		door1TimeGroup16Bit |= 1 << (timeGroupNo + 8)
	}
	var door2TimeGroup16bit uint16 = 0
	if enableDoorNos[1] {
		door2TimeGroup16bit |= 1 << (timeGroupNo + 8)
	}
	// 前两字节为door1，后两字节为door2
	var res uint32 = 0
	res |= uint32(door1TimeGroup16Bit) << 16
	res |= uint32(door2TimeGroup16bit)
	return res, nil
}

// transferAccessTimeGroupFourDoor 四门控制器的时间组权限位图转换（时间组号范围0-7）
func transferAccessTimeGroupFourDoor(timeGroupNo int, doorNos []int) (uint32, error) {
	if timeGroupNo >= consts2.FourDoorMaxTimeGroupNum {
		return 0, fmt.Errorf("four door timeGroupNo should be 0-7")
	}
	enableDoorNos := make([]bool, consts.FourDoorPerController)
	for _, doorNo := range doorNos {
		if doorNo > 0 && doorNo <= consts.FourDoorPerController {
			enableDoorNos[doorNo-1] = true
		}
	}

	var door1TimeGroup8Bit uint8 = 0
	if enableDoorNos[0] {
		door1TimeGroup8Bit |= 1 << timeGroupNo
	}
	var door2TimeGroup8Bit uint8 = 0
	if enableDoorNos[1] {
		door2TimeGroup8Bit |= 1 << timeGroupNo
	}
	var door3TimeGroup8Bit uint8 = 0
	if enableDoorNos[2] {
		door3TimeGroup8Bit |= 1 << timeGroupNo
	}
	var door4TimeGroup8Bit uint8 = 0
	if enableDoorNos[3] {
		door4TimeGroup8Bit |= 1 << timeGroupNo
	}
	var res uint32 = 0
	res |= uint32(door1TimeGroup8Bit) << 24
	res |= uint32(door2TimeGroup8Bit) << 16
	res |= uint32(door3TimeGroup8Bit) << 8
	res |= uint32(door4TimeGroup8Bit)
	return res, nil
}

// UpdateCard 更新卡片（底层复用AddCard逻辑）
func (c *Controller) UpdateCard(card driver.Card) error {
	return c.AddCard(card)
}

// GetCard 根据卡号获取单张卡片信息
func (c *Controller) GetCard(cardNo string) (driver.Card, error) {
	return c.getCardByCardNo(cardNo)
}

// getCardByCardNo 从数据库根据卡号查询卡片并转换为驱动模型
func (c *Controller) getCardByCardNo(cardNo string) (driver.Card, error) {
	driverCard, err := dac.GetRW().GetDriverCard(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID, cardNo)
	if err != nil {
		return driver.Card{}, err
	}
	return driver.Card{
		CardNo:      driverCard.CardNo,
		CardFlag:    driverCard.CardFlag,
		DoorNos:     driverCard.DoorNos,
		TimeGroupNo: driverCard.TimeGroupNo,
		UserName:    driverCard.UserName,
		Password:    driverCard.Password,
	}, nil
}

// DeleteCard 删除卡片（先从门控器删除，再逻辑删除数据库记录）
func (c *Controller) DeleteCard(cardNo string) error {
	return dac.GetRW().LogicDeleteDriverCard(
		context.Background(), c.baseInfo.ID,
		c.chanInfo.ChannelID, cardNo,
		func(driverCard db.DriverCard) error {
			req, _, err := c.buildCardReq(
				driverCard.CardNo, driverCard.Password,
				driverCard.CardFlag, driverCard.TimeGroupNo,
				driverCard.DoorNos, int(driverCard.CardIndex))
			if err != nil {
				return err
			}

			if err = c.deleteCardInControllerByDoors(req, driverCard.DoorNos); err != nil {
				c.logger.Errorf("controller delete card error: %v, cardNo: %s", err, driverCard.CardNo)
				return err
			}
			return nil
		})
}

// clearCards 清空门控器上指定门的所有卡片
func (c *Controller) clearCards(req xbrother.ClearCardsReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts2.GetRRPCClearCards(c.chanInfo.ChannelID), consts2.CommandClearCards)
}

// deleteCard 从门控器上删除指定门的单张卡片
func (c *Controller) deleteCard(req xbrother.DeleteCardReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts2.GetRRPCDeleteCard(c.chanInfo.ChannelID), consts2.CommandDeleteCard)
}

// addCard 向门控器上指定门添加单张卡片
func (c *Controller) addCard(req xbrother.AddCardReq, doorNo uint8) (xbrother.CommonResp, error) {
	return c.sendRequest(req, doorNo, consts2.GetRRPCAddCard(c.chanInfo.ChannelID), consts2.CommandAddCard)
}

// buildCardReq 构建卡片请求的公共逻辑：密码转换、门数获取、时间组转换、卡号转换、状态转换
func (c *Controller) buildCardReq(cardNo, password string,
	cardFlag, timeGroupNo int, doorNos []int, cardIndex int,
) (xbrother.AddCardReq, uint8, error) {
	pass, err := c.transferPasswordToUint16(password)
	if err != nil {
		return xbrother.AddCardReq{}, 0, err
	}

	doorNum, err := c.GetDoorNumber()
	if err != nil {
		return xbrother.AddCardReq{}, 0, fmt.Errorf("get door number failed, err: %w", err)
	}

	accessTimeGroup, err := transferAccessTimeGroup(timeGroupNo, doorNos, doorNum)
	if err != nil {
		return xbrother.AddCardReq{}, 0, err
	}

	cardId, err := transferCardId(cardNo)
	if err != nil {
		return xbrother.AddCardReq{}, 0, err
	}

	var status uint8 = 0
	switch cardFlag {
	case consts2.DriverCardStatusEnable:
		status = consts2.ControllerCardStatusEnable
	case consts2.DriverCardStatusDisable:
		status = consts2.ControllerCardStatusDisable
	default:
		return xbrother.AddCardReq{}, 0, fmt.Errorf("unknown card flag")
	}

	req := xbrother.AddCardReq{
		CardIndex:       uint16(cardIndex),
		CardId:          cardId,
		Password:        pass,
		AccessTimeGroup: accessTimeGroup,
		Status:          status,
	}
	return req, status, nil
}

// transferPasswordToUint16 将密码字符串转换为BCD编码的uint16（4位数字，首位不为0）
func (c *Controller) transferPasswordToUint16(password string) (uint16, error) {
	iPassword, err := strconv.Atoi(password)
	if err != nil {
		return 0, fmt.Errorf("wrong password format: %v", password)
	}

	// xbrother 协议规定：卡密码位4位数字，且不能以0开头，所以默认取数字的后4位做密码，判断不能以0开头
	iPassword = iPassword % 10000
	if iPassword < consts2.PasswordLowerBound {
		return 0, fmt.Errorf("wrong password format, it should be 4 digit, first digit is not 0, password: %v", password)
	}
	return utils.FourDigitIntToBCDUint16(iPassword), nil
}
