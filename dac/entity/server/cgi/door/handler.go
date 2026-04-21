// Package door 提供门管理的HTTP接口处理器。
package door

import (
	"fmt"
	"mime/multipart"
	"net/url"

	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/logic/cgi/door"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// BatchUpdate 批量更新门参数和名称
func BatchUpdate(c *gin.Context) {
	var (
		req struct {
			IDs    []int                `json:"ids"`
			Params driver.DoorParameter `json:"params"`
			Names  map[int]string       `json:"names"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = door.BatchUpdate(c, req.IDs, req.Params, req.Names, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Update 更新单个门的参数、编码和扩展信息
func Update(c *gin.Context) {
	var (
		req struct {
			ID     int `json:"id"`
			Name   string
			Params driver.DoorParameter   `json:"params"`
			Code   string                 `json:"code"`
			Extend map[string]interface{} `json:"extend"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = door.Update(c, req.ID, req.Code, req.Extend, req.Params, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// Get 获取指定ID的门信息
func Get(c *gin.Context) {
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

	d, err := door.Get(c, req.ID)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, d)
}

// UpdateCode 更新门的采集编码
func UpdateCode(c *gin.Context) {
	var (
		req struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	if err = door.UpdateCode(c, req.ID, req.Code); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// SetState 批量设置门的开关状态
func SetState(c *gin.Context) {
	var (
		req struct {
			IDs   []int `json:"ids"`
			State int   `json:"state"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	headers := c.Request.Header
	if err = door.SetState(c, req.IDs, req.State, headers); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}

// ExportCode 导出门列表为Excel文件
func ExportCode(c *gin.Context) {
	f, err := door.Export(c, utils.GetMozuID(c))
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=utf-8''%v", url.QueryEscape("门列表.xlsx")))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write doors excel error: %v", err)
	}
}

// ImportCode 从Excel文件导入门编码配置
func ImportCode(c *gin.Context) {
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

	if err = door.Import(c, file); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, "ok")
}
