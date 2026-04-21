// Package controller 提供门禁控制器管理的HTTP接口处理器。
package controller

import (
	"fmt"
	"mime/multipart"
	"net/url"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/cgi/controller"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Import 从Excel文件导入控制器配置
func Import(c *gin.Context) {
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

	if err = controller.Import(c, utils.GetMozuID(c), file); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Export 导出控制器配置为Excel文件
func Export(c *gin.Context) {
	f, err := controller.Export(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=utf-8''%v",
		url.QueryEscape("门禁控制器列表.xlsx")))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write controller excel error: %v", err)
	}
}

// BatchDelete 批量删除控制器
func BatchDelete(c *gin.Context) {
	var (
		req struct {
			IDs []int `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = controller.BatchDelete(c, req.IDs); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, "ok")
}

// Delete 删除单个控制器
func Delete(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = controller.BatchDelete(c, []int{req.ID}); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, "ok")
}

// Create 创建新的控制器
func Create(c *gin.Context) {
	var (
		req db.DoorController
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	controllers := make([]rt.DoorController, 1)
	controllers[0].DoorController = req

	if err = controller.BatchCreate(c, utils.GetMozuID(c), controllers, false); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, "ok")
}

// Update 更新控制器配置
func Update(c *gin.Context) {
	var (
		req db.DoorController
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = controller.Update(c, req.ID, req, utils.GetMozuID(c)); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	ghttp.SendResponseWithData(c, "ok")
}

// Get 获取所有控制器列表
func Get(c *gin.Context) {
	controllers, err := controller.GetAll(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, controllers)
}

// Clean 清空控制器的采集数据
func Clean(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = controller.Clean(c, req.ID); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Reset 重置控制器到初始状态
func Reset(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = controller.Reset(c, req.ID, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// AllReset 重置所有控制器到初始状态
func AllReset(c *gin.Context) {
	headers := c.Request.Header
	if err := controller.AllReset(c, utils.GetMozuID(c), headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// ClearTimeGroup 清空控制器的时间组配置
func ClearTimeGroup(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = controller.ClearTimeGroups(c, req.ID, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// AllSyncTime 同步所有控制器的时间
func AllSyncTime(c *gin.Context) {
	var err error
	headers := c.Request.Header
	if err = controller.AllSyncTime(c, utils.GetMozuID(c), headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// SyncTime 同步指定控制器的时间
func SyncTime(c *gin.Context) {
	var (
		req struct {
			ID int `json:"id"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = controller.SyncTime(c, req.ID, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// GetCardFromController 从门控器查询卡是否存在
func GetCardFromController(c *gin.Context) {
	var (
		req struct {
			ControllerID int    `json:"controller_id" binding:"required"`
			CardNo       string `json:"card_no" binding:"required"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	card, exists, err := controller.GetCardFromController(c, req.ControllerID, req.CardNo)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	var resp struct {
		Exists bool        `json:"exists"`
		Card   interface{} `json:"card,omitempty"`
	}
	resp.Exists = exists
	if exists {
		resp.Card = card
	}

	ghttp.SendResponseWithData(c, resp)
}
