// Package alarm 提供门禁告警记录的HTTP接口处理器。
package alarm

import (
	"fmt"

	"dac/entity/config"
	"dac/entity/model/cgi"
	"dac/entity/utils"
	"dac/logic/cgi/alarm"

	"dac/entity/utils/ghttp"
	"github.com/gin-gonic/gin"
)

// Export 导出告警记录为Excel文件
func Export(c *gin.Context) {
	var (
		req struct {
			ControllerIDs []int `json:"controller_ids"`
			cgi.TimeCondition
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	f, err := alarm.Export(c, utils.GetMozuID(c), req.ControllerIDs, req.TimeCondition)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", "告警记录.xlsx"))

	if err = f.Write(c.Writer); err != nil {
		config.Log.Warnf("write alarms excel error: %v", err)
	}
}

// Get 查询告警记录，支持分页和时间范围过滤
func Get(c *gin.Context) {
	var (
		req struct {
			ControllerIDs []int `json:"controller_ids"`
			cgi.TimeCondition
			cgi.OffsetCondition
		}
		err error
	)

	if err = c.ShouldBind(&req); err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}

	n, alarms, err := alarm.Get(c, utils.GetMozuID(c),
		req.ControllerIDs, req.Offset, req.Limit,
		req.BeginTime, req.EndTime,
	)
	if err != nil {
		ghttp.SendResponseWithError(c, err)
		return
	}
	var resp struct {
		Total int         `json:"total"`
		List  []cgi.Alarm `json:"list"`
	}
	resp.Total = int(n)
	resp.List = alarms
	ghttp.SendResponseWithData(c, resp)
}
