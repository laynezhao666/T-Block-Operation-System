// Package access 提供门禁权限管理的HTTP接口处理器，包括卡片、人员、时间组和权限组。
package access

import (
	"dac/entity/config"
	"fmt"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"dac/entity/model/cgi"
	"dac/entity/model/db"
	"dac/entity/utils"
	"dac/logic/cgi/card"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// GetCards 查询门禁卡列表，支持分页、模糊搜索和多条件过滤
func GetCards(c *gin.Context) {
	var (
		req struct {
			Offset           int      `json:"offset"`
			Limit            int      `json:"limit"`
			Cards            []string `json:"cards"`
			Query            string   `json:"query"`
			CardType         int      `json:"card_type"`
			QueryCardType    bool     `json:"query_card_type"`
			CardFlag         int      `json:"card_flag"`
			QueryCardFlag    bool     `json:"query_card_flag"`
			AccessGroup      int      `json:"access_group"`
			QueryAccessGroup bool     `json:"query_access_group"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	n, cards, err := card.GetCards(c, utils.GetMozuID(c), req.Offset, req.Limit, req.Cards, req.Query,
		db.CardType(req.CardType), req.QueryCardType, db.CardFlagType(req.CardFlag), req.QueryCardFlag,
		req.AccessGroup, req.QueryAccessGroup)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var resp struct {
		Total int        `json:"total"`
		List  []cgi.Card `json:"list"`
	}
	resp.Total = int(n)
	resp.List = cards
	ghttp.SendResponseWithData(c, resp)
}

// AddCard 新增门禁卡，自动清理卡号中的空格和前导零
func AddCard(c *gin.Context) {
	var (
		req struct {
			CardNumber string `json:"card_no" binding:"required"`
			CardFlag   int    `json:"card_flag"`
			CardType   int    `json:"card_type"`
			ValidTime  int64  `json:"valid_time"`
			Staff      struct {
				ID int `json:"id"`
			} `json:"staff"`
			AccessGroups []int `json:"access_groups"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	// 删除卡号包含的空格
	req.CardNumber = strings.ReplaceAll(req.CardNumber, " ", "")
	// 删除卡号前缀的零
	req.CardNumber = strings.TrimLeft(req.CardNumber, "0")

	_, err = strconv.ParseInt(req.CardNumber, 10, 64)
	if err != nil {
		ghttp.SendResponseWithError(c, fmt.Errorf("card number %v is not a number", req.CardNumber))
		return
	}

	if err = card.AddCard(c, utils.GetMozuID(c), req.CardNumber, req.CardFlag, req.CardType, req.ValidTime,
		req.Staff.ID, req.AccessGroups); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UpdateCardsFlag 批量更新门禁卡标志位
func UpdateCardsFlag(c *gin.Context) {
	var (
		req struct {
			Cards []string `json:"cards"`
			Flag  int      `json:"flag"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UpdateCardsFlag(c, req.Cards, utils.GetMozuID(c), req.Flag); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UpdateCardsType 批量更新门禁卡类型
func UpdateCardsType(c *gin.Context) {
	var (
		req struct {
			Cards []string `json:"cards"`
			Type  int      `json:"type"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UpdateCardsType(c, req.Cards, utils.GetMozuID(c), req.Type); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UpdateCardValidTime 批量更新门禁卡有效期
func UpdateCardValidTime(c *gin.Context) {
	var (
		req struct {
			Cards     []string `json:"cards"`
			ValidTime int64    `json:"valid_time"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UpdateCardValidTime(c, req.Cards, utils.GetMozuID(c), req.ValidTime); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UnbindCardStaff 解绑门禁卡与人员的关联
func UnbindCardStaff(c *gin.Context) {
	var (
		req struct {
			Card string `json:"card" binding:"required"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UnbindCardStaff(c, req.Card, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UpdateCardStaff 更新门禁卡关联的人员
func UpdateCardStaff(c *gin.Context) {
	var (
		req struct {
			Card  string `json:"card" binding:"required"`
			Staff int    `json:"staff"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UpdateCardStaff(c, req.Card, req.Staff, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// UpdateCardAccess 批量更新门禁卡的权限组
func UpdateCardAccess(c *gin.Context) {
	var (
		req struct {
			Cards  []string `json:"cards"`
			Groups []int    `json:"access_groups"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.UpdateCardsAccessGroups(c, req.Cards, req.Groups, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// DeleteCards 批量删除门禁卡
func DeleteCards(c *gin.Context) {
	var (
		req struct {
			Cards []string `json:"cards"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.DeleteCards(c, req.Cards, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// DeleteCard 删除单张门禁卡
func DeleteCard(c *gin.Context) {
	var (
		req struct {
			Card string `json:"card" binding:"required"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = card.DeleteCards(c, []string{req.Card}, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// ImportCards 从Excel文件导入门禁卡数据
func ImportCards(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var file *multipart.FileHeader = nil
	f := form.File["file"]
	if len(f) > 0 {
		file = f[0]
	}
	if file == nil {
		ghttp.SendResponseWithErrorMessage(c, fmt.Sprintf("file field is empty"))
		return
	}

	if err = card.Import(c, utils.GetMozuID(c), file); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// ExportCards 导出门禁卡数据为Excel文件
func ExportCards(c *gin.Context) {
	f, err := card.Export(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=utf-8''%v",
		url.QueryEscape("门禁卡信息.xlsx")))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write staff excel error: %v", err)
	}
}
