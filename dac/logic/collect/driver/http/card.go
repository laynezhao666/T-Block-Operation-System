// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/entity/utils/dhttp"

	"dac/entity/utils/parse"
	"dac/entity/utils/thttp"
)

// tryParseCard 尝试多种方式解析卡片数据（兼容不同协议版本）
func (c *Controller) tryParseCard(b []byte) (driver.Card, error) {
	var card driver.Card
	err := json.Unmarshal(b, &card)
	if err == nil {
		return card, nil
	}

	if config.C.VerifyMode && !config.C.IsEnableCompatible() {
		return card, err
	}

	c.filterLogger.Warnf("tryParseCard1", "get card failed, parse as object error: %v, try parse as array", err)
	// 门禁协议 v1.0.9.4 自身存在矛盾。
	// 数据输出定义中，data 为 array，但在数据输出示例中，data 为 object。
	// 此处尝试兼容
	var temp []driver.Card
	if err = json.Unmarshal(b, &temp); err == nil {
		if len(temp) == 1 {
			return temp[0], nil
		}

		return card, fmt.Errorf("length of cards error: %+v", temp)
	}

	// 尝试将卡号解析为整数类型（兼容旧版协议）
	var intCard driver.CardWithIntNo
	c.filterLogger.Warnf("tryParseCard2", "get card failed, parse as array error: %v, try parse card number as int", err)
	if err = json.Unmarshal(b, &intCard); err == nil {
		card = driver.Card{
			CardNo:      strconv.FormatInt(intCard.CardNo, 10),
			CardFlag:    intCard.CardFlag,
			DoorNos:     intCard.DoorNos,
			TimeGroupNo: intCard.TimeGroupNo,
			UserName:    intCard.UserName,
			Password:    intCard.Password,
		}
		return card, nil
	}

	return card, err

}

// GetCard 根据卡号获取单张卡片信息
func (c *Controller) GetCard(cardNo string) (driver.Card, error) {
	var (
		card driver.Card
		e    error
	)

	url := c.urlProducer.GetCardURL(cardNo)
	err := dhttp.GetJSONWithParseFunc(url, c.timeout, func(b []byte) error {
		card, e = c.tryParseCard(b)
		return e
	})
	return card, err
}

// AddCard 添加卡片到控制器
func (c *Controller) AddCard(card driver.Card) error {
	return c.postJSON(c.urlProducer.AddCardURL(), card, nil)
}

// UpdateCard 更新控制器上的卡片信息
func (c *Controller) UpdateCard(card driver.Card) error {
	return c.postJSON(c.urlProducer.UpdateCardURL(), card, nil)
}

// DeleteCard 从控制器删除指定卡片
func (c *Controller) DeleteCard(cardNo string) error {
	var req struct {
		Card string `json:"card_no"`
	}
	req.Card = cardNo

	b, err := c.getBody(req)
	if err != nil {
		return err
	}

	// 忽略 code 中的错误
	deleteURL := c.urlProducer.DeleteCardURL()
	_, err = thttp.Request(
		deleteURL, http.MethodPost,
		dhttp.FormContentHeader, b,
		int(c.timeout.Milliseconds()))
	return err
}

// GetAllCards 获取控制器上的所有卡片
func (c *Controller) GetAllCards() ([]driver.Card, error) {
	var cards []driver.Card
	url := c.urlProducer.GetAllCardsURL()

	err := dhttp.GetJSON(url, c.timeout, &cards)
	return cards, err
}

// GetCards 分页获取卡片列表
func (c *Controller) GetCards(offset int) (driver.CardData, error) {
	url := c.urlProducer.GetCardsURL(offset)

	var resp struct {
		Code    int         `json:"err_code"`
		Message string      `json:"err_msg"`
		Offset  int         `json:"card_last_get"`
		Total   int         `json:"card_count"`
		Data    interface{} `json:"data"`
	}
	b, err := thttp.Request(url, http.MethodGet, nil, nil, int(c.timeout.Milliseconds()))
	if err != nil {
		return driver.CardData{}, err
	}
	if err = json.Unmarshal(b, &resp); err != nil {
		return driver.CardData{}, err
	}
	if resp.Code != 0 {
		return driver.CardData{}, fmt.Errorf("code != 0, response: %+v", utils.GetJSONString(resp))
	}

	var cards []driver.Card
	if err = parse.JSON(&cards, resp.Data); err != nil {
		return driver.CardData{}, fmt.Errorf("parse error: %w, response: %+v", err, utils.GetJSONString(resp))
	}

	data := driver.CardData{
		Offset: resp.Offset,
		Total:  resp.Total,
		Cards:  cards,
	}
	return data, nil
}
