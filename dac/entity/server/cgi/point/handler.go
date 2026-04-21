// Package point 提供门禁测点数据的HTTP接口处理器。
package point

import (
	"dac/logic/cgi/point"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Rtd 获取实时测点数据，返回map格式
// 入参id格式为 controller_id.测点名
func Rtd(c *gin.Context) {
	// 入参id是 controller_id.测点名
	var (
		req struct {
			IDs []string `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	points, err := point.GetPoints(c, req.IDs)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, points)
}

// RtdList 获取实时测点数据，返回列表格式
func RtdList(c *gin.Context) {
	var (
		req struct {
			IDs []string `json:"ids"`
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	points, err := point.GetPointsList(c, req.IDs)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	ghttp.SendResponseWithData(c, points)
}
