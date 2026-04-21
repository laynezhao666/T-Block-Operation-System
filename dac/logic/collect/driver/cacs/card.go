package cacs

import (
	"fmt"
	"strconv"

	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"

	"dac/entity/utils/rrpc"
)

// 卡号查询/删除类型常量
var (
	GetCardByUserIdType    uint8 = 0 // 按用户ID查询卡
	GetCardByCardNoType    uint8 = 1 // 按卡号查询卡
	DeleteCardByCardNoType uint8 = 0 // 按卡号删除卡
	DeleteCardByUserIdType uint8 = 1 // 按用户ID删除卡
	DeleteCardAllType      uint8 = 2 // 删除所有卡
)

// haveAuthForAccess 判断门授权信息中是否有进入权限。
// 通过检查 PermitPeriod 的 D7 D6 位判断：01或10表示有权限。
func haveAuthForAccess(doorAuth cacs.DoorAuth) bool {
	first2Bit2 := doorAuth.PermitPeriod & 0xc0
	// 如果D7 D6位为01或者11， 就表示有权限进入此门
	if first2Bit2 == 0x40 || first2Bit2 == 0x80 {
		return true
	}
	return false
}

// getDoorNos 从卡信息中提取有授权的门编号列表。
func getDoorNos(cardInfo cacs.CardInfo) []int {
	res := make([]int, 0)
	if haveAuthForAccess(cardInfo.AuthDoor1) {
		res = append(res, 1)
	}
	if haveAuthForAccess(cardInfo.AuthDoor2) {
		res = append(res, 2)
	}
	if haveAuthForAccess(cardInfo.AuthDoor3) {
		res = append(res, 3)
	}
	if haveAuthForAccess(cardInfo.AuthDoor4) {
		res = append(res, 4)
	}
	return res
}

// getTimeGroupNoFromDoorAuth 从门授权信息中提取时间组编号。
// D7D6=01时从低4位读取星期准进列表序号，D7D6=11时不受时间限制返回0。
func getTimeGroupNoFromDoorAuth(doorAuth cacs.DoorAuth) int {
	first2Bits := doorAuth.PermitPeriod & 0xc0
	switch first2Bits {
	case 0x40:
		// D7 D6位为0 1,用户受节假日准进表和星期准进表限制 其中D5,D4两位表示节假日准进列表序号(0-3), D3,D2,D1,D0四位表示星期准进列表序号(0-7)
		return int(doorAuth.PermitPeriod & 0x0f)
	case 0xc0:
		// D7 D6位为1 1, 用户不受准进时间段限制; 没有对应的时间组，暂时返回0
		return 0
	default:
		// 不会出现这种情况
		return -1
	}
}

// getTimeGroupNo 从卡信息中获取统一的时间组编号。
// 要求所有有授权的门使用相同的时间组，否则返回错误。
func getTimeGroupNo(cardInfo cacs.CardInfo) (int, error) {
	// 收集所有有授权的门的时间组编号
	doorAuths := []cacs.DoorAuth{cardInfo.AuthDoor1, cardInfo.AuthDoor2, cardInfo.AuthDoor3, cardInfo.AuthDoor4}
	timeGroupNo := -1

	for _, doorAuth := range doorAuths {
		// 只检查有授权的门
		if !haveAuthForAccess(doorAuth) {
			continue
		}
		tgNo := getTimeGroupNoFromDoorAuth(doorAuth)
		if timeGroupNo == -1 {
			// 第一个有授权的门，记录其时间组
			timeGroupNo = tgNo
		} else if timeGroupNo != tgNo {
			// 后续有授权的门，时间组必须一致
			return -1, fmt.Errorf("门禁卡时间组不一致")
		}
	}

	// 如果没有任何门有授权，返回0作为默认值
	if timeGroupNo == -1 {
		return 0, nil
	}
	return timeGroupNo, nil
}

// GetCard 根据卡号从门控器查询单张卡信息。
func (c *Controller) GetCard(cardNo string) (driver.Card, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.Card{}, err
	}
	cardId, err := strconv.Atoi(cardNo)
	if err != nil {
		return driver.Card{}, err
	}
	resp, ok, packetRtn, _, err := c.getCards(cacs.GetCardsReq{
		Type: GetCardByCardNoType,
		Id:   uint32(cardId),
	})
	if !ok {
		return driver.Card{}, fmt.Errorf("get cards failed, err: %s", err.Error())
	}
	if packetRtn != consts.KRtnNormal {
		return driver.Card{}, fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	// cardNo对应的Card，最多只有一个. CACS协议中规定CardNum范围为(1-4)
	if resp.CardNum > 1 {
		return driver.Card{}, fmt.Errorf("card not unique, cardNo: %s", cardNo)
	}
	card := resp.Card1
	timeGroupNo, err := getTimeGroupNo(card)
	if err != nil {
		return driver.Card{}, err
	}
	return driver.Card{
		CardNo:      strconv.Itoa(int(card.Id)),
		CardFlag:    0, // 无法获取，默认置为0
		DoorNos:     getDoorNos(card),
		TimeGroupNo: timeGroupNo,
		UserName:    strconv.Itoa(int(card.UserId)),
		Password:    strconv.Itoa(int(card.Password)),
	}, nil
}

// getDoorAuth 根据时间组编号构建门授权信息。
// 设置 D7D6=01 表示受准进时段限制，授权方式为刷卡。
func getDoorAuth(timeGroupNo int) cacs.DoorAuth {
	var permitPeriod uint8 = 0
	// 设置D7 D6为01， 表示该用户受准进时段限制 D5 D4 为00，节假日准进列表无法获得，默认将节假日准进列表置为0
	permitPeriod = permitPeriod | 0x40
	// 设置D3 D2 D1 D0 表示星期准进列表序号
	permitPeriod = permitPeriod | uint8(timeGroupNo)

	var doorAuth cacs.DoorAuth
	doorAuth.PermitPeriod = permitPeriod
	// 设置授权方式为刷卡
	doorAuth.AuthType[0] = doorAuth.AuthType[0] | 0x01
	return doorAuth
}

// AddCard 向门控器下载一张卡信息。
// 根据 CardFlag 设置有效期：0=正常（1980-2079），1=禁用。
func (c *Controller) AddCard(card driver.Card) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	cardNo, err := strconv.Atoi(card.CardNo)
	if err != nil {
		return fmt.Errorf("driver.Card.CardIndex can not converse to int")
	}
	username, err := strconv.Atoi(card.UserName)
	if err != nil {
		return fmt.Errorf("driver.Card.UserName can not converse to int")
	}
	password, err := strconv.Atoi(card.Password)
	if err != nil {
		return fmt.Errorf("driver.Card.Password can not converse to int")
	}

	cardInfo := cacs.CardInfo{
		Id:       uint32(cardNo),
		UserId:   uint32(username),
		Password: uint32(password),
		CardType: 0, // 0表示普通卡，1表示胁迫卡
		AreaId:   0, // 0或者全F表示此信息为无效
	}
	if card.CardFlag == 0 {
		// 正常 开始有效期和结束有效期的数值没有特殊含义
		cardInfo.StartYear = 1980
		cardInfo.StartMonth = 1
		cardInfo.StartDay = 1
		cardInfo.StartHour = 0
		cardInfo.StartMinute = 0
		cardInfo.StartSecond = 0
		cardInfo.EndYear = 2079
		cardInfo.EndMonth = 12
		cardInfo.EndDay = 31
		cardInfo.EndHour = 23
		cardInfo.EndMinute = 59
		cardInfo.EndSecond = 59
	} else if card.CardFlag == 1 {
		// 禁用 开始有效期和结束有效期的数值没有特殊含义
		cardInfo.StartYear = 1980
		cardInfo.StartMonth = 1
		cardInfo.StartDay = 1
		cardInfo.StartHour = 0
		cardInfo.StartMinute = 0
		cardInfo.StartSecond = 0
		cardInfo.EndYear = 1980
		cardInfo.EndMonth = 1
		cardInfo.EndDay = 1
		cardInfo.EndHour = 0
		cardInfo.EndMinute = 0
		cardInfo.EndSecond = 1
	} else {
		return fmt.Errorf("unknown driver.Card")
	}
	for i := range card.DoorNos {
		doorNo := card.DoorNos[i]
		if doorNo == 1 {
			cardInfo.AuthDoor1 = getDoorAuth(card.TimeGroupNo)
		} else if doorNo == 2 {
			cardInfo.AuthDoor2 = getDoorAuth(card.TimeGroupNo)
		} else if doorNo == 3 {
			cardInfo.AuthDoor3 = getDoorAuth(card.TimeGroupNo)
		} else if doorNo == 4 {
			cardInfo.AuthDoor4 = getDoorAuth(card.TimeGroupNo)
		} else {
			return fmt.Errorf("doorNo is invalid, it should be 1-4")
		}
	}
	resp, ok, packetRtn, _, err := c.downloadCards(cacs.DownloadCardsReq{
		Num:   1,
		Cards: []cacs.CardInfo{cardInfo},
	})
	if !ok {
		return fmt.Errorf("downloadCards failed: %s", err.Error())
	}
	if packetRtn != consts.KRtnNormal {
		return fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	if resp.SuccessNum == 1 {
		return nil
	}
	return fmt.Errorf("Add Card失败")
}

// UpdateCard 更新门控器中的卡信息，先删除再添加。
func (c *Controller) UpdateCard(card driver.Card) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	err := c.DeleteCard(card.CardNo)
	if err != nil {
		return fmt.Errorf("delete card failed: %s", err.Error())
	}
	return c.AddCard(card)
}

// DeleteCard 从门控器中删除指定卡号的卡。
func (c *Controller) DeleteCard(cardNo string) error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	cardId, err := strconv.Atoi(cardNo)
	if err != nil {
		return err
	}
	_, ok, packetRtn, _, err := c.deleteCards(cacs.DeleteCardsReq{
		Type: DeleteCardByCardNoType,
		Id:   uint32(cardId),
	})
	if !ok {
		return fmt.Errorf("delete card failed: %s", err.Error())
	}
	if packetRtn != consts.KRtnNormal {
		return fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	return nil
}

// GetAllCards 从门控器中读取所有卡信息，通过分页遍历实现。
func (c *Controller) GetAllCards() ([]driver.Card, error) {
	if _, err := c.checkConnection(); err != nil {
		return nil, err
	}
	cards := make([]driver.Card, 0)
	var index uint32 = 0
	for {
		resp, ok, packetRtn, _, err := c.getCardsInfo(cacs.GetCardsInfoReq{Index: index})
		if !ok {
			return nil, fmt.Errorf("getCardsInfo failed: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return nil, fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
		// 当返回的人员信息个数为0时表示已经读完
		if resp.Num == 0 {
			break
		}
		for i := range resp.Cards {
			cacsCard := resp.Cards[i]
			timeGroupNo, err := getTimeGroupNo(cacsCard)
			if err != nil {
				return nil, err
			}
			cards = append(cards, driver.Card{
				CardNo:      strconv.Itoa(int(cacsCard.Id)),
				CardFlag:    0,
				DoorNos:     getDoorNos(cacsCard),
				TimeGroupNo: timeGroupNo,
				UserName:    strconv.Itoa(int(cacsCard.UserId)),
				Password:    strconv.Itoa(int(cacsCard.Password)),
			})
		}
		index = resp.NextIndex
	}
	return cards, nil

}

// GetCards 从指定偏移量开始分页读取卡信息。
func (c *Controller) GetCards(offset int) (driver.CardData, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.CardData{}, err
	}
	var cardData driver.CardData
	cardData.Cards = make([]driver.Card, 0)
	var index uint32 = uint32(offset)
	for {
		resp, ok, packetRtn, _, err := c.getCardsInfo(cacs.GetCardsInfoReq{Index: index})
		if !ok {
			return driver.CardData{}, fmt.Errorf("get cards error: %s", err.Error())
		}
		if packetRtn != consts.KRtnNormal {
			return driver.CardData{}, fmt.Errorf(consts.RtnInfoMap[packetRtn])
		}
		if resp.Num == 0 {
			break
		}
		for i := range resp.Cards {
			cacsCard := resp.Cards[i]
			timeGroupNo, err := getTimeGroupNo(cacsCard)
			if err != nil {
				return driver.CardData{}, err
			}
			cardData.Cards = append(cardData.Cards, driver.Card{
				CardNo:      strconv.Itoa(int(cacsCard.Id)),
				CardFlag:    0,
				DoorNos:     nil,
				TimeGroupNo: timeGroupNo,
				UserName:    strconv.Itoa(int(cacsCard.UserId)),
				Password:    strconv.Itoa(int(cacsCard.Password)),
			})
		}
		index = resp.NextIndex
		continue
	}
	cardData.Offset = offset + len(cardData.Cards)
	cardData.Total = offset + len(cardData.Cards)
	return cardData, nil
}

// downloadCards 向门控器下发卡信息的底层通信方法。
func (c *Controller) downloadCards(
	req cacs.DownloadCardsReq,
) (cacs.DownloadCardsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DownloadCardsResp{}, false, 0, consts.KRequestError, err
	}

	cmd := consts.KCommandRequestDownloadCards
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCDownloadCards(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(consts.KCommandResponseDownloadCards, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to DownloadCardsResp failed, err: %v", err)
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp tcpUnmarshal to DownloadCardsResp failed, err: %v", err)
	}
	downloadCardsInfoResp, ok := resp.(cacs.DownloadCardsResp)
	if !ok {
		c.Errorf("resp type error, it should be DownloadCardsResp")
		return cacs.DownloadCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, it should be DownloadCardsResp")
	}
	return downloadCardsInfoResp, true, server.p.rtn, consts.KNormal, nil
}

// getCards 从门控器查询卡信息的底层通信方法。
func (c *Controller) getCards(
	req cacs.GetCardsReq,
) (cacs.GetCardsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.GetCardsResp{}, false, 0, consts.KRequestError, err
	}

	cmd := consts.KCommandRequestGetCards
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCGetCards(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(consts.KCommandResponseGetCards, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to GetCardsResp failed, err: %v", err)
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp tcpUnmarshal to GetCardsResp failed, err: %v", err)
	}
	getCardsInfoResp, ok := resp.(cacs.GetCardsResp)
	if !ok {
		c.Errorf("resp type error, it should be GetCardsResp")
		return cacs.GetCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp type error, it should be GetCardsResp")
	}
	return getCardsInfoResp, true, server.p.rtn, consts.KNormal, nil
}

// deleteCards 从门控器删除卡信息的底层通信方法。
func (c *Controller) deleteCards(
	req cacs.DeleteCardsReq,
) (cacs.DeleteCardsResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.DeleteCardsResp{}, false, 0, consts.KRequestError, err
	}

	cmd := consts.KCommandRequestDeleteCards
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCDeleteCards(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(consts.KCommandResponseDeleteCards, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to DeleteCardsResp failed, err: %v", err)
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp tcpUnmarshal to DeleteCardsResp failed, err: %v", err)
	}
	deleteCardsResp, ok := resp.(cacs.DeleteCardsResp)
	if !ok {
		c.Errorf("resp type error, it should be DeleteCardsResp")
		return cacs.DeleteCardsResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("resp type error, it should be DeleteCardsResp")
	}
	return deleteCardsResp, true, server.p.rtn, consts.KNormal, nil
}

// getCardsInfo 从门控器批量读取卡信息的底层通信方法。
func (c *Controller) getCardsInfo(
	req cacs.GetCardsInfoReq,
) (cacs.GetCardsInfoResp, bool, uint32, int, error) {
	server, err := c.checkConnection()
	if err != nil {
		return cacs.GetCardsInfoResp{}, false, 0, consts.KRequestError, err
	}

	cmd := consts.KCommandRequestGetCardsInfo
	data, err := c.tcpMarshal.Marshal(cmd, req)
	if err != nil {
		c.Errorf("req marshal failed, err: %v", err)
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KMarshalError,
			fmt.Errorf("req marshal failed, err: %v", err)
	}
	if server.Request(cmd, data) < 0 {
		c.Errorf("req data send failed, err: %v", err)
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KRequestError,
			fmt.Errorf("req data send failed, err: %v", err)
	}
	rrpcKey := consts.GetRRPCGetCardsInfo(c.chanInfo.ChannelID)
	respRaw, ok := rrpc.Manager().Get(rrpcKey, c.timeout)
	if !ok {
		c.Errorf("rrpc get resp timeout")
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("rrpc get resp timeout")
	}
	bytes, ok := respRaw.([]byte)
	if !ok {
		c.Errorf("respRaw converse to []byte failed, err: %v", err)
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("respRaw converse to []byte failed, err: %v", err)
	}
	resp, err := c.tcpMarshal.Unmarshal(consts.KCommandResponseGetCardsInfo, bytes)
	if err != nil {
		c.Errorf("resp tcpUnmarshal to DeleteCardsResp failed, err: %v", err)
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KUnMarshalError,
			fmt.Errorf("resp tcpUnmarshal to GetCardsInfoResp failed, err: %v", err)
	}
	getCardsInfoResp, ok := resp.(cacs.GetCardsInfoResp)
	if !ok {
		c.Errorf("resp type error, it should be DeleteCardsResp")
		return cacs.GetCardsInfoResp{}, false, server.p.rtn,
			consts.KRecvRespError,
			fmt.Errorf("resp type error, it should be GetCardsInfoResp")
	}
	return getCardsInfoResp, true, server.p.rtn, consts.KNormal, nil
}

// DeleteAllCards 删除门控器中的所有卡信息。
func (c *Controller) DeleteAllCards() error {
	if _, err := c.checkConnection(); err != nil {
		return err
	}
	_, ok, packetRtn, _, err := c.deleteCards(cacs.DeleteCardsReq{
		Type: DeleteCardAllType,
		Id:   0,
	})
	if !ok {
		return fmt.Errorf("delete all cards failed, err: %s", err.Error())
	}
	if packetRtn != consts.KNormal {
		return fmt.Errorf(consts.RtnInfoMap[packetRtn])
	}
	return nil
}
